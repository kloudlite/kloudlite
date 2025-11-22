-- Migration: Add route_record_map column to ip_records table
-- Date: 2025-11-05
-- Description:
--   Add route_record_map JSONB column to enable domain-to-record-ID mapping
--   for differential DNS updates

BEGIN;

ALTER TABLE ip_records
ADD COLUMN IF NOT EXISTS route_record_map JSONB DEFAULT '{}'::jsonb;

COMMENT ON COLUMN ip_records.route_record_map IS 'Mapping of domain names to Cloudflare DNS record IDs for efficient differential updates';

-- Verify migration
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'ip_records' AND column_name = 'route_record_map'
    ) THEN
        RAISE EXCEPTION 'Migration failed: route_record_map column not added to ip_records';
    END IF;

    RAISE NOTICE 'Migration completed successfully: route_record_map column added';
END $$;

COMMIT;
