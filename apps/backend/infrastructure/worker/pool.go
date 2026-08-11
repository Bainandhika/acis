package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	zerolog "github.com/rs/zerolog/log"
)

type NotificationHandler func(ctx context.Context, job notification.NotificationJob) error

type WorkerPool struct {
	jobChan     chan notification.NotificationJob
	workerCount int
	handlers    map[string]NotificationHandler
	repo        notification.OutboxRepository
	wg          sync.WaitGroup
	stopChan    chan struct{}
	mu          sync.RWMutex
}

func NewWorkerPool(repo notification.OutboxRepository, workerCount int, chanBufferSize int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = 3
	}
	if chanBufferSize <= 0 {
		chanBufferSize = 100
	}
	return &WorkerPool{
		jobChan:     make(chan notification.NotificationJob, chanBufferSize),
		workerCount: workerCount,
		handlers:    make(map[string]NotificationHandler),
		repo:        repo,
		stopChan:    make(chan struct{}),
	}
}

func (p *WorkerPool) RegisterHandler(channel string, handler NotificationHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[channel] = handler
}

func (p *WorkerPool) Submit(job notification.NotificationJob) bool {
	select {
	case p.jobChan <- job:
		return true
	default:
		return false
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 1; i <= p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
	zerolog.Info().Int("workers", p.workerCount).Msg("Outbox worker pool started")
}

func (p *WorkerPool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	for {
		select {
		case <-p.stopChan:
			return
		case job, ok := <-p.jobChan:
			if !ok {
				return
			}
			p.processJob(ctx, job, id)
		}
	}
}

func (p *WorkerPool) processJob(ctx context.Context, job notification.NotificationJob, workerID int) {
	p.mu.RLock()
	handler, exists := p.handlers[job.Channel]
	p.mu.RUnlock()

	if !exists {
		errMsg := fmt.Sprintf("no handler registered for channel: %s", job.Channel)
		zerolog.Error().Str("job_id", job.ID).Str("channel", job.Channel).Msg(errMsg)
		_ = p.repo.MarkFailed(ctx, job.ID, errMsg)
		return
	}

	zerolog.Debug().Str("job_id", job.ID).Int("worker_id", workerID).Msg("Processing outbox notification job")

	if err := handler(ctx, job); err != nil {
		zerolog.Error().Err(err).Str("job_id", job.ID).Int("worker_id", workerID).Msg("Failed to execute notification job")
		_ = p.repo.MarkFailed(ctx, job.ID, err.Error())
	} else {
		zerolog.Info().Str("job_id", job.ID).Int("worker_id", workerID).Msg("Notification job executed successfully")
		_ = p.repo.MarkSent(ctx, job.ID)
	}
}

func (p *WorkerPool) Stop() {
	close(p.stopChan)
	close(p.jobChan)
	p.wg.Wait()
	zerolog.Info().Msg("Outbox worker pool stopped gracefully")
}
