-- Migration: Add processed_webhook_events table for webhook idempotency
-- Created: 2026-03-02
-- Description: Track processed Razorpay webhook events to prevent duplicate processing

-- ============================================================================
-- 1. PROCESSED WEBHOOK EVENTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS processed_webhook_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  razorpay_event_id TEXT NOT NULL UNIQUE,
  event_type TEXT NOT NULL,
  processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhook_events_processed_at
  ON processed_webhook_events(processed_at);

-- ============================================================================
-- 2. RLS POLICY
-- No direct user access — only the service role (server actions / webhooks)
-- ============================================================================

ALTER TABLE processed_webhook_events ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- 3. AUTO-CLEANUP: Remove events older than 30 days
-- This keeps the table small while retaining enough history for deduplication
-- ============================================================================

CREATE OR REPLACE FUNCTION cleanup_old_webhook_events()
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
  DELETE FROM processed_webhook_events
  WHERE processed_at < NOW() - INTERVAL '30 days';
END;
$$;

-- TODO: Enable daily cleanup via pg_cron (requires pg_cron extension from migration 007)
-- Run this manually after applying this migration:
-- SELECT cron.schedule('cleanup-webhook-events', '0 3 * * *', 'SELECT cleanup_old_webhook_events()');

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
