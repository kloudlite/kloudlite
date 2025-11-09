-- Migration: Add poller_active column to installations table
-- This tracks when the SubdomainPoller starts actively polling for configuration

ALTER TABLE installations
ADD COLUMN IF NOT EXISTS poller_active BOOLEAN DEFAULT FALSE;

-- Update existing installations: if they have a secret_key and last_health_check,
-- they're already polling, so set poller_active to true
UPDATE installations
SET poller_active = TRUE
WHERE secret_key IS NOT NULL
  AND last_health_check IS NOT NULL;
