-- Migration: Replace Razorpay billing tables with Stripe tables
-- Fresh start — no data migration needed.

-- 1. Drop old billing tables (order matters for FKs)
DROP TABLE IF EXISTS processed_webhook_events CASCADE;
DROP TABLE IF EXISTS cron_job_logs CASCADE;
DROP TABLE IF EXISTS renewal_jobs CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS subscriptions CASCADE;
DROP TABLE IF EXISTS subscription_plans CASCADE;

-- 2. Remove pg_cron schedule if it exists
DO $$
BEGIN
  PERFORM cron.unschedule('billing-cron-job');
EXCEPTION WHEN OTHERS THEN
  NULL;
END $$;

-- 3. Create stripe_customers
CREATE TABLE stripe_customers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  stripe_customer_id TEXT NOT NULL UNIQUE,
  stripe_subscription_id TEXT UNIQUE,
  billing_status TEXT NOT NULL DEFAULT 'incomplete'
    CHECK (billing_status IN ('active', 'past_due', 'cancelled', 'trialing', 'incomplete')),
  payment_issue BOOLEAN NOT NULL DEFAULT false,
  current_period_end TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_stripe_customers_installation UNIQUE (installation_id)
);

-- 4. Create subscription_items (entitlements cache)
CREATE TABLE subscription_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  stripe_subscription_item_id TEXT NOT NULL UNIQUE,
  stripe_price_id TEXT NOT NULL,
  tier INT NOT NULL CHECK (tier >= 0 AND tier <= 3),
  product_name TEXT NOT NULL,
  quantity INT NOT NULL DEFAULT 1 CHECK (quantity >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_subscription_items_installation ON subscription_items(installation_id);

-- 5. Create stripe_webhook_events (idempotency)
CREATE TABLE stripe_webhook_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  stripe_event_id TEXT NOT NULL UNIQUE,
  event_type TEXT NOT NULL,
  processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 6. Enable RLS
ALTER TABLE stripe_customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscription_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE stripe_webhook_events ENABLE ROW LEVEL SECURITY;

-- RLS: stripe_customers readable by installation members
-- user_id is TEXT, auth.uid() returns UUID — cast to TEXT
CREATE POLICY stripe_customers_select ON stripe_customers
  FOR SELECT USING (
    installation_id IN (
      SELECT id FROM installations WHERE user_id = auth.uid()::text
    )
    OR installation_id IN (
      SELECT installation_id FROM installation_members WHERE user_id = auth.uid()::text
    )
  );

-- RLS: subscription_items readable by installation members
CREATE POLICY subscription_items_select ON subscription_items
  FOR SELECT USING (
    installation_id IN (
      SELECT id FROM installations WHERE user_id = auth.uid()::text
    )
    OR installation_id IN (
      SELECT installation_id FROM installation_members WHERE user_id = auth.uid()::text
    )
  );
