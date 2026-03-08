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
  /** Client secret for stripe.confirmPayment() — always present when payment is needed */
  clientSecret?: string
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

  // Update subscription with default_incomplete — creates invoice but doesn't auto-charge
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

  // Check the latest invoice for payment status
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const invoice = updatedSubscription.latest_invoice as any
  if (invoice && typeof invoice !== 'string') {
    // No payment needed (downgrade with credit, $0 invoice)
    if (invoice.status === 'paid' || invoice.amount_due === 0) {
      return { success: true }
    }

    // Extract the PaymentIntent client_secret for client-side confirmation
    const paymentIntent = invoice.payment_intent
    if (paymentIntent && typeof paymentIntent !== 'string' && paymentIntent.client_secret) {
      return {
        success: false,
        clientSecret: paymentIntent.client_secret,
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
