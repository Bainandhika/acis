-- Migration: 002_add_otp_table.sql
-- Description: Add table for storing passwordless OTP codes
-- Date: 2026-06-07

CREATE TABLE otp_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    code_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index to quickly find active OTPs for an email
CREATE INDEX idx_otp_codes_email ON otp_codes(email);
CREATE INDEX idx_otp_codes_expires_at ON otp_codes(expires_at);