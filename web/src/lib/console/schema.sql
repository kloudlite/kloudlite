-- ============================================================================
-- Kloudlite Registration Database Schema
-- Multi-Installation Support
-- ============================================================================

-- Drop existing tables (CASCADE will drop dependent objects)
DROP TABLE IF EXISTS contact_submissions CASCADE;
DROP TABLE IF EXISTS tls_certificates CASCADE;
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
  poller_active BOOLEAN DEFAULT FALSE,
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
-- Stores IP addresses for installations and workmachines
-- ============================================================================

CREATE TABLE ip_records (
  id SERIAL PRIMARY KEY,
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('installation', 'workmachine')),
  ip TEXT NOT NULL,
  work_machine_name TEXT,
  configured_at TIMESTAMPTZ DEFAULT NOW(),
  dns_record_ids TEXT[] DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE (installation_id, type, work_machine_name)
);

-- Indexes for ip_records
CREATE INDEX idx_ip_records_installation_id ON ip_records(installation_id);

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
-- 5. TLS CERTIFICATES TABLE
-- Stores TLS certificates for installations
-- ============================================================================

CREATE TABLE tls_certificates (
  id SERIAL PRIMARY KEY,
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  cloudflare_cert_id TEXT,
  certificate TEXT NOT NULL,
  private_key TEXT NOT NULL,
  hostnames TEXT[] NOT NULL,
  scope TEXT NOT NULL CHECK (scope IN ('installation', 'workmachine', 'workspace')),
  scope_identifier TEXT,
  parent_scope_identifier TEXT,
  valid_from TIMESTAMPTZ NOT NULL,
  valid_until TIMESTAMPTZ NOT NULL,
  generated_at TIMESTAMPTZ DEFAULT NOW(),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for tls_certificates
CREATE INDEX idx_tls_certificates_installation_id ON tls_certificates(installation_id);
CREATE INDEX idx_tls_certificates_scope ON tls_certificates(scope);

-- ============================================================================
-- 6. TRIGGERS FOR UPDATED_AT TIMESTAMPS
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

CREATE TRIGGER update_tls_certificates_updated_at
    BEFORE UPDATE ON tls_certificates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 7. OPTIONAL: ROW LEVEL SECURITY (RLS)
-- Uncomment if you want to enable RLS
-- ============================================================================

-- Enable RLS on tables
-- ALTER TABLE user_registrations ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE installations ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE ip_records ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE domain_reservations ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE tls_certificates ENABLE ROW LEVEL SECURITY;

-- Example RLS policies
-- CREATE POLICY "Users can view their own data" ON user_registrations
--   FOR SELECT
--   USING (user_id = current_setting('app.current_user_id')::TEXT);

-- CREATE POLICY "Users can view their own installations" ON installations
--   FOR SELECT
--   USING (user_id = current_setting('app.current_user_id')::TEXT);

-- ============================================================================
-- 6. CONTACT SUBMISSIONS TABLE
-- Stores contact form submissions from website
-- ============================================================================

CREATE TABLE contact_submissions (
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
CREATE INDEX idx_contact_submissions_email ON contact_submissions(email);
CREATE INDEX idx_contact_submissions_submitted_at ON contact_submissions(submitted_at DESC);

-- ============================================================================
-- 7. TRIGGERS FOR UPDATED_AT TIMESTAMPS (Contact Submissions)
-- ============================================================================

CREATE TRIGGER update_contact_submissions_updated_at
    BEFORE UPDATE ON contact_submissions
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
    'domain_reservations',
    'tls_certificates',
    'contact_submissions'
  )
ORDER BY table_name;
