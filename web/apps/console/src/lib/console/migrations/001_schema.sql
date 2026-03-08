-- ============================================================================
-- Kloudlite Console Database Schema (squashed — main/operational DB)
-- PII tables (users, magic_link_tokens, contact_messages) live in the PII DB.
-- All tables, indexes, triggers, and RLS policies
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. ORGANIZATIONS
-- ============================================================================

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

-- ============================================================================
-- 2. ORGANIZATION MEMBERS
-- ============================================================================

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

-- ============================================================================
-- 3. ORGANIZATION INVITATIONS
-- ============================================================================

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

-- ============================================================================
-- 4. INSTALLATIONS
-- ============================================================================

CREATE TABLE IF NOT EXISTS installations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT,
  installation_key TEXT NOT NULL UNIQUE,
  secret_key TEXT,
  setup_completed BOOLEAN NOT NULL DEFAULT FALSE,
  subdomain TEXT UNIQUE,
  reserved_at TIMESTAMPTZ,
  deployment_ready BOOLEAN DEFAULT FALSE,
  last_health_check TIMESTAMPTZ,
  cloud_provider TEXT CHECK (cloud_provider IN ('aws', 'gcp', 'azure', 'oci')),
  cloud_location TEXT,
  deploy_job_execution_name TEXT,
  deploy_job_status TEXT CHECK (deploy_job_status IN ('pending', 'running', 'succeeded', 'failed', 'unknown')),
  deploy_job_started_at TIMESTAMPTZ,
  deploy_job_completed_at TIMESTAMPTZ,
  deploy_job_error TEXT,
  deploy_job_operation TEXT CHECK (deploy_job_operation IN ('install', 'uninstall')),
  deploy_job_current_step INTEGER,
  deploy_job_total_steps INTEGER,
  deploy_job_step_description TEXT,
  root_dns_target TEXT,
  root_dns_type TEXT CHECK (root_dns_type IN ('cname', 'a')),
  root_dns_record_id TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_installations_org_id ON installations(org_id);
CREATE INDEX IF NOT EXISTS idx_installations_subdomain ON installations(subdomain);
CREATE INDEX IF NOT EXISTS idx_installations_installation_key ON installations(installation_key);

-- ============================================================================
-- 5. DNS CONFIGURATIONS
-- ============================================================================

CREATE TABLE IF NOT EXISTS dns_configurations (
  id SERIAL PRIMARY KEY,
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  service_name TEXT NOT NULL,
  ip TEXT NOT NULL,
  ssh_record_id TEXT,
  route_record_ids TEXT[] DEFAULT '{}',
  route_record_map JSONB DEFAULT '{}'::jsonb,
  domain_routes JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(installation_id, service_name)
);

CREATE INDEX IF NOT EXISTS idx_dns_configurations_installation_id ON dns_configurations(installation_id);
CREATE INDEX IF NOT EXISTS idx_dns_configurations_service_name ON dns_configurations(service_name);

COMMENT ON COLUMN dns_configurations.route_record_map IS
  'Mapping of domain names to Cloudflare DNS record IDs for efficient differential updates';

-- ============================================================================
-- 6. DOMAIN RESERVATIONS
-- ============================================================================

CREATE TABLE IF NOT EXISTS domain_reservations (
  subdomain TEXT PRIMARY KEY,
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  user_id TEXT NOT NULL,
  user_email TEXT NOT NULL,
  user_name TEXT NOT NULL,
  reserved_at TIMESTAMPTZ DEFAULT NOW(),
  status TEXT NOT NULL DEFAULT 'reserved' CHECK (status IN ('reserved', 'active', 'cancelled')),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_domain_reservations_installation_id ON domain_reservations(installation_id);
CREATE INDEX IF NOT EXISTS idx_domain_reservations_user_id ON domain_reservations(user_id);

-- ============================================================================
-- 7. BILLING ACCOUNTS (one per org)
-- ============================================================================

CREATE TABLE IF NOT EXISTS billing_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  stripe_customer_id TEXT NOT NULL UNIQUE,
  stripe_subscription_id TEXT UNIQUE,
  billing_status TEXT NOT NULL DEFAULT 'incomplete'
    CHECK (billing_status IN ('active', 'past_due', 'cancelled', 'trialing', 'incomplete')),
  has_payment_issue BOOLEAN NOT NULL DEFAULT false,
  current_period_end TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_billing_accounts_org UNIQUE (org_id)
);

-- ============================================================================
-- 8. SUBSCRIPTION ITEMS (scoped to org, optional installation reference)
-- ============================================================================

