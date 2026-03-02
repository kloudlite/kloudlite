-- Migration: Add billing tables for subscription management
-- Created: 2026-03-02
-- Description: Add subscription_plans, subscriptions, and invoices tables for Razorpay billing integration

-- ============================================================================
-- 1. SUBSCRIPTION PLANS TABLE
-- Mirrors Razorpay plans with tier-based pricing
-- ============================================================================

CREATE TABLE IF NOT EXISTS subscription_plans (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  razorpay_plan_id TEXT UNIQUE,
  tier INTEGER NOT NULL CHECK (tier IN (1, 2, 3)),
  name TEXT NOT NULL,
  amount_per_user INTEGER NOT NULL,
  base_fee INTEGER NOT NULL DEFAULT 2900,
  currency TEXT NOT NULL DEFAULT 'INR',
  monthly_hours INTEGER NOT NULL DEFAULT 160,
  overage_rate INTEGER NOT NULL DEFAULT 0,
  cpu INTEGER NOT NULL,
  ram TEXT NOT NULL,
  storage TEXT NOT NULL,
  auto_suspend TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 2. SUBSCRIPTIONS TABLE
-- One subscription per installation
-- ============================================================================

CREATE TABLE IF NOT EXISTS subscriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  plan_id UUID NOT NULL REFERENCES subscription_plans(id),
  razorpay_subscription_id TEXT UNIQUE,
  razorpay_customer_id TEXT,
  status TEXT NOT NULL DEFAULT 'created' CHECK (status IN ('created', 'authenticated', 'active', 'paused', 'cancelled', 'expired')),
  quantity INTEGER NOT NULL DEFAULT 1,
  current_start TIMESTAMPTZ,
  current_end TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(installation_id, plan_id)
);

-- Indexes for subscriptions
CREATE INDEX IF NOT EXISTS idx_subscriptions_installation ON subscriptions(installation_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_razorpay ON subscriptions(razorpay_subscription_id);

-- ============================================================================
-- 3. INVOICES TABLE
-- Synced from Razorpay webhooks
-- ============================================================================

CREATE TABLE IF NOT EXISTS invoices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  razorpay_invoice_id TEXT UNIQUE,
  razorpay_payment_id TEXT,
  amount INTEGER NOT NULL,
  currency TEXT NOT NULL DEFAULT 'INR',
  status TEXT NOT NULL DEFAULT 'issued' CHECK (status IN ('issued', 'paid', 'expired', 'cancelled')),
  billing_start TIMESTAMPTZ,
  billing_end TIMESTAMPTZ,
  paid_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for invoices
CREATE INDEX IF NOT EXISTS idx_invoices_subscription ON invoices(subscription_id);
CREATE INDEX IF NOT EXISTS idx_invoices_installation ON invoices(installation_id);

-- ============================================================================
-- 4. TRIGGERS FOR UPDATED_AT TIMESTAMPS
-- ============================================================================

DROP TRIGGER IF EXISTS update_subscription_plans_updated_at ON subscription_plans;
CREATE TRIGGER update_subscription_plans_updated_at
    BEFORE UPDATE ON subscription_plans
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON subscriptions;
CREATE TRIGGER update_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 5. SEED DATA
-- Subscription plans (amounts in paise: ₹29=2900, ₹49=4900, ₹89=8900)
-- ============================================================================

INSERT INTO subscription_plans (tier, name, amount_per_user, base_fee, currency, monthly_hours, overage_rate, cpu, ram, storage, auto_suspend, description) VALUES
  (1, 'Small', 2900, 2900, 'INR', 160, 18, 8, '16 GB', '100 GB', '15 min', '8 vCPUs, 16 GB RAM — suitable for lightweight development'),
  (2, 'Medium', 4900, 2900, 'INR', 160, 30, 12, '32 GB', '200 GB', '30 min', '12 vCPUs, 32 GB RAM — balanced for most development workflows'),
  (3, 'Large', 8900, 2900, 'INR', 160, 55, 16, '64 GB', '500 GB', '1 hr', '16 vCPUs, 64 GB RAM — for resource-intensive workloads');

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
    'subscription_plans',
    'subscriptions',
    'invoices'
  )
ORDER BY table_name;
