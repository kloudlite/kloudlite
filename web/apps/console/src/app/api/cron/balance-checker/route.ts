import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { supabase } from '@/lib/console/supabase'
import { getStripe } from '@/lib/stripe'
import {
  getActiveUsagePeriods,
  debitCredits,
  updateLastBilledAt,
  getCreditAccount,
  updateCreditAccountWarnings,
  updateCreditAccountStripeCustomer,
} from '@/lib/console/storage/credits'

export const runtime = 'nodejs'

/**
 * POST /api/cron/balance-checker
 * Periodic cron job that debits accrued usage costs from org credit balances,
 * checks balance thresholds, and triggers auto top-ups when configured.
 */
export async function POST(request: NextRequest) {
  // Auth: x-cron-secret must match CRON_SECRET (skip auth in dev mode if not set)
  const cronSecret = process.env.CRON_SECRET
  if (cronSecret) {
    const headerSecret = request.headers.get('x-cron-secret')
    if (headerSecret !== cronSecret) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }
  }

  try {
    // 1. Query all orgs with active usage periods
    const { data: activeOrgs } = await supabase
      .from('usage_periods')
      .select('org_id')
      .is('ended_at', null)

    if (!activeOrgs || activeOrgs.length === 0) {
      return NextResponse.json({ success: true, orgsProcessed: 0 })
    }

    // Deduplicate org_ids
    const orgIds = [...new Set(activeOrgs.map((row) => row.org_id))]
    let orgsProcessed = 0

    for (const orgId of orgIds) {
      // 2a. Get all active usage periods for this org
      const activePeriods = await getActiveUsagePeriods(orgId)
      if (activePeriods.length === 0) continue

      // 2b. Calculate cost accrued since last_billed_at for each period
      const now = Date.now()
      let totalDebit = 0
      const periodIds: string[] = []

      for (const period of activePeriods) {
        const hoursSinceLastBilled =
          (now - Date.parse(period.lastBilledAt)) / (1000 * 60 * 60)
        if (hoursSinceLastBilled > 0) {
          totalDebit += hoursSinceLastBilled * period.hourlyRate
          periodIds.push(period.id)
        }
      }

      // 2c. Debit and update billing timestamps
      if (totalDebit > 0) {
        await debitCredits(orgId, totalDebit, 'Periodic usage debit')
        await updateLastBilledAt(periodIds)
      }

      // 2d. Get updated credit account
      const creditAccount = await getCreditAccount(orgId)
      if (!creditAccount) continue

      // 2e. Calculate burn rate (sum of all active period hourly rates)
      const burnRate = activePeriods.reduce(
        (sum, period) => sum + period.hourlyRate,
        0,
      )

      // 2f. Check balance thresholds
      if (creditAccount.balance > 0 && burnRate > 0 && creditAccount.balance / burnRate < 24) {
        await updateCreditAccountWarnings(orgId, true)
      }

      if (creditAccount.balance <= 0) {
        console.log(`Credit exhaustion for org ${orgId}`)
        // Actual WorkMachine pausing is future work — just flag it for now
      }

      if (creditAccount.balance < -10) {
        await updateCreditAccountWarnings(orgId, true, true)
      }

      // 2g. Auto top-up if enabled and balance below threshold
      if (
        creditAccount.autoTopupEnabled &&
        creditAccount.autoTopupThreshold !== null &&
        creditAccount.autoTopupAmount !== null &&
        creditAccount.balance < creditAccount.autoTopupThreshold
      ) {
        const stripe = getStripe()

        // Ensure Stripe customer exists
        let stripeCustomerId = creditAccount.stripeCustomerId
        if (!stripeCustomerId) {
          const customer = await stripe.customers.create({
            metadata: { org_id: orgId },
          })
          stripeCustomerId = customer.id
          await updateCreditAccountStripeCustomer(orgId, stripeCustomerId)
        }

        // Create Stripe invoice for auto top-up
        const invoice = await stripe.invoices.create({
          customer: stripeCustomerId,
          collection_method: 'send_invoice',
          days_until_due: 0,
          auto_advance: true,
          metadata: {
            org_id: orgId,
            type: 'credit_topup',
          },
        })

        await stripe.invoiceItems.create({
          customer: stripeCustomerId,
          invoice: invoice.id,
          amount: Math.round(creditAccount.autoTopupAmount * 100), // cents
          currency: 'usd',
          description: `Kloudlite Auto Top-Up: $${creditAccount.autoTopupAmount.toFixed(2)}`,
        })

        await stripe.invoices.finalizeInvoice(invoice.id)
        console.log(
          `[Balance Checker] Auto top-up invoice created for org ${orgId}: $${creditAccount.autoTopupAmount.toFixed(2)}`,
        )
      }

      orgsProcessed++
    }

    return NextResponse.json({ success: true, orgsProcessed })
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error'
    console.error('Balance checker cron error:', message)
    return NextResponse.json({ error: 'Processing failed' }, { status: 500 })
  }
}
