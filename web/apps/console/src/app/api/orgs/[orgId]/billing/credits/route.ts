import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireOrgAccess, requireOrgOwner } from '@/lib/console/authorization'
import {
  getCreditAccount,
  getCreditTransactions,
  getActiveUsagePeriods,
  getPricingTiers,
  updateCreditAccountAutoTopup,
} from '@/lib/console/storage/credits'
import { getRegistrationSession } from '@/lib/console-auth'

export const runtime = 'nodejs'

/**
 * GET /api/orgs/[orgId]/billing/credits
 * Return credit account balance, recent transactions, active usage periods,
 * and current pricing tiers for the organization.
 */
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgAccess(orgId)

    const session = await getRegistrationSession()
    if (!session?.user) {
      return apiError('Not authenticated', 401)
    }

    const [account, transactions, activePeriods, pricingTiers] =
      await Promise.all([
        getCreditAccount(orgId),
        getCreditTransactions(orgId, 50),
        getActiveUsagePeriods(orgId),
        getPricingTiers(),
      ])

    return NextResponse.json({
      account,
      transactions,
      activePeriods,
      pricingTiers,
    })
  } catch (error) {
    return apiCatchError(error, 'Failed to fetch credit details')
  }
}

/**
 * PATCH /api/orgs/[orgId]/billing/credits
 * Update auto-topup settings for the organization's credit account.
 * Requires org owner role.
 */
export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgOwner(orgId)

    const session = await getRegistrationSession()
    if (!session?.user) {
      return apiError('Not authenticated', 401)
    }

    const body = await request.json()
    const { autoTopupEnabled, autoTopupThreshold, autoTopupAmount } = body as {
      autoTopupEnabled: boolean
      autoTopupThreshold?: number
      autoTopupAmount?: number
    }

    if (typeof autoTopupEnabled !== 'boolean') {
      return apiError('autoTopupEnabled must be a boolean', 400)
    }

    await updateCreditAccountAutoTopup(
      orgId,
      autoTopupEnabled,
      autoTopupThreshold,
      autoTopupAmount,
    )

    return NextResponse.json({ success: true })
  } catch (error) {
    return apiCatchError(error, 'Failed to update auto-topup settings')
  }
}
