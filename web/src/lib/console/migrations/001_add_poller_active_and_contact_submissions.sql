-- ============================================================================
-- Migration: Add poller_active field and contact_submissions table
-- Date: 2025-11-10
-- ============================================================================

-- Add poller_active field to installations table
ALTER TABLE installations
ADD COLUMN IF NOT EXISTS poller_active BOOLEAN DEFAULT FALSE;

-- Add index for poller_active if needed for queries
CREATE INDEX IF NOT EXISTS idx_installations_poller_active ON installations(poller_active);

-- ============================================================================
-- Create contact_submissions table
-- ============================================================================

CREATE TABLE IF NOT EXISTS contact_submissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  subject TEXT NOT NULL,
  message TEXT NOT NULL,
  submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for contact_submissions
CREATE INDEX IF NOT EXISTS idx_contact_submissions_email ON contact_submissions(email);
CREATE INDEX IF NOT EXISTS idx_contact_submissions_submitted_at ON contact_submissions(submitted_at DESC);

-- Add trigger for updated_at timestamp on contact_submissions
-- Drop trigger if it exists, then recreate
DROP TRIGGER IF EXISTS update_contact_submissions_updated_at ON contact_submissions;
CREATE TRIGGER update_contact_submissions_updated_at
    BEFORE UPDATE ON contact_submissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Migration Complete
-- ============================================================================

-- Verify changes
DO $$
BEGIN
    -- Check if poller_active column was added
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'installations'
        AND column_name = 'poller_active'
    ) THEN
        RAISE NOTICE 'Successfully added poller_active column to installations table';
    END IF;

    -- Check if contact_submissions table was created
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'contact_submissions'
    ) THEN
        RAISE NOTICE 'Successfully created contact_submissions table';
    END IF;
END $$;