CREATE TABLE IF NOT EXISTS subscription_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  installation_id UUID REFERENCES installations(id) ON DELETE SET NULL,
  stripe_item_id TEXT NOT NULL UNIQUE,
  stripe_price_id TEXT NOT NULL,
  tier INT NOT NULL CHECK (tier >= 0 AND tier <= 3),
  product_name TEXT NOT NULL,
  quantity INT NOT NULL DEFAULT 1 CHECK (quantity >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscription_items_org ON subscription_items(org_id);
CREATE INDEX IF NOT EXISTS idx_subscription_items_installation ON subscription_items(installation_id);

-- ============================================================================
-- 9. PROCESSED WEBHOOK EVENTS (idempotency guard)
-- ============================================================================

CREATE TABLE IF NOT EXISTS processed_webhook_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  stripe_event_id TEXT NOT NULL UNIQUE,
  event_type TEXT NOT NULL,
  processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 10. TRIGGERS FOR UPDATED_AT TIMESTAMPS
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = ''
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_organization_members_updated_at ON organization_members;
CREATE TRIGGER update_organization_members_updated_at
    BEFORE UPDATE ON organization_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_organization_invitations_updated_at ON organization_invitations;
CREATE TRIGGER update_organization_invitations_updated_at
    BEFORE UPDATE ON organization_invitations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_installations_updated_at ON installations;
CREATE TRIGGER update_installations_updated_at
    BEFORE UPDATE ON installations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_dns_configurations_updated_at ON dns_configurations;
CREATE TRIGGER update_dns_configurations_updated_at
    BEFORE UPDATE ON dns_configurations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_domain_reservations_updated_at ON domain_reservations;
CREATE TRIGGER update_domain_reservations_updated_at
    BEFORE UPDATE ON domain_reservations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 11. ROW LEVEL SECURITY (all service-role-only; app-level auth via PII DB)
-- ============================================================================

ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage organizations" ON organizations;
CREATE POLICY "Service role can manage organizations"
  ON organizations FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE organization_members ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage organization_members" ON organization_members;
CREATE POLICY "Service role can manage organization_members"
  ON organization_members FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE organization_invitations ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage organization_invitations" ON organization_invitations;
CREATE POLICY "Service role can manage organization_invitations"
  ON organization_invitations FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE installations ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage installations" ON installations;
CREATE POLICY "Service role can manage installations"
  ON installations FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE dns_configurations ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage dns_configurations" ON dns_configurations;
CREATE POLICY "Service role can manage dns_configurations"
  ON dns_configurations FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE domain_reservations ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage domain_reservations" ON domain_reservations;
CREATE POLICY "Service role can manage domain_reservations"
  ON domain_reservations FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE billing_accounts ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage billing_accounts" ON billing_accounts;
CREATE POLICY "Service role can manage billing_accounts"
  ON billing_accounts FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE subscription_items ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage subscription_items" ON subscription_items;
CREATE POLICY "Service role can manage subscription_items"
  ON subscription_items FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE processed_webhook_events ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage processed_webhook_events" ON processed_webhook_events;
CREATE POLICY "Service role can manage processed_webhook_events"
  ON processed_webhook_events FOR ALL TO service_role
  USING (true) WITH CHECK (true);

-- ============================================================================
-- 12. STORED PROCEDURES (atomic multi-table operations)
-- ============================================================================

-- Atomic org creation: creates org + owner member in one transaction
CREATE OR REPLACE FUNCTION create_organization_with_owner(
  p_name text, p_slug text, p_created_by text
) RETURNS uuid
LANGUAGE plpgsql
SET search_path = ''
AS $$
DECLARE
  v_org_id uuid;
BEGIN
  INSERT INTO public.organizations (name, slug, created_by)
  VALUES (p_name, p_slug, p_created_by)
  RETURNING id INTO v_org_id;

  INSERT INTO public.organization_members (org_id, user_id, role, added_by)
  VALUES (v_org_id, p_created_by, 'owner', p_created_by);

  RETURN v_org_id;
END;
$$;

-- Atomic ownership transfer: demotes old owner + promotes new owner in one transaction
CREATE OR REPLACE FUNCTION transfer_org_ownership(
  p_org_id uuid, p_old_owner text, p_new_owner text
) RETURNS void
LANGUAGE plpgsql
SET search_path = ''
AS $$
BEGIN
  UPDATE public.organization_members
  SET role = 'admin', updated_at = now()
  WHERE org_id = p_org_id AND user_id = p_old_owner AND role = 'owner';

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Current owner not found';
  END IF;

  UPDATE public.organization_members
  SET role = 'owner', updated_at = now()
  WHERE org_id = p_org_id AND user_id = p_new_owner;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'New owner not found in organization';
  END IF;
END;
$$;

-- Atomic subscription items sync: deletes old items and inserts new ones in one transaction
CREATE OR REPLACE FUNCTION sync_subscription_items(
  p_org_id uuid,
  p_items jsonb
) RETURNS void
LANGUAGE plpgsql
SET search_path = ''
AS $$
BEGIN
  DELETE FROM public.subscription_items WHERE org_id = p_org_id;

  INSERT INTO public.subscription_items (org_id, installation_id, stripe_item_id, stripe_price_id, tier, product_name, quantity)
  SELECT
    p_org_id,
    (item->>'installation_id')::uuid,
    item->>'stripe_item_id',
    item->>'stripe_price_id',
    (item->>'tier')::integer,
    item->>'product_name',
    (item->>'quantity')::integer
  FROM jsonb_array_elements(p_items) AS item;
END;
$$;

-- Auto-update updated_at for billing tables
CREATE TRIGGER update_billing_accounts_updated_at
  BEFORE UPDATE ON billing_accounts
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscription_items_updated_at
  BEFORE UPDATE ON subscription_items
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
