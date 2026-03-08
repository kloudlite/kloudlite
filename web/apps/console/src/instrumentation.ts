/**
 * Next.js Instrumentation Hook
 *
 * Runs once when the Next.js server starts.
 * Pre-warms the pricing tier cache from the database.
 *
 * @see https://nextjs.org/docs/app/building-your-application/optimizing/instrumentation
 */
export async function register() {
  // Only run on the server (not during build or edge runtime)
  if (process.env.NEXT_RUNTIME === 'nodejs') {
    const { getActivePricingTiers } = await import('@/lib/stripe-bootstrap')
    await getActivePricingTiers()
  }
}
