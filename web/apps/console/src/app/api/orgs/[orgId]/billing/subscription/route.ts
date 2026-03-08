import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireOrgAccess, requireOrgOwner } from '@/lib/console/authorization'
import { getStripe } from '@/lib/stripe'
import {
  getBillingAccount,
  getSubscriptionItems,
  syncSubscriptionItemsFromStripe,
} from '@/lib/console/storage'

export const runtime = 'nodejs'

/**
 * GET /api/orgs/[orgId]/billing/subscription
 * Fetch billing account and subscription items for an org
 */
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgAccess(orgId)

    const [customer, items] = await Promise.all([
      getBillingAccount(orgId),
      getSubscriptionItems(orgId),
    ])

    return NextResponse.json({ customer, items })
  } catch (error) {
    return apiCatchError(error, 'Failed to fetch subscription')
  }
}

/**
 * PATCH /api/orgs/[orgId]/billing/subscription
 * Modify subscription items (change quantities, add/remove tiers)
 */
export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgOwner(orgId)

    const stripeCustomer = await getBillingAccount(orgId)
    if (!stripeCustomer?.stripeSubscriptionId) {
      return apiError('No active subscription found', 400)
    }

    const body = await request.json()
    const modifications: Array<{ priceId: string; quantity: number }> = body.modifications
    if (!modifications || !Array.isArray(modifications)) {
      return apiError('Invalid request: modifications array required', 400)
    }

    const stripe = getStripe()
    const subscription = await stripe.subscriptions.retrieve(stripeCustomer.stripeSubscriptionId)

    const items: Array<{
      id?: string
      price?: string
      quantity?: number
      deleted?: boolean
    }> = []

    for (const mod of modifications) {
      const existing = subscription.items.data.find((i) => i.price.id === mod.priceId)

      if (existing) {
        if (mod.quantity === 0) {
          items.push({ id: existing.id, deleted: true })
        } else {
          items.push({ id: existing.id, quantity: mod.quantity })
        }
      } else if (mod.quantity > 0) {
        items.push({ price: mod.priceId, quantity: mod.quantity })
      }
    }

    if (items.length === 0) {
      return apiError('No changes to apply', 400)
    }

    // Apply the subscription change immediately
    await stripe.subscriptions.update(stripeCustomer.stripeSubscriptionId, {
      items,
      proration_behavior: 'always_invoice',
    })

    // Sync updated items back to DB
    await syncSubscriptionItemsFromStripe(orgId, stripeCustomer.stripeSubscriptionId)

    return NextResponse.json({ success: true })
  } catch (error) {
    return apiCatchError(error, 'Failed to modify subscription')
  }
}

/**
 * DELETE /api/orgs/[orgId]/billing/subscription
 * Cancel subscription at end of billing period
 */
export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgOwner(orgId)

    const stripeCustomer = await getBillingAccount(orgId)
    if (!stripeCustomer?.stripeSubscriptionId) {
      return apiError('No active subscription found', 400)
    }

    const stripe = getStripe()
    await stripe.subscriptions.update(stripeCustomer.stripeSubscriptionId, {
      cancel_at_period_end: true,
    })

    return NextResponse.json({ success: true, cancelAtPeriodEnd: true })
  } catch (error) {
    return apiCatchError(error, 'Failed to cancel subscription')
  }
}
