ALTER TABLE subscriptions
  ADD COLUMN IF NOT EXISTS scheduled_billing_period TEXT
  CHECK (scheduled_billing_period IN ('monthly', 'annual'));
