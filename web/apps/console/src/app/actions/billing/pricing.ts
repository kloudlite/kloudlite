'use server'

import { getActivePricingTiers } from '@/lib/stripe-bootstrap'
import type { PricingTier } from '@/lib/console/storage/credits-types'

/**
 * Fetch the active pricing tiers from the database.
 * Results are cached in memory on the server.
 */
export async function fetchPricingTiers(): Promise<PricingTier[]> {
  return await getActivePricingTiers()
}
