'use server'

import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getStripe } from '@/lib/stripe'
import { getStripeCustomer, getInstallationById, syncSubscriptionItemsFromStripe, updateBillingStatus } from '@/lib/console/storage'

interface SubscriptionModification {
  priceId: string
  quantity: number
}

export async function modifySubscription(
  installationId: string,
  modifications: SubscriptionModification[],
): Promise<{ success: boolean }> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const installation = await getInstallationById(installationId)
  if (!installation || installation.userId !== session.user.id) {
    throw new Error('Forbidden: only installation owner can modify billing')
  }

  const stripeCustomer = await getStripeCustomer(installationId)
  if (!stripeCustomer?.stripeSubscriptionId) {
    throw new Error('No active subscription found')
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
    throw new Error('No changes to apply')
  }

  // Apply the subscription change immediately.
  // Stripe automatically handles proration — any amount owed or credited
  // is applied to the next invoice.
  await stripe.subscriptions.update(stripeCustomer.stripeSubscriptionId, {
    items,
    proration_behavior: 'always_invoice',
  })

  // Sync updated items back to DB
  await syncSubscriptionItemsFromStripe(installationId, stripeCustomer.stripeSubscriptionId)

  return { success: true }
}

export async function cancelSubscription(installationId: string) {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const installation = await getInstallationById(installationId)
  if (!installation || installation.userId !== session.user.id) {
    throw new Error('Forbidden: only installation owner can cancel billing')
  }

  const stripeCustomer = await getStripeCustomer(installationId)
  if (!stripeCustomer?.stripeSubscriptionId) {
    throw new Error('No active subscription found')
  }

  const stripe = getStripe()
  await stripe.subscriptions.update(stripeCustomer.stripeSubscriptionId, {
    cancel_at_period_end: true,
  })

  // Update local DB to reflect cancellation
  await updateBillingStatus(stripeCustomer.stripeCustomerId, 'cancelled')

  return { success: true }
}
