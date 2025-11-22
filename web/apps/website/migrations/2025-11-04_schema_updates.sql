-- Migration: Update schema for unified DNS architecture
-- Date: 2025-11-04
-- Description:
--   1. Add origin certificate fields to installations table
--   2. Migrate ip_records table from type-based to domain-based schema
--   3. Simplify edge_certificates table by removing scope concept

BEGIN;

-- ============================================================================
-- 1. Add origin certificate fields to installations table
-- ============================================================================

ALTER TABLE installations
ADD COLUMN IF NOT EXISTS origin_certificate TEXT,
ADD COLUMN IF NOT EXISTS origin_private_key TEXT,
ADD COLUMN IF NOT EXISTS origin_cert_id TEXT,
ADD COLUMN IF NOT EXISTS origin_cert_valid_from TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS origin_cert_valid_until TIMESTAMPTZ;

COMMENT ON COLUMN installations.origin_certificate IS 'Cloudflare Origin Certificate PEM';
COMMENT ON COLUMN installations.origin_private_key IS 'Private key for the origin certificate';
COMMENT ON COLUMN installations.origin_cert_id IS 'Cloudflare certificate ID';
COMMENT ON COLUMN installations.origin_cert_valid_from IS 'Certificate validity start date';
COMMENT ON COLUMN installations.origin_cert_valid_until IS 'Certificate validity end date';

-- ============================================================================
-- 2. Migrate ip_records table schema
-- ============================================================================

-- Create new ip_records table with updated schema
CREATE TABLE IF NOT EXISTS ip_records_new (
    id SERIAL PRIMARY KEY,
    installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
    domain_request_name TEXT NOT NULL,
    ip TEXT NOT NULL,
    configured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ssh_record_id TEXT,
    route_record_ids TEXT[] DEFAULT '{}',
    domain_routes JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(installation_id, domain_request_name)
);

-- Migrate existing data (if any exists)
-- Note: This assumes old records with type='installation' should be migrated
-- Adjust the logic if you have specific requirements for workmachine records
INSERT INTO ip_records_new (
    installation_id,
    domain_request_name,
    ip,
    configured_at,
    ssh_record_id,
    route_record_ids,
    domain_routes,
    created_at,
    updated_at
)
SELECT
    installation_id,
    COALESCE(work_machine_name, 'default-domain-request') as domain_request_name,
    ip,
    configured_at,
    CASE WHEN ARRAY_LENGTH(dns_record_ids, 1) > 0 THEN dns_record_ids[1] ELSE NULL END as ssh_record_id,
    CASE WHEN ARRAY_LENGTH(dns_record_ids, 1) > 1 THEN dns_record_ids[2:] ELSE '{}' END as route_record_ids,
    '[]'::jsonb as domain_routes,
    created_at,
    updated_at
FROM ip_records
WHERE type = 'installation'
ON CONFLICT (installation_id, domain_request_name) DO NOTHING;

-- Drop old table and rename new one
DROP TABLE IF EXISTS ip_records CASCADE;
ALTER TABLE ip_records_new RENAME TO ip_records;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_ip_records_installation_id ON ip_records(installation_id);
CREATE INDEX IF NOT EXISTS idx_ip_records_domain_request_name ON ip_records(domain_request_name);

-- ============================================================================
-- 3. Simplify edge_certificates table
-- ============================================================================

-- Create new edge_certificates table with updated schema
CREATE TABLE IF NOT EXISTS edge_certificates_new (
    id SERIAL PRIMARY KEY,
    installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
    cloudflare_cert_pack_id TEXT NOT NULL,
    hostnames TEXT[] NOT NULL,
    domain_request_name TEXT NOT NULL,
    ordered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(installation_id, domain_request_name)
);

-- Migrate existing data
-- Only migrate records with scope='domainrequest' as that's what we're using now
INSERT INTO edge_certificates_new (
    installation_id,
    cloudflare_cert_pack_id,
    hostnames,
    domain_request_name,
    ordered_at,
    status,
    created_at,
    updated_at
)
SELECT
    installation_id,
    cloudflare_cert_pack_id,
    hostnames,
    COALESCE(scope_identifier, 'default-domain-request') as domain_request_name,
    ordered_at,
    status,
    created_at,
    updated_at
FROM edge_certificates
WHERE scope = 'domainrequest'
ON CONFLICT (installation_id, domain_request_name) DO NOTHING;

-- Drop old table and rename new one
DROP TABLE IF EXISTS edge_certificates CASCADE;
ALTER TABLE edge_certificates_new RENAME TO edge_certificates;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_edge_certificates_installation_id ON edge_certificates(installation_id);
CREATE INDEX IF NOT EXISTS idx_edge_certificates_domain_request_name ON edge_certificates(domain_request_name);
CREATE INDEX IF NOT EXISTS idx_edge_certificates_status ON edge_certificates(status);

-- ============================================================================
-- Create updated_at trigger functions (if not exists)
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers for updated_at
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

-- ============================================================================
-- Verify migration
-- ============================================================================

-- Check that new columns exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'installations' AND column_name = 'origin_certificate'
    ) THEN
        RAISE EXCEPTION 'Migration failed: origin_certificate column not added to installations';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'ip_records' AND column_name = 'domain_request_name'
    ) THEN
        RAISE EXCEPTION 'Migration failed: domain_request_name column not added to ip_records';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'edge_certificates' AND column_name = 'domain_request_name'
    ) THEN
        RAISE EXCEPTION 'Migration failed: domain_request_name column not added to edge_certificates';
    END IF;

    RAISE NOTICE 'Migration completed successfully!';
END $$;

COMMIT;
