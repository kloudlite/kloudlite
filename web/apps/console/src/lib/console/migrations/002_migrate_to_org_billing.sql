-- ============================================================================
-- PART 1: Run this first
-- Creates org tables, populates from existing data, migrates columns
-- ============================================================================

-- 1. CREATE ORGANIZATION TABLES

CREATE TABLE IF NOT EXISTS organizations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  created_by TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);
CREATE INDEX IF NOT EXISTS idx_organizations_created_by ON organizations(created_by);

CREATE TABLE IF NOT EXISTS organization_members (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('owner', 'admin')),
  added_by TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(org_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_organization_members_org_id ON organization_members(org_id);
CREATE INDEX IF NOT EXISTS idx_organization_members_user_id ON organization_members(user_id);

CREATE TABLE IF NOT EXISTS organization_invitations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  email TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('admin')),
  invited_by TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'expired')),
  expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_invitations_unique_pending
  ON organization_invitations(org_id, email, status)
  WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_organization_invitations_org_id ON organization_invitations(org_id);
CREATE INDEX IF NOT EXISTS idx_organization_invitations_email ON organization_invitations(email);
CREATE INDEX IF NOT EXISTS idx_organization_invitations_status ON organization_invitations(status);

-- 2. AUTO-CREATE ORGS FROM EXISTING INSTALLATIONS
-- One org per unique user_id, using sanitized user_id as slug

