import { getPricingTiers } from '@/lib/console/storage/credits'
import type { PricingTier } from '@/lib/console/storage/credits-types'

let cachedTiers: PricingTier[] | null = null

/**
 * Get active pricing tiers from the database.
 * Results are cached in memory for the lifetime of the server process.
 */
export async function getActivePricingTiers(): Promise<PricingTier[]> {
  if (cachedTiers) return cachedTiers
  cachedTiers = await getPricingTiers()
  return cachedTiers
}

/**
 * Clear the pricing tier cache. Call this if tiers are updated.
 */
export function clearPricingCache(): void {
  cachedTiers = null
}

/**
 * Calculate projected monthly cost for a set of resources assuming 24/7 uptime.
 * Used by the installation form to show estimated costs.
 */
export function calculateProjectedMonthlyCost(
  tiers: PricingTier[],
  selectedResources: Array<{
    resourceType: string
    quantity?: number
    sizeGb?: number
  }>
): number {
  let total = 0
  for (const resource of selectedResources) {
    const tier = tiers.find(t => t.resourceType === resource.resourceType)
    if (!tier) continue
    if (tier.category === 'storage') {
      // Storage: rate is per GB-hour, multiply by size and hours in month
      total += tier.hourlyRate * (resource.sizeGb ?? 0) * 24 * 30
    } else {
      // Compute: rate is per hour, multiply by quantity and hours in month
      total += tier.hourlyRate * (resource.quantity ?? 1) * 24 * 30
    }
  }
  return total
}

/**
 * Calculate the minimum top-up amount for a set of resources.
 * This is the projected 30-day cost minus current balance.
 */
export function calculateMinimumTopup(
  projectedMonthlyCost: number,
  currentBalance: number
): number {
  const needed = projectedMonthlyCost - currentBalance
  return Math.max(needed, 5) // Minimum $5 top-up
}
