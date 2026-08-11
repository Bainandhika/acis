package worker

import (
	"context"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	zerolog "github.com/rs/zerolog/log"
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
	zerolog.Info().Dur("interval", p.interval).Int("batch_size", p.batchSize).Msg("Outbox poller started")

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
		zerolog.Error().Err(err).Msg("Outbox poller failed to fetch pending notifications")
		return
	}

	if len(jobs) == 0 {
		return
	}

	zerolog.Debug().Int("job_count", len(jobs)).Msg("Fetched pending outbox jobs")

	for _, job := range jobs {
		submitted := p.pool.Submit(job)
		if !submitted {
			zerolog.Warn().Str("job_id", job.ID).Msg("Worker pool channel full, job will be retried next tick")
		}
	}
}

func (p *OutboxPoller) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
	close(p.stopChan)
	zerolog.Info().Msg("Outbox poller stopped")
}
