package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	"github.com/redis/go-redis/v9"
)

type OutboxPoller struct {
	repo        notification.OutboxRepository
	pool        *WorkerPool
	redisClient *redis.Client
	interval    time.Duration
	batchSize   int
	stopChan    chan struct{}
	ticker      *time.Ticker
	pubsub      *redis.PubSub
}

func NewOutboxPoller(repo notification.OutboxRepository, pool *WorkerPool, interval time.Duration, batchSize int, redisClient ...*redis.Client) *OutboxPoller {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	if batchSize <= 0 {
		batchSize = 20
	}
	var rClient *redis.Client
	if len(redisClient) > 0 {
		rClient = redisClient[0]
	}
	return &OutboxPoller{
		repo:        repo,
		pool:        pool,
		redisClient: rClient,
		interval:    interval,
		batchSize:   batchSize,
		stopChan:    make(chan struct{}),
	}
}

func (p *OutboxPoller) Start(ctx context.Context) {
	p.ticker = time.NewTicker(p.interval)
	slog.Info("Outbox poller started", slog.Duration("fallback_interval", p.interval), slog.Int("batch_size", p.batchSize))

	// Initial poll on startup to pick up any left-over pending notifications
	p.poll(ctx)

	var pubsubCh <-chan *redis.Message
	if p.redisClient != nil {
		p.pubsub = p.redisClient.Subscribe(ctx, notification.OutboxNotifyChannel)
		pubsubCh = p.pubsub.Channel()
		slog.Info("Outbox poller subscribed to Redis event channel", slog.String("channel", notification.OutboxNotifyChannel))
	}

	go func() {
		for {
			select {
			case <-p.stopChan:
				return
			case msg, ok := <-pubsubCh:
				if !ok {
					continue
				}
				slog.Debug("Received outbox notification signal from Redis", slog.String("payload", msg.Payload))
				p.poll(ctx)
			case <-p.ticker.C:
				p.poll(ctx)
			}
		}
	}()
}

func (p *OutboxPoller) poll(ctx context.Context) {
	jobs, err := p.repo.FetchAndLockPending(ctx, p.batchSize)
	if err != nil {
		slog.Error("Outbox poller failed to fetch pending notifications", slog.Any("error", err))
		return
	}

	if len(jobs) == 0 {
		return
	}

	slog.Debug("Fetched pending outbox jobs", slog.Int("job_count", len(jobs)))

	for _, job := range jobs {
		submitted := p.pool.Submit(job)
		if !submitted {
			slog.Warn("Worker pool channel full, job will be retried next tick", slog.String("job_id", job.ID))
		}
	}
}

func (p *OutboxPoller) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
	if p.pubsub != nil {
		_ = p.pubsub.Close()
	}
	close(p.stopChan)
	slog.Info("Outbox poller stopped")
}
