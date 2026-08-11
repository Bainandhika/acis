-- Migration: 005_performance_indexes.sql
-- Description: Add performance indexes for compound queries and rate limiting lookups
-- Date: 2026-08-11

CREATE INDEX IF NOT EXISTS idx_proposals_wallet_status ON proposals(wallet_id, status);
CREATE INDEX IF NOT EXISTS idx_family_members_user_family ON family_members(user_id, family_id);
CREATE INDEX IF NOT EXISTS idx_otp_codes_email_active ON otp_codes(email, is_used, expires_at);
