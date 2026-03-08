import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { supabase } from '@/lib/console/supabase'
import { getStripe } from '@/lib/stripe'
import { getUsagePeriodsByDateRange } from '@/lib/console/storage/credits'

export const runtime = 'nodejs'

/**
 * POST /api/cron/monthly-invoice
 * Monthly cron job that generates informational Stripe invoices
 * summarizing the previous month's usage per org.
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
    // 1. Get all orgs with credit accounts that have a stripeCustomerId
    const { data: creditAccounts, error: fetchError } = await supabase
      .from('credit_accounts')
      .select('org_id, stripe_customer_id')
      .not('stripe_customer_id', 'is', null)

    if (fetchError || !creditAccounts || creditAccounts.length === 0) {
      return NextResponse.json({ success: true, invoicesCreated: 0 })
    }

    // 2. Calculate previous month date range
    const now = new Date()
    const previousMonth = new Date(now.getFullYear(), now.getMonth() - 1, 1)
    const startDate = previousMonth.toISOString()
    const endDate = new Date(now.getFullYear(), now.getMonth(), 1).toISOString()
    const previousMonthString = `${previousMonth.getFullYear()}-${String(previousMonth.getMonth() + 1).padStart(2, '0')}`

    const stripe = getStripe()
    let invoicesCreated = 0

    for (const account of creditAccounts) {
      const orgId = account.org_id
      const stripeCustomerId = account.stripe_customer_id as string

      // 2b. Get usage periods for that range
      const periods = await getUsagePeriodsByDateRange(orgId, startDate, endDate)

      // 2c. Group by resource_type, calculate totals
      const groups = new Map<string, { hours: number; rate: number; total: number }>()

      for (const period of periods) {
        const existing = groups.get(period.resourceType) || {
          hours: 0,
          rate: period.hourlyRate,
          total: 0,
        }
        const hours = period.endedAt
          ? (Date.parse(period.endedAt) - Date.parse(period.startedAt)) / (1000 * 60 * 60)
          : (Date.parse(endDate) - Date.parse(period.startedAt)) / (1000 * 60 * 60)
        existing.hours += hours
        existing.total += hours * period.hourlyRate
        groups.set(period.resourceType, existing)
      }

      // 2d. If no usage, skip this org
      if (groups.size === 0) continue

      // 2e. Create Stripe informational invoice
      const invoice = await stripe.invoices.create({
        customer: stripeCustomerId,
        collection_method: 'send_invoice',
        days_until_due: 0,
        auto_advance: false, // informational only
        metadata: {
          org_id: orgId,
          type: 'monthly_summary',
          period: previousMonthString,
        },
      })

      // Add line items per resource type
      for (const [resourceType, group] of groups) {
        await stripe.invoiceItems.create({
          customer: stripeCustomerId,
          invoice: invoice.id,
          amount: 0,
          currency: 'usd',
          description: `${resourceType}: ${group.hours.toFixed(1)} hours × $${group.rate}/hr = $${group.total.toFixed(2)}`,
        })
      }

      await stripe.invoices.finalizeInvoice(invoice.id)
      invoicesCreated++
      console.log(
        `[Monthly Invoice] Created summary invoice for org ${orgId} (${previousMonthString})`,
      )
    }

    return NextResponse.json({ success: true, invoicesCreated })
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error'
    console.error('Monthly invoice cron error:', message)
    return NextResponse.json({ error: 'Processing failed' }, { status: 500 })
  }
}
