-- Rollback Migration: Restore deprecated origin certificate columns to installations table
-- Date: 2025-11-05
-- Description:
--   Rollback migration to restore origin_certificate, origin_private_key, origin_cert_id,
--   origin_cert_valid_from, and origin_cert_valid_until columns to installations table.
--   Use this ONLY if you need to rollback to the previous code version.

BEGIN;

-- ============================================================================
-- Restore deprecated origin certificate columns to installations table
-- ============================================================================

ALTER TABLE installations
ADD COLUMN IF NOT EXISTS origin_certificate TEXT,
ADD COLUMN IF NOT EXISTS origin_private_key TEXT,
ADD COLUMN IF NOT EXISTS origin_cert_id TEXT,
ADD COLUMN IF NOT EXISTS origin_cert_valid_from TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS origin_cert_valid_until TIMESTAMPTZ;

COMMENT ON COLUMN installations.origin_certificate IS 'DEPRECATED: Cloudflare Origin Certificate PEM - Use tls_certificates table instead';
COMMENT ON COLUMN installations.origin_private_key IS 'DEPRECATED: Private key for the origin certificate - Use tls_certificates table instead';
COMMENT ON COLUMN installations.origin_cert_id IS 'DEPRECATED: Cloudflare certificate ID - Use tls_certificates table instead';
COMMENT ON COLUMN installations.origin_cert_valid_from IS 'DEPRECATED: Certificate validity start date - Use tls_certificates table instead';
COMMENT ON COLUMN installations.origin_cert_valid_until IS 'DEPRECATED: Certificate validity end date - Use tls_certificates table instead';

-- ============================================================================
-- Verify rollback
-- ============================================================================

DO $$
BEGIN
    -- Check that columns were restored
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'installations' AND column_name = 'origin_certificate'
    ) THEN
        RAISE EXCEPTION 'Rollback failed: origin_certificate column not restored to installations';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'installations' AND column_name = 'origin_cert_id'
    ) THEN
        RAISE EXCEPTION 'Rollback failed: origin_cert_id column not restored to installations';
    END IF;

    RAISE NOTICE 'Rollback completed successfully! Origin certificate columns restored to installations table.';
    RAISE WARNING 'Note: These columns are deprecated. The tls_certificates table should be used for certificate storage.';
END $$;

COMMIT;
