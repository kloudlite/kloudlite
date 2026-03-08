import { NextResponse } from 'next/server'
import { apiCatchError } from '@/lib/api-helpers'
import { getPricingTiers } from '@/lib/console/storage/credits'

export const runtime = 'nodejs'

/**
 * GET /api/pricing
 * Public endpoint — returns all active pricing tiers.
 * Cached for 5 minutes.
 */
export async function GET() {
  try {
    const tiers = await getPricingTiers()

    return NextResponse.json(
      { tiers },
      {
        headers: {
          'Cache-Control': 'public, max-age=300',
        },
      },
    )
  } catch (error) {
    return apiCatchError(error, 'Failed to fetch pricing tiers')
  }
}
