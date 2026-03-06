-- Migration: Harden billing and cron security
-- Created: 2026-03-06
-- Description:
--   1) Enforce non-negative subscription quantities
--   2) Enable RLS + service-role-only policies on renewal_jobs and cron_job_logs

-- ============================================================================
-- 1. DATA SANITY + QUANTITY CONSTRAINT
-- ============================================================================

-- Normalize any unexpected historical negatives before adding the constraint
UPDATE subscriptions
SET quantity = 0
WHERE quantity < 0;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'subscriptions_quantity_non_negative'
      AND conrelid = 'subscriptions'::regclass
  ) THEN
    ALTER TABLE subscriptions
      ADD CONSTRAINT subscriptions_quantity_non_negative
      CHECK (quantity >= 0);
  END IF;
END $$;

-- ============================================================================
-- 2. ENABLE RLS ON CRON TABLES
-- ============================================================================

ALTER TABLE renewal_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE cron_job_logs ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS renewal_jobs_service_role_only ON renewal_jobs;
CREATE POLICY renewal_jobs_service_role_only
  ON renewal_jobs
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);

DROP POLICY IF EXISTS cron_job_logs_service_role_only ON cron_job_logs;
CREATE POLICY cron_job_logs_service_role_only
  ON cron_job_logs
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);
