import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireOrgOwner } from '@/lib/console/authorization'
import { getStripe } from '@/lib/stripe'
import { getBillingAccount } from '@/lib/console/storage'

export const runtime = 'nodejs'

/**
 * POST /api/orgs/[orgId]/billing/portal
 * Create a Stripe Customer Portal session (org-scoped)
 */
export async function POST(
  _request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgOwner(orgId)

    const stripeCustomer = await getBillingAccount(orgId)
    if (!stripeCustomer?.stripeCustomerId) {
      return apiError('No billing account found. Please subscribe first.', 400)
    }

    const stripe = getStripe()
    const portalSession = await stripe.billingPortal.sessions.create({
      customer: stripeCustomer.stripeCustomerId,
      return_url: `${process.env.NEXT_PUBLIC_APP_URL}/installations/settings/billing`,
    })

    return NextResponse.json({ url: portalSession.url })
  } catch (error) {
    return apiCatchError(error, 'Failed to create portal session')
  }
}
