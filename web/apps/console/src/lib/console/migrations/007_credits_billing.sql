-- ============================================================================
-- Migration 007: Credit Accounts, Transactions, Usage Events & Pricing
-- Pay-as-you-go billing system tables, RPCs, and seed data
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. CREDIT ACCOUNTS (one per org)
-- ============================================================================

CREATE TABLE IF NOT EXISTS credit_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
  balance NUMERIC(12,4) NOT NULL DEFAULT 0,
  auto_topup_enabled BOOLEAN NOT NULL DEFAULT false,
  auto_topup_threshold NUMERIC(10,2),
  auto_topup_amount NUMERIC(10,2),
  stripe_customer_id TEXT UNIQUE,
  negative_balance_flagged BOOLEAN NOT NULL DEFAULT false,
  low_balance_warning BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================================
-- 2. CREDIT TRANSACTIONS (append-only ledger)
-- ============================================================================

CREATE TABLE IF NOT EXISTS credit_transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('topup', 'usage_debit', 'adjustment')),
  amount NUMERIC(12,4) NOT NULL,
  description TEXT,
  stripe_invoice_id TEXT,
  usage_period_id UUID,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_credit_transactions_org_id ON credit_transactions(org_id);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_expires_at ON credit_transactions(expires_at) WHERE expires_at IS NOT NULL AND type = 'topup';
CREATE INDEX IF NOT EXISTS idx_credit_transactions_created_at ON credit_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_type ON credit_transactions(type);

-- ============================================================================
-- 3. USAGE EVENTS (raw events from controllers)
-- ============================================================================

CREATE TABLE IF NOT EXISTS usage_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL CHECK (event_type IN (
    'workmachine.started',
    'workmachine.stopped',
    'workmachine.resized',
    'controlplane.started',
    'controlplane.stopped',
    'storage.provisioned',
    'storage.resized',
    'storage.deleted'
  )),
  resource_id TEXT NOT NULL,
  resource_type TEXT,
  metadata JSONB DEFAULT '{}',
  event_timestamp TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(installation_id, resource_id, event_type, event_timestamp)
);

CREATE INDEX IF NOT EXISTS idx_usage_events_installation_id ON usage_events(installation_id);
CREATE INDEX IF NOT EXISTS idx_usage_events_event_type ON usage_events(event_type);

-- ============================================================================
-- 4. USAGE PERIODS (active resource usage intervals for billing)
-- ============================================================================

