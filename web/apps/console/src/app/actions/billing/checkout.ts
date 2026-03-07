'use server'

import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getStripe } from '@/lib/stripe'
import {
  getStripeCustomer,
  upsertStripeCustomer,
  getMemberRole,
  getInstallationById,
} from '@/lib/console/storage'

interface CheckoutAllocation {
  priceId: string
  quantity: number
}

export async function createCheckoutSession(
  installationId: string,
  allocations: CheckoutAllocation[],
) {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const installation = await getInstallationById(installationId)
  if (!installation || installation.userId !== session.user.id) {
    throw new Error('Forbidden: only installation owner can manage billing')
  }

  const stripe = getStripe()

  let stripeCustomer = await getStripeCustomer(installationId)
  let customerId: string

  if (stripeCustomer?.stripeCustomerId) {
    customerId = stripeCustomer.stripeCustomerId
  } else {
    const customer = await stripe.customers.create({
      email: session.user.email,
      name: session.user.name ?? undefined,
      metadata: { installation_id: installationId },
    })
    customerId = customer.id
    await upsertStripeCustomer({
      installationId,
      stripeCustomerId: customerId,
      stripeSubscriptionId: null,
      billingStatus: 'incomplete',
      currentPeriodEnd: null,
    })
  }

  const lineItems = allocations
    .filter((a) => a.quantity > 0)
    .map((a) => ({ price: a.priceId, quantity: a.quantity }))

  const checkoutSession = await stripe.checkout.sessions.create({
    customer: customerId,
    mode: 'subscription',
    line_items: lineItems,
    success_url: `${process.env.NEXT_PUBLIC_APP_URL}/installations/${installationId}?checkout=success`,
    cancel_url: `${process.env.NEXT_PUBLIC_APP_URL}/installations/${installationId}/billing?checkout=cancelled`,
    metadata: { installation_id: installationId },
    subscription_data: {
      metadata: { installation_id: installationId },
    },
  })

  return { url: checkoutSession.url }
}

export async function createPortalSession(installationId: string) {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const role = await getMemberRole(installationId, session.user.id)
  const installation = await getInstallationById(installationId)
  if (!role && installation?.userId !== session.user.id) {
    throw new Error('Forbidden')
  }

  const stripeCustomer = await getStripeCustomer(installationId)
  if (!stripeCustomer?.stripeCustomerId) {
    throw new Error('No billing account found. Please subscribe first.')
  }

  const stripe = getStripe()
  const portalSession = await stripe.billingPortal.sessions.create({
    customer: stripeCustomer.stripeCustomerId,
    return_url: `${process.env.NEXT_PUBLIC_APP_URL}/installations/${installationId}/billing`,
  })

  return { url: portalSession.url }
}
