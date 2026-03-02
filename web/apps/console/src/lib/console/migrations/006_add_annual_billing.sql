-- Migration: Add annual billing support
-- Created: 2026-03-02
-- Description: Add billing_period to subscriptions and annual_discount_pct to subscription_plans

-- ============================================================================
-- 1. ADD BILLING PERIOD TO SUBSCRIPTIONS
-- ============================================================================

ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS billing_period TEXT NOT NULL DEFAULT 'monthly'
  CHECK (billing_period IN ('monthly', 'annual'));

-- ============================================================================
-- 2. ADD ANNUAL DISCOUNT PERCENTAGE TO SUBSCRIPTION PLANS
-- ============================================================================

ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS annual_discount_pct INTEGER NOT NULL DEFAULT 20;

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
