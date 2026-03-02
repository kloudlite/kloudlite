-- Migration: Convert prices from USD to INR
-- Created: 2026-03-02
-- Description: Update subscription_plans prices to convert USD values to INR (1 USD ≈ 83 INR)
--
-- Current values are USD cents stored as INR paise:
--   Small: 2900 → Should be ~240700 ( $29 × 83 = ₹2407 = 240700 paise )
--   Medium: 4900 → Should be ~406700 ( $49 × 83 = ₹4067 = 406700 paise )
--   Large: 8900 → Should be ~738700 ( $89 × 83 = ₹7387 = 738700 paise )

-- ============================================================================
-- 1. UPDATE SUBSCRIPTION PLANS
-- ============================================================================

UPDATE subscription_plans
SET
  amount_per_user = ROUND(amount_per_user * 83),
  base_fee = ROUND(base_fee * 83)
WHERE tier IN (1, 2, 3);

-- ============================================================================
-- 2. VERIFY UPDATED VALUES
-- ============================================================================

SELECT
  tier,
  name,
  amount_per_user,
  base_fee,
  currency,
  ROUND(amount_per_user::numeric / 100.0, 2) as inr_amount_per_user,
  ROUND(base_fee::numeric / 100.0, 2) as inr_base_fee,
  ROUND((amount_per_user + base_fee)::numeric / 100.0, 2) as inr_total
FROM subscription_plans
ORDER BY tier;

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
