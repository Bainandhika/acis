package worker_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
	"github.com/Bainandhika/acis/apps/backend/internal/worker"
)

type MockOutboxRepository struct {
	mock.Mock
}

func (m *MockOutboxRepository) EnqueueTx(ctx context.Context, exec repository.DBExecutor, channel, recipient string, payload interface{}) error {
	args := m.Called(ctx, exec, channel, recipient, payload)
	return args.Error(0)
}

func (m *MockOutboxRepository) FetchAndLockPending(ctx context.Context, limit int) ([]domain.NotificationJob, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]domain.NotificationJob), args.Error(0)
}

func (m *MockOutboxRepository) MarkSent(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOutboxRepository) MarkFailed(ctx context.Context, id string, errMsg string) error {
	args := m.Called(ctx, id, errMsg)
	return args.Error(0)
}

func TestWorkerPool_SuccessJob(t *testing.T) {
	mockRepo := new(MockOutboxRepository)
	pool := worker.NewWorkerPool(mockRepo, 2, 10)

	var wg sync.WaitGroup
	wg.Add(1)

	mockRepo.On("MarkSent", mock.Anything, "job-1").Return(nil).Run(func(args mock.Arguments) {
		wg.Done()
	})

	pool.RegisterHandler("email_otp", func(ctx context.Context, job domain.NotificationJob) error {
		return nil
	})

	ctx := context.Background()
	pool.Start(ctx)

	job := domain.NotificationJob{
		ID:        "job-1",
		Channel:   "email_otp",
		Recipient: "test@example.com",
	}

	submitted := pool.Submit(job)
	assert.True(t, submitted)

	wg.Wait()
	pool.Stop()
	mockRepo.AssertExpectations(t)
}

func TestWorkerPool_FailureJob(t *testing.T) {
	mockRepo := new(MockOutboxRepository)
	pool := worker.NewWorkerPool(mockRepo, 2, 10)

	var wg sync.WaitGroup
	wg.Add(1)

	mockRepo.On("MarkFailed", mock.Anything, "job-2", "send failed").Return(nil).Run(func(args mock.Arguments) {
		wg.Done()
	})

	pool.RegisterHandler("email_otp", func(ctx context.Context, job domain.NotificationJob) error {
		return errors.New("send failed")
	})

	ctx := context.Background()
	pool.Start(ctx)

	job := domain.NotificationJob{
		ID:        "job-2",
		Channel:   "email_otp",
		Recipient: "fail@example.com",
	}

	submitted := pool.Submit(job)
	assert.True(t, submitted)

	wg.Wait()
	pool.Stop()
	mockRepo.AssertExpectations(t)
}

func TestWorkerPool_MissingHandler(t *testing.T) {
	mockRepo := new(MockOutboxRepository)
	pool := worker.NewWorkerPool(mockRepo, 2, 10)

	var wg sync.WaitGroup
	wg.Add(1)

	mockRepo.On("MarkFailed", mock.Anything, "job-3", "no handler registered for channel: unknown_channel").Return(nil).Run(func(args mock.Arguments) {
		wg.Done()
	})

	ctx := context.Background()
	pool.Start(ctx)

	job := domain.NotificationJob{
		ID:        "job-3",
		Channel:   "unknown_channel",
		Recipient: "test@example.com",
	}

	submitted := pool.Submit(job)
	assert.True(t, submitted)

	wg.Wait()
	pool.Stop()
	mockRepo.AssertExpectations(t)
}
