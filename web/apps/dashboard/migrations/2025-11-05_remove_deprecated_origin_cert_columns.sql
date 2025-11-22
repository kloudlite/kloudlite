-- Migration: Remove deprecated origin certificate columns from installations table
-- Date: 2025-11-05
-- Description:
--   Remove deprecated origin_certificate, origin_private_key, origin_cert_id,
--   origin_cert_valid_from, and origin_cert_valid_until columns from installations table.
--   These fields are now stored in the tls_certificates table with scope='installation'.

BEGIN;

-- ============================================================================
-- Drop deprecated origin certificate columns from installations table
-- ============================================================================

ALTER TABLE installations
DROP COLUMN IF EXISTS origin_certificate,
DROP COLUMN IF EXISTS origin_private_key,
DROP COLUMN IF EXISTS origin_cert_id,
DROP COLUMN IF EXISTS origin_cert_valid_from,
DROP COLUMN IF EXISTS origin_cert_valid_until;

-- ============================================================================
-- Verify migration
-- ============================================================================

DO $$
BEGIN
    -- Check that deprecated columns no longer exist
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'installations' AND column_name = 'origin_certificate'
    ) THEN
        RAISE EXCEPTION 'Migration failed: origin_certificate column still exists in installations';
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'installations' AND column_name = 'origin_cert_id'
    ) THEN
        RAISE EXCEPTION 'Migration failed: origin_cert_id column still exists in installations';
    END IF;

    -- Verify tls_certificates table exists (our new architecture)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'tls_certificates'
    ) THEN
        RAISE EXCEPTION 'Migration failed: tls_certificates table does not exist';
    END IF;

    RAISE NOTICE 'Migration completed successfully! Deprecated origin certificate columns removed from installations table.';
END $$;

COMMIT;
