import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireOrgOwner } from '@/lib/console/authorization'
import { getStripe } from '@/lib/stripe'
import { getBillingAccount, upsertBillingAccount } from '@/lib/console/storage'
import { getRegistrationSession } from '@/lib/console-auth'

export const runtime = 'nodejs'

/**
 * POST /api/orgs/[orgId]/billing/checkout
 * Create a Stripe Checkout session for a new subscription (org-scoped)
 */
export async function POST(
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
    const { allocations, installationId } = body as {
      allocations: Array<{ priceId: string; quantity: number }>
      installationId?: string
    }

    if (!allocations || !Array.isArray(allocations)) {
      return apiError('Invalid request: allocations array required', 400)
    }

    const stripe = getStripe()

    const stripeCustomer = await getBillingAccount(orgId)
    let customerId: string

    if (stripeCustomer?.stripeCustomerId) {
      customerId = stripeCustomer.stripeCustomerId
    } else {
      const customer = await stripe.customers.create({
        email: session.user.email,
        name: session.user.name ?? undefined,
        metadata: { org_id: orgId },
      })
      customerId = customer.id
      await upsertBillingAccount({
        orgId,
        stripeCustomerId: customerId,
        stripeSubscriptionId: null,
        billingStatus: 'incomplete',
        currentPeriodEnd: null,
      })
    }

    const lineItems = allocations
      .filter((a) => a.quantity > 0)
      .map((a) => ({ price: a.priceId, quantity: a.quantity }))

    // Build success/cancel URLs based on whether we have an installation context
    const appUrl = process.env.NEXT_PUBLIC_APP_URL
    let successUrl: string
    let cancelUrl: string

    if (installationId) {
      // Coming from installation creation flow — route back through /continue
      // which handles Stripe sync, session updates, and state-based routing
      successUrl = `${appUrl}/api/installations/${installationId}/continue`
      cancelUrl = `${appUrl}/installations/new-kl-cloud?installation=${installationId}&checkout=cancelled`
    } else {
      // Coming from billing settings — route back there
      successUrl = `${appUrl}/installations/settings/billing`
      cancelUrl = `${appUrl}/installations/settings/billing?checkout=cancelled`
    }

    const metadata: Record<string, string> = { org_id: orgId }
    if (installationId) {
      metadata.installation_id = installationId
    }

    const checkoutSession = await stripe.checkout.sessions.create({
      customer: customerId,
      mode: 'subscription',
      line_items: lineItems,
      success_url: successUrl,
      cancel_url: cancelUrl,
      metadata,
      subscription_data: {
        metadata,
      },
    })

    return NextResponse.json({ url: checkoutSession.url })
  } catch (error) {
    return apiCatchError(error, 'Failed to create checkout session')
  }
}
