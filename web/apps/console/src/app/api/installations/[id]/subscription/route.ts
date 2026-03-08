import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import { getStripe } from '@/lib/stripe'
import {
  getStripeCustomer,
  getSubscriptionItems,
  syncSubscriptionItemsFromStripe,
  getMemberRole,
  getInstallationById,
} from '@/lib/console/storage'

export const runtime = 'nodejs'

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const session = await getRegistrationSession()
  if (!session?.user) {
    return apiError('Not authenticated', 401)
  }

  const { id } = await params

  // Verify user has access to this installation
  const role = await getMemberRole(id, session.user.id)
  const installation = await getInstallationById(id)
  if (!role && installation?.userId !== session.user.id) {
    return apiError('Forbidden', 403)
  }

  const [customer, items] = await Promise.all([
    getStripeCustomer(id),
    getSubscriptionItems(id),
  ])

  return NextResponse.json({ customer, items })
}

/**
 * PATCH /api/installations/[id]/subscription
 * Modify subscription items (change quantities, add/remove tiers)
 */
export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const session = await getRegistrationSession()
  if (!session?.user) {
    return apiError('Not authenticated', 401)
  }

  const { id } = await params

  const installation = await getInstallationById(id)
  if (!installation || installation.userId !== session.user.id) {
    return apiError('Forbidden: only installation owner can modify billing', 403)
  }

  const stripeCustomer = await getStripeCustomer(id)
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
  await syncSubscriptionItemsFromStripe(id, stripeCustomer.stripeSubscriptionId)

  return NextResponse.json({ success: true })
}
