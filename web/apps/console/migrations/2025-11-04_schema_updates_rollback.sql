-- Rollback Migration: Revert schema updates for unified DNS architecture
-- Date: 2025-11-04
-- WARNING: This will revert to the old schema and may result in data loss!

BEGIN;

-- ============================================================================
-- 1. Remove origin certificate fields from installations table
-- ============================================================================

ALTER TABLE installations
DROP COLUMN IF EXISTS origin_certificate,
DROP COLUMN IF EXISTS origin_private_key,
DROP COLUMN IF EXISTS origin_cert_id,
DROP COLUMN IF EXISTS origin_cert_valid_from,
DROP COLUMN IF EXISTS origin_cert_valid_until;

-- ============================================================================
-- 2. Revert ip_records table schema
-- ============================================================================

-- Create old ip_records table structure
CREATE TABLE IF NOT EXISTS ip_records_old (
    id SERIAL PRIMARY KEY,
    installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('installation', 'workmachine')),
    ip TEXT NOT NULL,
    work_machine_name TEXT,
    configured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dns_record_ids TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(installation_id, type, work_machine_name)
);

-- Migrate data back (with data loss warning)
INSERT INTO ip_records_old (
    installation_id,
    type,
    ip,
    work_machine_name,
    configured_at,
    dns_record_ids,
    created_at,
    updated_at
)
SELECT
    installation_id,
    'installation' as type,
    ip,
    domain_request_name as work_machine_name,
    configured_at,
    ARRAY[ssh_record_id] || route_record_ids as dns_record_ids,
    created_at,
    updated_at
FROM ip_records
ON CONFLICT (installation_id, type, work_machine_name) DO NOTHING;

-- Drop new table and rename old one back
DROP TABLE IF EXISTS ip_records CASCADE;
ALTER TABLE ip_records_old RENAME TO ip_records;

-- ============================================================================
-- 3. Revert edge_certificates table schema
-- ============================================================================

-- Create old edge_certificates table structure
CREATE TABLE IF NOT EXISTS edge_certificates_old (
    id SERIAL PRIMARY KEY,
    installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
    cloudflare_cert_pack_id TEXT NOT NULL,
    hostnames TEXT[] NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN ('installation', 'workmachine')),
    scope_identifier TEXT,
    ordered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(installation_id, scope, scope_identifier)
);

-- Migrate data back
INSERT INTO edge_certificates_old (
    installation_id,
    cloudflare_cert_pack_id,
    hostnames,
    scope,
    scope_identifier,
    ordered_at,
    status,
    created_at,
    updated_at
)
SELECT
    installation_id,
    cloudflare_cert_pack_id,
    hostnames,
    'domainrequest' as scope,
    domain_request_name as scope_identifier,
    ordered_at,
    status,
    created_at,
    updated_at
FROM edge_certificates
ON CONFLICT (installation_id, scope, scope_identifier) DO NOTHING;

-- Drop new table and rename old one back
DROP TABLE IF EXISTS edge_certificates CASCADE;
ALTER TABLE edge_certificates_old RENAME TO edge_certificates;

-- ============================================================================
-- Recreate triggers
-- ============================================================================

DROP TRIGGER IF EXISTS update_ip_records_updated_at ON ip_records;
CREATE TRIGGER update_ip_records_updated_at
    BEFORE UPDATE ON ip_records
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_edge_certificates_updated_at ON edge_certificates;
CREATE TRIGGER update_edge_certificates_updated_at
    BEFORE UPDATE ON edge_certificates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

RAISE NOTICE 'Rollback completed!';

COMMIT;