CREATE TABLE IF NOT EXISTS usage_periods (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  resource_id TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL,
  ended_at TIMESTAMPTZ,
  hourly_rate NUMERIC(10,6) NOT NULL,
  total_cost NUMERIC(12,4) NOT NULL DEFAULT 0,
  last_billed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_usage_periods_active
  ON usage_periods(org_id)
  WHERE ended_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_usage_periods_installation_id ON usage_periods(installation_id);
CREATE INDEX IF NOT EXISTS idx_usage_periods_org_id ON usage_periods(org_id);

-- ============================================================================
-- 5. PRICING TIERS (configurable resource pricing)
-- ============================================================================

CREATE TABLE IF NOT EXISTS pricing_tiers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  resource_type TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  hourly_rate NUMERIC(10,6) NOT NULL,
  unit TEXT NOT NULL DEFAULT 'hour',
  category TEXT NOT NULL CHECK (category IN ('compute', 'storage')),
  specs JSONB DEFAULT '{}',
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================================
-- 6. UPDATED_AT TRIGGER ON CREDIT ACCOUNTS
-- ============================================================================

DROP TRIGGER IF EXISTS update_credit_accounts_updated_at ON credit_accounts;
CREATE TRIGGER update_credit_accounts_updated_at
  BEFORE UPDATE ON credit_accounts
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 7. ROW LEVEL SECURITY
-- ============================================================================

ALTER TABLE credit_accounts ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage credit_accounts" ON credit_accounts;
CREATE POLICY "Service role can manage credit_accounts"
  ON credit_accounts FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE credit_transactions ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage credit_transactions" ON credit_transactions;
CREATE POLICY "Service role can manage credit_transactions"
  ON credit_transactions FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE usage_events ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage usage_events" ON usage_events;
CREATE POLICY "Service role can manage usage_events"
  ON usage_events FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE usage_periods ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage usage_periods" ON usage_periods;
CREATE POLICY "Service role can manage usage_periods"
  ON usage_periods FOR ALL TO service_role
  USING (true) WITH CHECK (true);

ALTER TABLE pricing_tiers ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role can manage pricing_tiers" ON pricing_tiers;
CREATE POLICY "Service role can manage pricing_tiers"
  ON pricing_tiers FOR ALL TO service_role
  USING (true) WITH CHECK (true);

-- ============================================================================
-- 8. RPCs
-- ============================================================================

-- Atomically debit credits: decrements balance and records transaction
CREATE OR REPLACE FUNCTION debit_credits(
  p_org_id uuid,
  p_amount numeric,
  p_description text,
  p_usage_period_id uuid
) RETURNS numeric
LANGUAGE plpgsql
SET search_path = ''
AS $$
DECLARE
  v_new_balance numeric;
BEGIN
  UPDATE public.credit_accounts
  SET balance = balance - p_amount
  WHERE org_id = p_org_id
  RETURNING balance INTO v_new_balance;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Credit account not found for org %', p_org_id;
  END IF;

  INSERT INTO public.credit_transactions (org_id, type, amount, description, usage_period_id)
  VALUES (p_org_id, 'usage_debit', -p_amount, p_description, p_usage_period_id);

  -- Flag negative balance
  IF v_new_balance < 0 THEN
    UPDATE public.credit_accounts
    SET negative_balance_flagged = true
    WHERE org_id = p_org_id;
  END IF;

  RETURN v_new_balance;
END;
$$;

-- Atomically top up credits: increments balance, clears warnings, records transaction
CREATE OR REPLACE FUNCTION topup_credits(
  p_org_id uuid,
  p_amount numeric,
  p_description text,
  p_stripe_invoice_id text
) RETURNS numeric
LANGUAGE plpgsql
SET search_path = ''
AS $$
DECLARE
  v_new_balance numeric;
BEGIN
  UPDATE public.credit_accounts
  SET balance = balance + p_amount,
      negative_balance_flagged = false,
      low_balance_warning = false
  WHERE org_id = p_org_id
  RETURNING balance INTO v_new_balance;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Credit account not found for org %', p_org_id;
  END IF;

  INSERT INTO public.credit_transactions (org_id, type, amount, description, stripe_invoice_id)
  VALUES (p_org_id, 'topup', p_amount, p_description, p_stripe_invoice_id);

  RETURN v_new_balance;
END;
$$;

-- ============================================================================
-- 9. SEED DATA (pricing tiers)
-- ============================================================================

INSERT INTO pricing_tiers (resource_type, display_name, hourly_rate, unit, category, specs)
VALUES
  ('controlplane', 'Control Plane', 0.02, 'hour', 'compute', '{}')
ON CONFLICT (resource_type) DO NOTHING;

INSERT INTO pricing_tiers (resource_type, display_name, hourly_rate, unit, category, specs)
VALUES
  ('workmachine.tier1', 'Tier 1 — Light', 0.18, 'hour', 'compute', '{"vcpu": 8, "memory_gb": 16, "storage_gb": 100, "suspend_minutes": 15}')
ON CONFLICT (resource_type) DO UPDATE SET display_name = EXCLUDED.display_name, hourly_rate = EXCLUDED.hourly_rate, specs = EXCLUDED.specs;

INSERT INTO pricing_tiers (resource_type, display_name, hourly_rate, unit, category, specs)
VALUES
  ('workmachine.tier2', 'Tier 2 — Standard', 0.30, 'hour', 'compute', '{"vcpu": 12, "memory_gb": 32, "storage_gb": 200, "suspend_minutes": 30}')
ON CONFLICT (resource_type) DO UPDATE SET display_name = EXCLUDED.display_name, hourly_rate = EXCLUDED.hourly_rate, specs = EXCLUDED.specs;

INSERT INTO pricing_tiers (resource_type, display_name, hourly_rate, unit, category, specs)
VALUES
  ('workmachine.tier3', 'Tier 3 — Power', 0.55, 'hour', 'compute', '{"vcpu": 16, "memory_gb": 64, "storage_gb": 500, "suspend_minutes": 60}')
ON CONFLICT (resource_type) DO UPDATE SET display_name = EXCLUDED.display_name, hourly_rate = EXCLUDED.hourly_rate, specs = EXCLUDED.specs;

INSERT INTO pricing_tiers (resource_type, display_name, hourly_rate, unit, category, specs)
VALUES
  ('storage.vm', 'VM Storage', 0.000056, 'gb_hour', 'storage', '{"monthly_rate_per_gb": 0.04}')
ON CONFLICT (resource_type) DO NOTHING;

INSERT INTO pricing_tiers (resource_type, display_name, hourly_rate, unit, category, specs)
VALUES
  ('storage.object', 'Object Storage', 0.000056, 'gb_hour', 'storage', '{"monthly_rate_per_gb": 0.04}')
ON CONFLICT (resource_type) DO NOTHING;

COMMIT;
