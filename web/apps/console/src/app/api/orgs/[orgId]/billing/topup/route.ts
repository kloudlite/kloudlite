import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireOrgOwner } from '@/lib/console/authorization'
import { getStripe } from '@/lib/stripe'
import {
  ensureCreditAccount,
  updateCreditAccountStripeCustomer,
  topupCredits,
} from '@/lib/console/storage/credits'
import { getRegistrationSession } from '@/lib/console-auth'
import { isWebhookEventProcessed, markWebhookEventProcessed } from '@/lib/console/storage'

export const runtime = 'nodejs'

/**
 * GET /api/orgs/[orgId]/billing/topup?session_id=xxx
 * Verify a completed Stripe Checkout Session and credit the account.
 * Called after Stripe redirects back to the app.
 */
export async function GET(
  request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params
  const url = new URL(request.url)
  const sessionId = url.searchParams.get('session_id')

  if (!sessionId) {
    return apiError('Missing session_id', 400)
  }

  try {
    await requireOrgOwner(orgId)

    // Idempotency: skip if already processed
    const alreadyProcessed = await isWebhookEventProcessed(sessionId)
    if (alreadyProcessed) {
      return NextResponse.json({ status: 'already_processed' })
    }

    const stripe = getStripe()
    const session = await stripe.checkout.sessions.retrieve(sessionId)

    if (session.payment_status !== 'paid') {
      return apiError('Payment not completed', 400)
    }

    if (session.metadata?.type !== 'credit_topup' || session.metadata?.org_id !== orgId) {
      return apiError('Invalid session', 400)
    }

    const amount = (session.amount_total || 0) / 100
    if (amount > 0) {
      await topupCredits(orgId, amount, `Top-up via Stripe Checkout ${sessionId}`, sessionId)
      await markWebhookEventProcessed(sessionId, 'checkout.session.completed')
      console.log(`[Topup] Credited $${amount.toFixed(2)} to org ${orgId}`)
    }

    return NextResponse.json({ status: 'credited', amount })
  } catch (error) {
    return apiCatchError(error, 'Failed to verify payment')
  }
}

/**
 * POST /api/orgs/[orgId]/billing/topup
 * Create a Stripe Checkout Session for a one-time credit top-up.
 * Returns the checkout URL — Stripe redirects back after payment.
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
    const { amount, returnUrl } = body as {
      amount: number
      returnUrl?: string
    }

    if (typeof amount !== 'number' || amount < 5) {
      return apiError('Minimum top-up amount is $5', 400)
    }

    const stripe = getStripe()
    const creditAccount = await ensureCreditAccount(orgId)

    // Get or create Stripe customer
    let stripeCustomerId: string

    if (creditAccount.stripeCustomerId) {
      stripeCustomerId = creditAccount.stripeCustomerId
      const existing = await stripe.customers.retrieve(stripeCustomerId)
      if (!('deleted' in existing && existing.deleted) && !existing.email) {
        await stripe.customers.update(stripeCustomerId, {
          email: session.user.email,
          name: session.user.name,
        })
      }
    } else {
      const customer = await stripe.customers.create({
        email: session.user.email,
        name: session.user.name,
        metadata: { org_id: orgId },
      })
      stripeCustomerId = customer.id
      await updateCreditAccountStripeCustomer(orgId, stripeCustomerId)
    }

    const baseUrl = process.env.NEXT_PUBLIC_BASE_URL || 'http://localhost:3002'
    const rawReturnUrl = returnUrl || '/installations/settings/billing'

    // Build URLs with Stripe's template variable for session ID
    const successParsed = new URL(rawReturnUrl, baseUrl)
    successParsed.searchParams.set('payment', 'success')
    successParsed.searchParams.set('checkout_session', '{CHECKOUT_SESSION_ID}')
    const successUrl = successParsed.toString().replace('%7BCHECKOUT_SESSION_ID%7D', '{CHECKOUT_SESSION_ID}')

    const cancelParsed = new URL(rawReturnUrl, baseUrl)
    cancelParsed.searchParams.set('payment', 'cancelled')
    const cancelUrl = cancelParsed.toString()

    const checkoutSession = await stripe.checkout.sessions.create({
      customer: stripeCustomerId,
      mode: 'payment',
      line_items: [
        {
          price_data: {
            currency: 'usd',
            unit_amount: Math.round(amount * 100),
            product_data: {
              name: 'Kloudlite Credit Top-Up',
              description: `Add $${amount.toFixed(2)} credits to your account`,
            },
          },
          quantity: 1,
        },
      ],
      metadata: {
        org_id: orgId,
        type: 'credit_topup',
      },
      success_url: successUrl,
      cancel_url: cancelUrl,
    })

    return NextResponse.json({ url: checkoutSession.url })
  } catch (error) {
    return apiCatchError(error, 'Failed to create checkout session')
  }
}
