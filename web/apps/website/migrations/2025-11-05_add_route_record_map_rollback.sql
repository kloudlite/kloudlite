-- Rollback Migration: Remove route_record_map column from ip_records table
-- Date: 2025-11-05
-- Description:
--   Rollback the route_record_map column addition

BEGIN;

ALTER TABLE ip_records
DROP COLUMN IF EXISTS route_record_map;

-- Verify rollback
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'ip_records' AND column_name = 'route_record_map'
    ) THEN
        RAISE EXCEPTION 'Rollback failed: route_record_map column still exists in ip_records';
    END IF;

    RAISE NOTICE 'Rollback completed successfully: route_record_map column removed';
END $$;

COMMIT;
