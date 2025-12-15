-- ============================================================================
-- Kloudlite Registration Database Schema
-- Multi-Installation Support
-- Updated: 2025-12-15 (removed certificate tables - TLS termination at Cloudflare)
-- ============================================================================

-- Drop existing tables (CASCADE will drop dependent objects)
DROP TABLE IF EXISTS ip_records CASCADE;
DROP TABLE IF EXISTS domain_reservations CASCADE;
DROP TABLE IF EXISTS installations CASCADE;
DROP TABLE IF EXISTS user_registrations CASCADE;

-- ============================================================================
-- 1. USER REGISTRATIONS TABLE
-- Stores user authentication data only
-- ============================================================================

CREATE TABLE user_registrations (
  user_id TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  providers TEXT[] DEFAULT '{}',
  registered_at TIMESTAMPTZ DEFAULT NOW(),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for user_registrations
CREATE INDEX idx_user_registrations_email ON user_registrations(email);

-- ============================================================================
-- 2. INSTALLATIONS TABLE
-- Stores installation-specific data (multiple per user)
-- ============================================================================

CREATE TABLE installations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id TEXT NOT NULL REFERENCES user_registrations(user_id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT,
  installation_key TEXT NOT NULL UNIQUE,
  secret_key TEXT,
  has_completed_installation BOOLEAN NOT NULL DEFAULT FALSE,
  subdomain TEXT UNIQUE,
  reserved_at TIMESTAMPTZ,
  deployment_ready BOOLEAN DEFAULT FALSE,
  last_health_check TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for installations
CREATE INDEX idx_installations_user_id ON installations(user_id);
CREATE INDEX idx_installations_subdomain ON installations(subdomain);
CREATE INDEX idx_installations_installation_key ON installations(installation_key);

-- ============================================================================
-- 3. IP RECORDS TABLE
-- Stores IP addresses and DNS records for domain requests
-- ============================================================================

CREATE TABLE ip_records (
  id SERIAL PRIMARY KEY,
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  domain_request_name TEXT NOT NULL,
  ip TEXT NOT NULL,
  configured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ssh_record_id TEXT,
  route_record_ids TEXT[] DEFAULT '{}',
  route_record_map JSONB DEFAULT '{}'::jsonb,
  domain_routes JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(installation_id, domain_request_name)
);

-- Indexes for ip_records
CREATE INDEX idx_ip_records_installation_id ON ip_records(installation_id);
CREATE INDEX idx_ip_records_domain_request_name ON ip_records(domain_request_name);

-- Comments
COMMENT ON COLUMN ip_records.route_record_map IS
'Mapping of domain names to Cloudflare DNS record IDs for efficient differential updates';

-- ============================================================================
-- 4. DOMAIN RESERVATIONS TABLE
-- Stores subdomain reservations for installations
-- ============================================================================

CREATE TABLE domain_reservations (
  subdomain TEXT PRIMARY KEY,
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  user_id TEXT NOT NULL REFERENCES user_registrations(user_id) ON DELETE CASCADE,
  user_email TEXT NOT NULL,
  user_name TEXT NOT NULL,
  reserved_at TIMESTAMPTZ DEFAULT NOW(),
  status TEXT NOT NULL DEFAULT 'reserved' CHECK (status IN ('reserved', 'active', 'cancelled')),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for domain_reservations
CREATE INDEX idx_domain_reservations_installation_id ON domain_reservations(installation_id);
CREATE INDEX idx_domain_reservations_user_id ON domain_reservations(user_id);

-- ============================================================================
-- 5. TRIGGERS FOR UPDATED_AT TIMESTAMPS
-- ============================================================================

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers to all tables
CREATE TRIGGER update_user_registrations_updated_at
    BEFORE UPDATE ON user_registrations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_installations_updated_at
    BEFORE UPDATE ON installations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ip_records_updated_at
    BEFORE UPDATE ON ip_records
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_domain_reservations_updated_at
    BEFORE UPDATE ON domain_reservations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- SCHEMA SETUP COMPLETE
-- ============================================================================

-- Verify tables were created
SELECT
  table_name,
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = t.table_name) as column_count
FROM information_schema.tables t
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE'
  AND table_name IN (
    'user_registrations',
    'installations',
    'ip_records',
    'domain_reservations'
  )
ORDER BY table_name;
