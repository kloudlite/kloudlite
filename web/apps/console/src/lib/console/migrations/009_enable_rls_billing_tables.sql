-- Migration: Enable RLS on billing tables
-- Created: 2026-03-02
-- Description: Enable Row Level Security on subscription_plans, subscriptions, and invoices tables
--
-- Security: Ensure only authorized users can read/write billing data

-- ============================================================================
-- 1. ENABLE RLS ON BILLING TABLES
-- ============================================================================

ALTER TABLE subscription_plans ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- 2. SUBSCRIPTION PLANS POLICIES
-- ============================================================================
-- Read access: Allow authenticated users to see all plans
DROP POLICY IF EXISTS plans_select_policy ON subscription_plans;
CREATE POLICY plans_select_policy
ON subscription_plans
FOR SELECT
USING (auth.role() = 'authenticated');

-- ============================================================================
-- 3. SUBSCRIPTIONS POLICIES
-- ============================================================================
-- Read access: Allow members to see subscriptions for installations they're part of
DROP POLICY IF EXISTS subscriptions_select_policy ON subscriptions;
CREATE POLICY subscriptions_select_policy
ON subscriptions
FOR SELECT
USING (
  EXISTS (
    SELECT 1 FROM installation_members
    WHERE installation_members.installation_id = subscriptions.installation_id
    AND installation_members.user_id = auth.uid()
  )
);

-- Write access: Only installation owners can create/update subscriptions
DROP POLICY IF EXISTS subscriptions_write_policy ON subscriptions;
CREATE POLICY subscriptions_write_policy
ON subscriptions
FOR ALL
TO authenticated
USING (
  EXISTS (
    SELECT 1 FROM installations
    WHERE installations.id = subscriptions.installation_id
    AND installations.user_id = auth.uid()
  )
)
WITH CHECK (
  EXISTS (
    SELECT 1 FROM installations
    WHERE installations.id = NEW.installation_id
    AND installations.user_id = auth.uid()
  )
);

-- ============================================================================
-- 4. INVOICES POLICIES
-- ============================================================================
-- Read access: Allow members to see invoices for installations they're part of
DROP POLICY IF EXISTS invoices_select_policy ON invoices;
CREATE POLICY invoices_select_policy
ON invoices
FOR SELECT
USING (
  EXISTS (
    SELECT 1 FROM installation_members
    WHERE installation_members.installation_id = invoices.installation_id
    AND installation_members.user_id = auth.uid()
  )
);

-- Write access: Only installation owners can create/update invoices
DROP POLICY IF EXISTS invoices_write_policy ON invoices;
CREATE POLICY invoices_write_policy
ON invoices
FOR ALL
TO authenticated
USING (
  EXISTS (
    SELECT 1 FROM installations
    WHERE installations.id = invoices.installation_id
    AND installations.user_id = auth.uid()
  )
)
WITH CHECK (
  EXISTS (
    SELECT 1 FROM installations
    WHERE installations.id = NEW.installation_id
    AND installations.user_id = auth.uid()
  )
);

-- ============================================================================
-- 5. VERIFY RLS IS ENABLED
-- ============================================================================

SELECT
  schemaname,
  tablename,
  rowsecurity
FROM pg_tables
WHERE schemaname = 'public'
  AND tablename IN ('subscription_plans', 'subscriptions', 'invoices')
ORDER BY tablename;

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
