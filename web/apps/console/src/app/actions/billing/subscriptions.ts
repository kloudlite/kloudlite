'use server'

import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getStripe } from '@/lib/stripe'
import { getStripeCustomer, getInstallationById, syncSubscriptionItemsFromStripe } from '@/lib/console/storage'

interface SubscriptionModification {
  priceId: string
  quantity: number
}

export interface ModifyResult {
  success: boolean
  /** If set, the client must confirm payment with this secret (3D Secure) */
  clientSecret?: string
  /** The payment intent status */
  paymentStatus?: string
}

export async function modifySubscription(
  installationId: string,
  modifications: SubscriptionModification[],
): Promise<ModifyResult> {
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

  // Use default_incomplete so we can handle 3DS if needed
  const updatedSubscription = await stripe.subscriptions.update(
    stripeCustomer.stripeSubscriptionId,
    {
      items,
      proration_behavior: 'always_invoice',
      payment_behavior: 'default_incomplete',
      expand: ['latest_invoice.payment_intent'],
    },
  )

  // Sync updated items back to DB
  await syncSubscriptionItemsFromStripe(installationId, stripeCustomer.stripeSubscriptionId)

  // Check if the latest invoice needs 3DS authentication
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const invoice = updatedSubscription.latest_invoice as any
  if (invoice && typeof invoice !== 'string') {
    const paymentIntent = invoice.payment_intent
    if (paymentIntent && typeof paymentIntent !== 'string') {
      if (paymentIntent.status === 'requires_action' || paymentIntent.status === 'requires_confirmation') {
        return {
          success: false,
          clientSecret: paymentIntent.client_secret ?? undefined,
          paymentStatus: paymentIntent.status,
        }
      }
    }
  }

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

  return { success: true }
}
