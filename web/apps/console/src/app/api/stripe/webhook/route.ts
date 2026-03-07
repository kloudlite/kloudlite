import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'
import Stripe from 'stripe'
import { getStripe } from '@/lib/stripe'
import {
  getStripeCustomerByCustomerId,
  isWebhookEventProcessed,
  markWebhookEventProcessed,
  syncSubscriptionItems,
  updateBillingStatus,
  upsertStripeCustomer,
} from '@/lib/console/storage'
import type { StripeCustomer } from '@/lib/console/storage'

function mapStripeStatus(
  status: string,
  cancelAtPeriodEnd: boolean,
): StripeCustomer['billingStatus'] {
  if (cancelAtPeriodEnd) return 'cancelled'
  switch (status) {
    case 'active':
      return 'active'
    case 'past_due':
      return 'past_due'
    case 'canceled':
      return 'cancelled'
    case 'trialing':
      return 'trialing'
    default:
      return 'incomplete'
  }
}

function extractSubscriptionItems(subscription: Stripe.Subscription) {
  return subscription.items.data.map((item) => {
    const product = item.price.product as Stripe.Product
    const tier = product.metadata?.tier ? parseInt(product.metadata.tier, 10) : 0
    return {
      stripeSubscriptionItemId: item.id,
      stripePriceId: item.price.id,
      tier,
      productName: product.name,
      quantity: item.quantity ?? 1,
    }
  })
}

async function handleCheckoutSessionCompleted(
  stripe: Stripe,
  session: Stripe.Checkout.Session,
) {
  const installationId = session.metadata?.installation_id
  if (!installationId) {
    console.error('checkout.session.completed: missing installation_id in metadata')
    return
  }

  const subscriptionId = session.subscription as string | null
  if (!subscriptionId) {
    console.error('checkout.session.completed: missing subscription ID')
    return
  }

  const customerId = session.customer as string | null
  if (!customerId) {
    console.error('checkout.session.completed: missing customer ID')
    return
  }

  // Retrieve full subscription with expanded product data
  const subscription = await stripe.subscriptions.retrieve(subscriptionId, {
    expand: ['items.data.price.product'],
  })

  // Derive current_period_end from the first subscription item
  const firstItem = subscription.items.data[0]
  const periodEnd = firstItem?.current_period_end ?? null

  await upsertStripeCustomer({
    installationId,
    stripeCustomerId: customerId,
    stripeSubscriptionId: subscriptionId,
    billingStatus: mapStripeStatus(subscription.status, subscription.cancel_at_period_end),
    currentPeriodEnd: periodEnd
      ? new Date(periodEnd * 1000).toISOString()
      : null,
  })

  const items = extractSubscriptionItems(subscription)
  await syncSubscriptionItems(installationId, items)
}

async function handleSubscriptionUpdated(
  stripe: Stripe,
  subscription: Stripe.Subscription,
) {
  const stripeCustomerId = subscription.customer as string
  const customer = await getStripeCustomerByCustomerId(stripeCustomerId)
  if (!customer) {
    console.error(
      'customer.subscription.updated: no matching customer for stripe_customer_id',
      stripeCustomerId,
    )
    return
  }

  const billingStatus = mapStripeStatus(
    subscription.status,
    subscription.cancel_at_period_end,
  )
  // Derive current_period_end from the first subscription item
  const firstItemPeriodEnd = subscription.items.data[0]?.current_period_end ?? null
  const currentPeriodEnd = firstItemPeriodEnd
    ? new Date(firstItemPeriodEnd * 1000).toISOString()
    : null

  await updateBillingStatus(stripeCustomerId, billingStatus, currentPeriodEnd)

  // Retrieve subscription with expanded product data for item sync
  const fullSubscription = await stripe.subscriptions.retrieve(subscription.id, {
    expand: ['items.data.price.product'],
  })

  const items = extractSubscriptionItems(fullSubscription)
  await syncSubscriptionItems(customer.installationId, items)
}

async function handleSubscriptionDeleted(subscription: Stripe.Subscription) {
  const stripeCustomerId = subscription.customer as string
  const customer = await getStripeCustomerByCustomerId(stripeCustomerId)
  if (!customer) {
    console.error(
      'customer.subscription.deleted: no matching customer for stripe_customer_id',
      stripeCustomerId,
    )
    return
  }

  await updateBillingStatus(stripeCustomerId, 'cancelled')
  await syncSubscriptionItems(customer.installationId, [])
}

async function handleInvoicePaymentFailed(invoice: Stripe.Invoice) {
  const stripeCustomerId = invoice.customer as string | null
  if (!stripeCustomerId) {
    console.error('invoice.payment_failed: missing customer ID')
    return
  }

  const customer = await getStripeCustomerByCustomerId(stripeCustomerId)
  if (!customer) {
    console.error(
      'invoice.payment_failed: no matching customer for stripe_customer_id',
      stripeCustomerId,
    )
    return
  }

  await updateBillingStatus(
    stripeCustomerId,
    customer.billingStatus,
    undefined,
    true,
  )
}

export async function POST(request: NextRequest) {
  const webhookSecret = process.env.STRIPE_WEBHOOK_SECRET
  if (!webhookSecret) {
    console.error('STRIPE_WEBHOOK_SECRET not configured')
    return NextResponse.json({ error: 'Webhook secret not configured' }, { status: 500 })
  }

  const body = await request.text()
  const signature = request.headers.get('stripe-signature')

  if (!signature) {
    return NextResponse.json({ error: 'Missing stripe-signature header' }, { status: 401 })
  }

  let event: Stripe.Event
  try {
    const stripe = getStripe()
    event = stripe.webhooks.constructEvent(body, signature, webhookSecret)
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Unknown verification error'
    console.error('Webhook signature verification failed:', message)
    return NextResponse.json({ error: 'Invalid signature' }, { status: 401 })
  }

  // Idempotency: skip already-processed events
  const alreadyProcessed = await isWebhookEventProcessed(event.id)
  if (alreadyProcessed) {
    return NextResponse.json({ status: 'ok', skipped: 'duplicate' })
  }

  try {
    const stripe = getStripe()

    switch (event.type) {
      case 'checkout.session.completed': {
        const session = event.data.object as Stripe.Checkout.Session
        await handleCheckoutSessionCompleted(stripe, session)
        break
      }
      case 'customer.subscription.updated': {
        const subscription = event.data.object as Stripe.Subscription
        await handleSubscriptionUpdated(stripe, subscription)
        break
      }
      case 'customer.subscription.deleted': {
        const subscription = event.data.object as Stripe.Subscription
        await handleSubscriptionDeleted(subscription)
        break
      }
      case 'invoice.payment_failed': {
        const invoice = event.data.object as Stripe.Invoice
        await handleInvoicePaymentFailed(invoice)
        break
      }
      default:
        console.log('Unhandled Stripe webhook event:', event.type)
    }

    await markWebhookEventProcessed(event.id, event.type)
    return NextResponse.json({ status: 'ok' })
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error'
    console.error('Stripe webhook processing error:', { eventType: event.type, error: message })
    return NextResponse.json({ error: 'Processing failed' }, { status: 500 })
  }
}
