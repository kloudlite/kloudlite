-- Migration: Add renewal_jobs and cron_job_logs tables
-- Created: 2026-03-02
-- Description: Scheduled renewal/expiry jobs per installation + audit log

-- ============================================================================
-- 1. RENEWAL JOBS TABLE
-- Explicitly scheduled jobs: created on subscribe, cancelled on unsubscribe
-- ============================================================================

CREATE TABLE IF NOT EXISTS renewal_jobs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  job_type TEXT NOT NULL CHECK (job_type IN ('renewal', 'expire')),
  scheduled_at TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
  attempts INTEGER NOT NULL DEFAULT 0,
  last_error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_renewal_jobs_pending ON renewal_jobs(status, scheduled_at)
  WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_renewal_jobs_installation ON renewal_jobs(installation_id, status);

DROP TRIGGER IF EXISTS update_renewal_jobs_updated_at ON renewal_jobs;
CREATE TRIGGER update_renewal_jobs_updated_at
    BEFORE UPDATE ON renewal_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 2. CRON JOB LOGS TABLE
-- Audit trail for every job execution (success or failure)
-- ============================================================================

CREATE TABLE IF NOT EXISTS cron_job_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  job_id UUID REFERENCES renewal_jobs(id) ON DELETE SET NULL,
  job_type TEXT NOT NULL,
  installation_id UUID REFERENCES installations(id) ON DELETE SET NULL,
  status TEXT NOT NULL CHECK (status IN ('success', 'failed')),
  details TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cron_job_logs_status ON cron_job_logs(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_cron_job_logs_installation ON cron_job_logs(installation_id);

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
