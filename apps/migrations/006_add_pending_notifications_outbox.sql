-- Migration: 006_add_pending_notifications_outbox.sql
CREATE TABLE IF NOT EXISTS pending_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel VARCHAR(50) NOT NULL,        -- 'email_otp', 'telegram_alert', 'telegram_msg'
    recipient VARCHAR(255) NOT NULL,    -- Email address or Telegram chat_id
    payload JSONB NOT NULL,             -- Message content or template variables
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'processing', 'sent', 'failed', 'dead'
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    last_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pending_notifications_status ON pending_notifications(status, created_at);
