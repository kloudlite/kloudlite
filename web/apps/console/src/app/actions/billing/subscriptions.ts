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
  /** Redirect URL for payment (hosted invoice page) */
  url?: string
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

  // Update subscription — payment_behavior: 'default_incomplete' so the invoice
  // isn't auto-charged. We'll redirect the user to pay via hosted invoice page.
  const updatedSubscription = await stripe.subscriptions.update(
    stripeCustomer.stripeSubscriptionId,
    {
      items,
      proration_behavior: 'always_invoice',
      payment_behavior: 'default_incomplete',
    },
  )

  // Sync updated items back to DB
  await syncSubscriptionItemsFromStripe(installationId, stripeCustomer.stripeSubscriptionId)

  // Get the latest invoice and redirect to its hosted payment page
  const invoiceId = typeof updatedSubscription.latest_invoice === 'string'
    ? updatedSubscription.latest_invoice
    : updatedSubscription.latest_invoice?.id

  if (invoiceId) {
    const invoice = await stripe.invoices.retrieve(invoiceId)

    // If the invoice is open (needs payment), redirect to the hosted page
    if (invoice.status === 'open' && invoice.hosted_invoice_url) {
      return { success: false, url: invoice.hosted_invoice_url }
    }

    // If amount due is 0 (e.g., downgrade with credit), it's already paid
    if (invoice.status === 'paid' || invoice.amount_due === 0) {
      return { success: true }
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
