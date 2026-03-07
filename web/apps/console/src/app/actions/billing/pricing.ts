'use server'

import { getTierConfig } from '@/lib/stripe-bootstrap'

export interface TierConfigItem {
  tier: number
  name: string
  description: string
  fixed: boolean
  priceId: string
  pricePerUnit: number
}

/**
 * Fetch the tier pricing configuration for a given currency.
 * Reads from the in-memory cache populated at server startup.
 */
export async function fetchTierPricing(currency: string = 'usd'): Promise<TierConfigItem[]> {
  const validCurrencies = ['usd', 'inr', 'eur', 'gbp'] as const
  const curr = validCurrencies.includes(currency.toLowerCase() as typeof validCurrencies[number])
    ? (currency.toLowerCase() as typeof validCurrencies[number])
    : 'usd'
  return getTierConfig(curr)
}
