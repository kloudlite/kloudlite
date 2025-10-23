-- Kloudlite Registration System Database Schema
-- PostgreSQL with ACID transactions for atomic operations

-- User registrations table
CREATE TABLE IF NOT EXISTS user_registrations (
  email TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  providers TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  installation_key TEXT NOT NULL UNIQUE,
  secret_key TEXT,
  has_completed_installation BOOLEAN NOT NULL DEFAULT FALSE,
  subdomain TEXT,
  reserved_at TIMESTAMPTZ,
  deployment_ready BOOLEAN DEFAULT FALSE,
  last_health_check TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- IP records table (related to user registrations)
CREATE TABLE IF NOT EXISTS ip_records (
  id SERIAL PRIMARY KEY,
  user_email TEXT NOT NULL REFERENCES user_registrations(email) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('installation', 'workmachine')),
  ip TEXT NOT NULL,
  work_machine_name TEXT,
  configured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  dns_record_ids JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_email, type, work_machine_name)
);

-- Domain reservations table
CREATE TABLE IF NOT EXISTS domain_reservations (
  subdomain TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  user_email TEXT NOT NULL REFERENCES user_registrations(email) ON DELETE CASCADE,
  user_name TEXT NOT NULL,
  reserved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  status TEXT NOT NULL CHECK (status IN ('reserved', 'active', 'cancelled')) DEFAULT 'reserved',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- TLS certificates table (Cloudflare Origin CA certificates)
CREATE TABLE IF NOT EXISTS tls_certificates (
  id SERIAL PRIMARY KEY,
  user_email TEXT NOT NULL REFERENCES user_registrations(email) ON DELETE CASCADE,
  cloudflare_cert_id TEXT,
  certificate TEXT NOT NULL, -- PEM-encoded certificate
  private_key TEXT NOT NULL, -- PEM-encoded private key
  hostnames TEXT[] NOT NULL, -- List of hostnames covered by this certificate
  scope TEXT NOT NULL CHECK (scope IN ('installation', 'workmachine', 'workspace')) DEFAULT 'installation',
  scope_identifier TEXT, -- wm-user for workmachine, workspace name for workspace, null for installation
  parent_scope_identifier TEXT, -- wm-user for workspace scope, null for others
  valid_from TIMESTAMPTZ NOT NULL,
  valid_until TIMESTAMPTZ NOT NULL,
  generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_email, cloudflare_cert_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_registrations_installation_key ON user_registrations(installation_key);
CREATE INDEX IF NOT EXISTS idx_user_registrations_subdomain ON user_registrations(subdomain);
CREATE INDEX IF NOT EXISTS idx_ip_records_user_email ON ip_records(user_email);
CREATE INDEX IF NOT EXISTS idx_ip_records_type ON ip_records(type);
CREATE INDEX IF NOT EXISTS idx_domain_reservations_user_email ON domain_reservations(user_email);
CREATE INDEX IF NOT EXISTS idx_tls_certificates_user_email ON tls_certificates(user_email);
CREATE INDEX IF NOT EXISTS idx_tls_certificates_valid_until ON tls_certificates(valid_until);
CREATE INDEX IF NOT EXISTS idx_tls_certificates_scope ON tls_certificates(user_email, scope, scope_identifier, parent_scope_identifier);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers to automatically update updated_at
DROP TRIGGER IF EXISTS update_user_registrations_updated_at ON user_registrations;
CREATE TRIGGER update_user_registrations_updated_at
  BEFORE UPDATE ON user_registrations
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_ip_records_updated_at ON ip_records;
CREATE TRIGGER update_ip_records_updated_at
  BEFORE UPDATE ON ip_records
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_domain_reservations_updated_at ON domain_reservations;
CREATE TRIGGER update_domain_reservations_updated_at
  BEFORE UPDATE ON domain_reservations
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_tls_certificates_updated_at ON tls_certificates;
CREATE TRIGGER update_tls_certificates_updated_at
  BEFORE UPDATE ON tls_certificates
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

-- Row Level Security (optional, can enable later)
-- ALTER TABLE user_registrations ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE ip_records ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE domain_reservations ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE tls_certificates ENABLE ROW LEVEL SECURITY;
