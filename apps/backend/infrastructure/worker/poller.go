package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
)

type OutboxPoller struct {
	repo      notification.OutboxRepository
	pool      *WorkerPool
	interval  time.Duration
	batchSize int
	stopChan  chan struct{}
	ticker    *time.Ticker
}

func NewOutboxPoller(repo notification.OutboxRepository, pool *WorkerPool, interval time.Duration, batchSize int) *OutboxPoller {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if batchSize <= 0 {
		batchSize = 20
	}
	return &OutboxPoller{
		repo:      repo,
		pool:      pool,
		interval:  interval,
		batchSize: batchSize,
		stopChan:  make(chan struct{}),
	}
}

func (p *OutboxPoller) Start(ctx context.Context) {
	p.ticker = time.NewTicker(p.interval)
	slog.Info("Outbox poller started", slog.Duration("interval", p.interval), slog.Int("batch_size", p.batchSize))

	go func() {
		for {
			select {
			case <-p.stopChan:
				return
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
	close(p.stopChan)
	slog.Info("Outbox poller stopped")
}
