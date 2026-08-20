package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/redis/go-redis/v9"
)

const OutboxNotifyChannel = "outbox:notify"

type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// NotificationJob represents a pending or processed outbox notification job
type NotificationJob struct {
	ID         string          `db:"id" json:"id"`
	Channel    string          `db:"channel" json:"channel"`
	Recipient  string          `db:"recipient" json:"recipient"`
	Payload    json.RawMessage `db:"payload" json:"payload"`
	Status     string          `db:"status" json:"status"`
	RetryCount int             `db:"retry_count" json:"retry_count"`
	MaxRetries int             `db:"max_retries" json:"max_retries"`
	LastError  *string         `db:"last_error" json:"last_error"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at" json:"updated_at"`
}

type OutboxRepository interface {
	EnqueueTx(ctx context.Context, exec DBExecutor, channel, recipient string, payload interface{}) error
	FetchAndLockPending(ctx context.Context, limit int) ([]NotificationJob, error)
	MarkSent(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, errMsg string) error
	PublishSignal(ctx context.Context) error
}

type outboxRepository struct {
	db          *database.AppDB
	redisClient *redis.Client
}

func NewOutboxRepository(db *database.AppDB, redisClient ...*redis.Client) OutboxRepository {
	var rClient *redis.Client
	if len(redisClient) > 0 {
		rClient = redisClient[0]
	}
	return &outboxRepository{
		db:          db,
		redisClient: rClient,
	}
}

func (r *outboxRepository) PublishSignal(ctx context.Context) error {
	if r.redisClient == nil {
		return nil
	}
	if err := r.redisClient.Publish(ctx, OutboxNotifyChannel, "1").Err(); err != nil {
		slog.Warn("Failed to publish outbox notification signal to Redis", slog.Any("error", err))
		return err
	}
	return nil
}

func (r *outboxRepository) EnqueueTx(ctx context.Context, exec DBExecutor, channel, recipient string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal outbox payload: %w", err)
	}

	query := `
		INSERT INTO pending_notifications (channel, recipient, payload, status)
		VALUES ($1, $2, $3, 'pending')
	`
	_, err = exec.ExecContext(ctx, query, channel, recipient, string(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to enqueue notification: %w", err)
	}

	return nil
}

func (r *outboxRepository) FetchAndLockPending(ctx context.Context, limit int) ([]NotificationJob, error) {
	query := `
		UPDATE pending_notifications
		SET status = 'processing', updated_at = NOW()
		WHERE id IN (
			SELECT id FROM pending_notifications
			WHERE status = 'pending' 
			   OR (status = 'failed' AND retry_count < max_retries AND updated_at < NOW() - INTERVAL '30 seconds')
			ORDER BY created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, channel, recipient, payload, status, retry_count, max_retries, last_error, created_at, updated_at
	`

	var jobs []NotificationJob
	err := r.db.AdminDB().SelectContext(ctx, &jobs, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch and lock pending notifications: %w", err)
	}

	return jobs, nil
}

func (r *outboxRepository) MarkSent(ctx context.Context, id string) error {
	query := `
		UPDATE pending_notifications
		SET status = 'sent', updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.AdminDB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as sent: %w", err)
	}
	return nil
}

func (r *outboxRepository) MarkFailed(ctx context.Context, id string, errMsg string) error {
	query := `
		UPDATE pending_notifications
		SET retry_count = retry_count + 1,
		    last_error = $2,
		    status = CASE WHEN retry_count + 1 >= max_retries THEN 'dead' ELSE 'failed' END,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.AdminDB().ExecContext(ctx, query, id, errMsg)
	if err != nil {
		return fmt.Errorf("failed to mark notification as failed: %w", err)
	}
	return nil
}
