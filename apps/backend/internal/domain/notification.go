package domain

import (
	"encoding/json"
	"time"
)

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
