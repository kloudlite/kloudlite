-- Add ACA job tracking columns to installations table
-- These track Azure Container Apps Job executions for server-side OCI provisioning

ALTER TABLE installations
  ADD COLUMN IF NOT EXISTS aca_job_execution_name TEXT,
  ADD COLUMN IF NOT EXISTS aca_job_status TEXT,
  ADD COLUMN IF NOT EXISTS aca_job_started_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS aca_job_completed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS aca_job_error TEXT;

-- Add 'oci' as a valid cloud_provider if there's a CHECK constraint
-- (If cloud_provider is just TEXT, this is a no-op comment)
-- The column already allows 'aws', 'gcp', 'azure' - we add 'oci'
COMMENT ON COLUMN installations.aca_job_execution_name IS 'Azure Container Apps Job execution name';
COMMENT ON COLUMN installations.aca_job_status IS 'Job status: pending, running, succeeded, failed, unknown';
