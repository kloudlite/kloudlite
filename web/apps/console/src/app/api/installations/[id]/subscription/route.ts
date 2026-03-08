import { NextResponse } from 'next/server'
import { requireInstallationAccess } from '@/lib/console/authorization'
import {
  getBillingAccount,
  getSubscriptionItems,
  upsertBillingAccount,
  syncSubscriptionItemsFromStripe,
} from '@/lib/console/storage'
import { getStripe } from '@/lib/stripe'
import { apiCatchError } from '@/lib/api-helpers'

export const runtime = 'nodejs'

/**
 * GET /api/installations/[id]/subscription
 * Returns the billing account and subscription items for the org that owns
 * the given installation. Used by the deploy page to verify an active
 * subscription before triggering deployment.
 *
 * Includes a Stripe-direct fallback for the webhook race condition:
 * after Stripe Checkout the redirect may land before the webhook fires,
 * so we check Stripe directly if the DB still shows incomplete.
 */
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const { id } = await params

  try {
    const context = await requireInstallationAccess(id)
    const orgId = context.orgId
    let customer = await getBillingAccount(orgId)

    // Handle webhook race condition: if billing account is incomplete,
    // check Stripe directly for an active subscription
    if (
      customer?.stripeCustomerId &&
      customer.billingStatus !== 'active' &&
      customer.billingStatus !== 'cancelled'
    ) {
      try {
        const stripe = getStripe()
        const subs = await stripe.subscriptions.list({
          customer: customer.stripeCustomerId,
          status: 'active',
          limit: 1,
        })

        if (subs.data.length > 0) {
          const subscription = subs.data[0]
          const periodEnd = subscription.items.data[0]?.current_period_end

          await upsertBillingAccount({
            orgId,
            stripeCustomerId: customer.stripeCustomerId,
            stripeSubscriptionId: subscription.id,
            billingStatus: 'active',
            currentPeriodEnd: periodEnd
              ? new Date(periodEnd * 1000).toISOString()
              : null,
          })

          // Sync subscription items from Stripe
          await syncSubscriptionItemsFromStripe(orgId, subscription.id)

          // Re-fetch updated billing account
          customer = await getBillingAccount(orgId)
        }
      } catch (err) {
        console.error('[subscription] Failed to verify with Stripe:', err)
      }
    }

    // If we have a subscription but no items in DB, sync them
    let items = customer?.stripeSubscriptionId
      ? await getSubscriptionItems(orgId)
      : []

    if (items.length === 0 && customer?.stripeSubscriptionId) {
      await syncSubscriptionItemsFromStripe(orgId, customer.stripeSubscriptionId)
      items = await getSubscriptionItems(orgId)
    }

    return NextResponse.json({ customer, items })
  } catch (error) {
    return apiCatchError(error, 'Failed to get subscription status')
  }
}