INSERT INTO organizations (name, slug, created_by)
SELECT
  trim(TRAILING '-' FROM left(
    CASE
      WHEN regexp_replace(regexp_replace(regexp_replace(lower(user_id), '[^a-z0-9-]', '-', 'g'), '-+', '-', 'g'), '^-|-$', '', 'g') ~ '^[a-z]'
      THEN regexp_replace(regexp_replace(regexp_replace(lower(user_id), '[^a-z0-9-]', '-', 'g'), '-+', '-', 'g'), '^-|-$', '', 'g')
      ELSE 'org-' || regexp_replace(regexp_replace(regexp_replace(lower(user_id), '[^a-z0-9-]', '-', 'g'), '-+', '-', 'g'), '^-|-$', '', 'g')
    END,
  63)) || '''s Org' AS name,
  trim(TRAILING '-' FROM left(
    CASE
      WHEN regexp_replace(regexp_replace(regexp_replace(lower(user_id), '[^a-z0-9-]', '-', 'g'), '-+', '-', 'g'), '^-|-$', '', 'g') ~ '^[a-z]'
      THEN regexp_replace(regexp_replace(regexp_replace(lower(user_id), '[^a-z0-9-]', '-', 'g'), '-+', '-', 'g'), '^-|-$', '', 'g')
      ELSE 'org-' || regexp_replace(regexp_replace(regexp_replace(lower(user_id), '[^a-z0-9-]', '-', 'g'), '-+', '-', 'g'), '^-|-$', '', 'g')
    END,
  63)) AS slug,
  user_id AS created_by
FROM (SELECT DISTINCT user_id FROM installations) t;

-- Add owner membership for each auto-created org
INSERT INTO organization_members (org_id, user_id, role, added_by)
SELECT o.id, o.created_by, 'owner', o.created_by
FROM organizations o;

-- 3. DROP OLD RLS POLICIES that reference user_id / old tables
DROP POLICY IF EXISTS "Users can read own installations" ON installations;
DROP POLICY IF EXISTS "Users can manage own installations" ON installations;
DROP POLICY IF EXISTS "Users can insert installations" ON installations;
DROP POLICY IF EXISTS "Users can update own installations" ON installations;
DROP POLICY IF EXISTS "Users can delete own installations" ON installations;
DROP POLICY IF EXISTS "Service role can manage installations" ON installations;
DROP POLICY IF EXISTS "Users can read installation members" ON installation_members;
DROP POLICY IF EXISTS "Users can read installation invitations" ON installation_invitations;
DROP POLICY IF EXISTS "stripe_customers_select" ON billing_accounts;
DROP POLICY IF EXISTS "Service role can manage billing_accounts" ON billing_accounts;
DROP POLICY IF EXISTS "subscription_items_select" ON subscription_items;
DROP POLICY IF EXISTS "Service role can manage subscription_items" ON subscription_items;
DROP POLICY IF EXISTS "Service role can manage dns_configurations" ON dns_configurations;
DROP POLICY IF EXISTS "Service role can manage domain_reservations" ON domain_reservations;
DROP POLICY IF EXISTS "Service role can manage processed_webhook_events" ON processed_webhook_events;

-- 4. ADD org_id TO INSTALLATIONS
ALTER TABLE installations ADD COLUMN IF NOT EXISTS org_id UUID;

UPDATE installations i
SET org_id = om.org_id
FROM organization_members om
WHERE om.user_id = i.user_id AND om.role = 'owner';

ALTER TABLE installations ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE installations ADD CONSTRAINT fk_installations_org
  FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_installations_org_id ON installations(org_id);

ALTER TABLE installations DROP COLUMN IF EXISTS user_id;

-- 5. MIGRATE BILLING_ACCOUNTS: installation_id -> org_id
ALTER TABLE billing_accounts ADD COLUMN IF NOT EXISTS org_id UUID;

UPDATE billing_accounts ba
SET org_id = i.org_id
FROM installations i
WHERE ba.installation_id = i.id;

DELETE FROM billing_accounts WHERE org_id IS NULL;

ALTER TABLE billing_accounts ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE billing_accounts ADD CONSTRAINT fk_billing_accounts_org
  FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE billing_accounts DROP CONSTRAINT IF EXISTS billing_accounts_installation_id_key;
ALTER TABLE billing_accounts DROP CONSTRAINT IF EXISTS uq_billing_accounts_installation;
ALTER TABLE billing_accounts DROP COLUMN IF EXISTS installation_id;
ALTER TABLE billing_accounts ADD CONSTRAINT uq_billing_accounts_org UNIQUE (org_id);

-- 6. MIGRATE SUBSCRIPTION_ITEMS: add org_id, make installation_id nullable
ALTER TABLE subscription_items ADD COLUMN IF NOT EXISTS org_id UUID;

UPDATE subscription_items si
SET org_id = i.org_id
FROM installations i
WHERE si.installation_id = i.id;

DELETE FROM subscription_items WHERE org_id IS NULL;

ALTER TABLE subscription_items ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE subscription_items ADD CONSTRAINT fk_subscription_items_org
  FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_subscription_items_org ON subscription_items(org_id);

ALTER TABLE subscription_items ALTER COLUMN installation_id DROP NOT NULL;
ALTER TABLE subscription_items ADD CONSTRAINT fk_subscription_items_installation
  FOREIGN KEY (installation_id) REFERENCES installations(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_subscription_items_installation ON subscription_items(installation_id);

-- 7. DROP OLD TABLES
DROP TABLE IF EXISTS installation_invitations CASCADE;
DROP TABLE IF EXISTS installation_members CASCADE;

-- 8. ROW LEVEL SECURITY (service-role only for all tables)
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Service role can manage organizations"
  ON organizations FOR ALL TO service_role USING (true) WITH CHECK (true);

ALTER TABLE organization_members ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Service role can manage organization_members"
  ON organization_members FOR ALL TO service_role USING (true) WITH CHECK (true);

ALTER TABLE organization_invitations ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Service role can manage organization_invitations"
  ON organization_invitations FOR ALL TO service_role USING (true) WITH CHECK (true);

ALTER TABLE installations ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Service role can manage installations"
  ON installations FOR ALL TO service_role USING (true) WITH CHECK (true);

ALTER TABLE dns_configurations ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Service role can manage dns_configurations"
  ON dns_configurations FOR ALL TO service_role USING (true) WITH CHECK (true);

ALTER TABLE domain_reservations ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Service role can manage domain_reservations"
  ON domain_reservations FOR ALL TO service_role USING (true) WITH CHECK (true);

ALTER TABLE billing_accounts ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Service role can manage billing_accounts"
  ON billing_accounts FOR ALL TO service_role USING (true) WITH CHECK (true);

ALTER TABLE subscription_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Service role can manage subscription_items"
  ON subscription_items FOR ALL TO service_role USING (true) WITH CHECK (true);

ALTER TABLE processed_webhook_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Service role can manage processed_webhook_events"
  ON processed_webhook_events FOR ALL TO service_role USING (true) WITH CHECK (true);
