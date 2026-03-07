import Stripe from 'stripe'
import { getStripe } from './stripe'

/**
 * Stripe product/price bootstrap
 *
 * Idempotently creates products and prices in Stripe at app startup.
 * Uses product metadata to find existing resources — never creates duplicates.
 *
 * Products are identified by `metadata.kloudlite_tier`.
 * Prices are identified by `lookup_key` (e.g., "tier1_seat_usd").
 */

// --- Configuration ---

const CURRENCIES = ['usd', 'inr', 'eur', 'gbp'] as const
type Currency = (typeof CURRENCIES)[number]

// Base prices in USD cents — other currencies derived via exchange rates
const BASE_PRICES_USD: Record<number, number> = {
  0: 2900, // Control Plane: $29/mo
  1: 2900, // Tier 1: $29/user/mo
  2: 4900, // Tier 2: $49/user/mo
  3: 8900, // Tier 3: $89/user/mo
}

// Approximate exchange rates from USD (multiply USD amount)
const EXCHANGE_RATES: Record<Currency, number> = {
  usd: 1,
  inr: 83,
  eur: 0.92,
  gbp: 0.79,
}

interface ProductConfig {
  tier: number
  name: string
  description: string
  /** If true, quantity is always 1 (flat fee, not per-seat) */
  fixed: boolean
}

const PRODUCTS: ProductConfig[] = [
  { tier: 0, name: 'Control Plane', description: 'Dashboard, user management, billing', fixed: true },
  { tier: 1, name: 'Tier 1 — Light Workloads', description: '1 vCPU, 1 GB RAM, 5 GB storage', fixed: false },
  { tier: 2, name: 'Tier 2 — Standard Workloads', description: '2 vCPU, 4 GB RAM, 20 GB storage', fixed: false },
  { tier: 3, name: 'Tier 3 — Power Users', description: '4 vCPU, 8 GB RAM, 50 GB storage', fixed: false },
]

// --- Types ---

export interface TierPricing {
  tier: number
  name: string
  description: string
  fixed: boolean
  productId: string
  /** Price IDs keyed by currency (e.g., { usd: 'price_xxx', inr: 'price_yyy' }) */
  prices: Record<Currency, { priceId: string; unitAmount: number }>
}

// --- Module-level cache ---

let cachedPricing: TierPricing[] | null = null

/**
 * Get the bootstrapped pricing config. Returns cached result after first call.
 * Call `bootstrapStripeProducts()` first (done automatically via instrumentation).
 */
export function getStripePricing(): TierPricing[] {
  if (!cachedPricing) {
    throw new Error('Stripe products not bootstrapped yet. Ensure bootstrapStripeProducts() ran at startup.')
  }
  return cachedPricing
}

/**
 * Get pricing for a specific currency. Defaults to USD.
 */
export function getTierConfig(currency: Currency = 'usd') {
  const pricing = getStripePricing()
  return pricing.map((tier) => ({
    tier: tier.tier,
    name: tier.name,
    description: tier.description,
    fixed: tier.fixed,
    priceId: tier.prices[currency].priceId,
    pricePerUnit: tier.prices[currency].unitAmount,
  }))
}

// --- Bootstrap logic ---

function computeAmount(baseCentsUsd: number, currency: Currency): number {
  const rate = EXCHANGE_RATES[currency]
  // Round to nearest whole unit (cents/paise/etc.)
  return Math.round(baseCentsUsd * rate)
}

function lookupKey(tier: number, currency: Currency): string {
  const tierName = tier === 0 ? 'control_plane' : `tier${tier}_seat`
  return `${tierName}_${currency}`
}

async function findOrCreateProduct(
  stripe: Stripe,
  config: ProductConfig,
): Promise<string> {
  // Search for existing product by metadata
  const existing = await stripe.products.search({
    query: `metadata["kloudlite_tier"]:"${config.tier}"`,
    limit: 1,
  })

  if (existing.data.length > 0) {
    const product = existing.data[0]
    // Update if name/description changed
    if (product.name !== config.name || product.description !== config.description) {
      await stripe.products.update(product.id, {
        name: config.name,
        description: config.description,
      })
    }
    return product.id
  }

  // Create new product
  const product = await stripe.products.create({
    name: config.name,
    description: config.description,
    metadata: {
      kloudlite_tier: String(config.tier),
      tier: String(config.tier),
    },
  })

  return product.id
}

async function findOrCreatePrice(
  stripe: Stripe,
  productId: string,
  tier: number,
  currency: Currency,
): Promise<{ priceId: string; unitAmount: number }> {
  const key = lookupKey(tier, currency)
  const unitAmount = computeAmount(BASE_PRICES_USD[tier], currency)

  // Search for existing price by lookup_key
  const existing = await stripe.prices.list({
    lookup_keys: [key],
    limit: 1,
  })

  if (existing.data.length > 0) {
    const price = existing.data[0]
    // If amount changed, create a new price and transfer the lookup_key
    if (price.unit_amount !== unitAmount) {
      const newPrice = await stripe.prices.create({
        product: productId,
        unit_amount: unitAmount,
        currency,
        recurring: { interval: 'month' },
        lookup_key: key,
        transfer_lookup_key: true,
      })
      return { priceId: newPrice.id, unitAmount }
    }
    return { priceId: price.id, unitAmount: price.unit_amount ?? unitAmount }
  }

  // Create new price
  const price = await stripe.prices.create({
    product: productId,
    unit_amount: unitAmount,
    currency,
    recurring: { interval: 'month' },
    lookup_key: key,
    metadata: {
      kloudlite_tier: String(tier),
    },
  })

  return { priceId: price.id, unitAmount }
}

/**
 * Idempotently bootstrap Stripe products and prices.
 * Safe to call multiple times — finds existing resources by metadata/lookup_key.
 */
export async function bootstrapStripeProducts(): Promise<TierPricing[]> {
  if (cachedPricing) return cachedPricing

  if (!process.env.STRIPE_SECRET_KEY) {
    console.warn('[stripe-bootstrap] STRIPE_SECRET_KEY not set, skipping bootstrap')
    return []
  }

  const stripe = getStripe()
  const pricing: TierPricing[] = []

  console.log('[stripe-bootstrap] Ensuring Stripe products and prices exist...')

  for (const config of PRODUCTS) {
    const productId = await findOrCreateProduct(stripe, config)

    const prices: Record<string, { priceId: string; unitAmount: number }> = {}
    for (const currency of CURRENCIES) {
      prices[currency] = await findOrCreatePrice(stripe, productId, config.tier, currency)
    }

    pricing.push({
      tier: config.tier,
      name: config.name,
      description: config.description,
      fixed: config.fixed,
      productId,
      prices: prices as Record<Currency, { priceId: string; unitAmount: number }>,
    })

    console.log(`[stripe-bootstrap] ✓ ${config.name} (${productId})`)
  }

  console.log('[stripe-bootstrap] All products and prices ready.')
  cachedPricing = pricing
  return pricing
}
