import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireOrgOwner } from '@/lib/console/authorization'
import { getStripe } from '@/lib/stripe'
import {
  ensureCreditAccount,
  updateCreditAccountStripeCustomer,
} from '@/lib/console/storage/credits'
import { getRegistrationSession } from '@/lib/console-auth'

export const runtime = 'nodejs'

/**
 * POST /api/orgs/[orgId]/billing/topup
 * Create a Stripe Invoice for a one-time credit top-up.
 * Returns the hosted invoice URL for payment.
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
    const { amount, installationId } = body as {
      amount: number
      installationId?: string
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
    } else {
      const customer = await stripe.customers.create({
        metadata: { org_id: orgId },
      })
      stripeCustomerId = customer.id
      await updateCreditAccountStripeCustomer(orgId, stripeCustomerId)
    }

    // Create Stripe Invoice for the top-up
    const invoice = await stripe.invoices.create({
      customer: stripeCustomerId,
      collection_method: 'send_invoice',
      days_until_due: 0,
      auto_advance: true,
      metadata: {
        org_id: orgId,
        type: 'credit_topup',
        installation_id: installationId || '',
      },
    })

    await stripe.invoiceItems.create({
      customer: stripeCustomerId,
      invoice: invoice.id,
      amount: Math.round(amount * 100), // cents
      currency: 'usd',
      description: `Kloudlite Credit Top-Up: $${amount.toFixed(2)}`,
    })

    const finalizedInvoice = await stripe.invoices.finalizeInvoice(invoice.id)

    return NextResponse.json({ url: finalizedInvoice.hosted_invoice_url })
  } catch (error) {
    return apiCatchError(error, 'Failed to create top-up invoice')
  }
}
