-- Migration: Add edge_certificates table and remove edge_certificate_pack_id from installations
-- Created: 2025-11-03
--
-- This migration creates a new edge_certificates table to support hierarchical certificate management.
-- Edge certificates are needed at multiple scopes (installation and workmachine) to handle the 50 hostname
-- limit per CloudFlare certificate pack. The old single edge_certificate_pack_id column in installations
-- table cannot scale beyond this limit.

-- Create edge_certificates table
CREATE TABLE IF NOT EXISTS edge_certificates (
  id SERIAL PRIMARY KEY,
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  cloudflare_cert_pack_id TEXT NOT NULL,
  hostnames TEXT[] NOT NULL,
  scope TEXT NOT NULL CHECK (scope IN ('installation', 'workmachine')),
  scope_identifier TEXT,
  ordered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'failed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add indexes for common queries
CREATE INDEX IF NOT EXISTS idx_edge_certificates_installation_id
  ON edge_certificates(installation_id);

CREATE INDEX IF NOT EXISTS idx_edge_certificates_scope
  ON edge_certificates(installation_id, scope, scope_identifier);

CREATE INDEX IF NOT EXISTS idx_edge_certificates_cloudflare_cert_pack_id
  ON edge_certificates(cloudflare_cert_pack_id);

CREATE INDEX IF NOT EXISTS idx_edge_certificates_status
  ON edge_certificates(status);

-- Add unique constraint to prevent duplicate certificates for the same scope
CREATE UNIQUE INDEX IF NOT EXISTS idx_edge_certificates_unique_scope
  ON edge_certificates(installation_id, scope, COALESCE(scope_identifier, ''));

-- Add comments to explain the table and columns
COMMENT ON TABLE edge_certificates IS
'CloudFlare Edge Certificate Packs for browser-to-CloudFlare TLS encryption. Supports hierarchical certificates at installation and workmachine scopes.';

COMMENT ON COLUMN edge_certificates.installation_id IS
'Reference to the installation that owns this certificate';

COMMENT ON COLUMN edge_certificates.cloudflare_cert_pack_id IS
'CloudFlare Advanced Certificate Manager certificate pack ID';

COMMENT ON COLUMN edge_certificates.hostnames IS
'List of hostnames covered by this certificate pack (max 50 per CloudFlare limit)';

COMMENT ON COLUMN edge_certificates.scope IS
'Certificate scope: installation (for {subdomain}.khost.dev and *.{subdomain}.khost.dev) or workmachine (for {wm}.{subdomain}.khost.dev and *.{wm}.{subdomain}.khost.dev)';

COMMENT ON COLUMN edge_certificates.scope_identifier IS
'Identifier for the scope: NULL for installation scope, workmachine name for workmachine scope';

COMMENT ON COLUMN edge_certificates.ordered_at IS
'Timestamp when the certificate was ordered from CloudFlare';

COMMENT ON COLUMN edge_certificates.status IS
'Certificate status: pending (being provisioned), active (ready to use), or failed (provisioning failed)';

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_edge_certificates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER edge_certificates_updated_at
  BEFORE UPDATE ON edge_certificates
  FOR EACH ROW
  EXECUTE FUNCTION update_edge_certificates_updated_at();

-- Remove edge_certificate_pack_id column from installations table
-- This column is being replaced by the new edge_certificates table
ALTER TABLE installations
DROP COLUMN IF EXISTS edge_certificate_pack_id;
