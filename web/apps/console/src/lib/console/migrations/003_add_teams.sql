-- Migration: Add teams and invitations support
-- Created: 2026-01-17
-- Description: Add installation_members and installation_invitations tables for team collaboration

-- ============================================================================
-- 1. INSTALLATION MEMBERS TABLE
-- Team members with roles for each installation
-- ============================================================================

CREATE TABLE IF NOT EXISTS installation_members (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  user_id TEXT NOT NULL REFERENCES user_registrations(user_id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
  added_by TEXT REFERENCES user_registrations(user_id),
  added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(installation_id, user_id)
);

-- Indexes for installation_members
CREATE INDEX IF NOT EXISTS idx_installation_members_installation_id ON installation_members(installation_id);
CREATE INDEX IF NOT EXISTS idx_installation_members_user_id ON installation_members(user_id);
CREATE INDEX IF NOT EXISTS idx_installation_members_role ON installation_members(role);

-- ============================================================================
-- 2. INSTALLATION INVITATIONS TABLE
-- Pending invitations for joining installation teams
-- ============================================================================

CREATE TABLE IF NOT EXISTS installation_invitations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  email TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('admin', 'member', 'viewer')),
  invited_by TEXT NOT NULL REFERENCES user_registrations(user_id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'expired')),
  expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create partial unique index for pending invitations only
CREATE UNIQUE INDEX IF NOT EXISTS idx_installation_invitations_unique_pending
  ON installation_invitations(installation_id, email, status)
  WHERE status = 'pending';

-- Indexes for installation_invitations
CREATE INDEX IF NOT EXISTS idx_installation_invitations_installation_id ON installation_invitations(installation_id);
CREATE INDEX IF NOT EXISTS idx_installation_invitations_email ON installation_invitations(email);
CREATE INDEX IF NOT EXISTS idx_installation_invitations_status ON installation_invitations(status);
CREATE INDEX IF NOT EXISTS idx_installation_invitations_expires_at ON installation_invitations(expires_at);

-- ============================================================================
-- 3. TRIGGERS FOR UPDATED_AT TIMESTAMPS
-- ============================================================================

-- Drop triggers if they exist, then create them
DROP TRIGGER IF EXISTS update_installation_members_updated_at ON installation_members;
CREATE TRIGGER update_installation_members_updated_at
    BEFORE UPDATE ON installation_members
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_installation_invitations_updated_at ON installation_invitations;
CREATE TRIGGER update_installation_invitations_updated_at
    BEFORE UPDATE ON installation_invitations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 4. DATA MIGRATION
-- Add existing installation owners as 'owner' role members
-- ============================================================================

-- Migrate existing installations: Add current owner as 'owner' role member
INSERT INTO installation_members (installation_id, user_id, role, added_by, added_at, created_at)
SELECT
  id as installation_id,
  user_id,
  'owner' as role,
  user_id as added_by,  -- Self-added
  created_at as added_at,
  created_at
FROM installations
WHERE NOT EXISTS (
  SELECT 1 FROM installation_members
  WHERE installation_members.installation_id = installations.id
    AND installation_members.user_id = installations.user_id
)
ON CONFLICT (installation_id, user_id) DO NOTHING;

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================

-- Verify new tables were created
SELECT
  table_name,
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = t.table_name) as column_count
FROM information_schema.tables t
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE'
  AND table_name IN (
    'installation_members',
    'installation_invitations'
  )
ORDER BY table_name;
