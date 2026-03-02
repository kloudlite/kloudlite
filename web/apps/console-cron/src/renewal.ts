import { getRazorpay } from './razorpay'
import { withRetry } from './retry'
import {
  getDueJobs,
  claimJob,
  completeJob,
  failJob,
  getActiveSubscriptionsByInstallation,
  hasPendingInvoice,
  getPlansByIds,
  insertInvoice,
  expireSubscriptionsByInstallation,
  expireIssuedInvoicesByInstallation,
  insertJobLog,
  type RenewalJobRow,
} from './queries'

const MAX_ATTEMPTS = 3

export async function processDueJobs(): Promise<void> {
  const jobs = await getDueJobs()
  if (jobs.length === 0) return

  console.log(`[cron] Found ${jobs.length} due job(s)`)

  for (const job of jobs) {
    // Claim the job (prevents duplicate processing if multiple cron instances)
    const claimed = await claimJob(job.id)
    if (!claimed) {
      console.log(`[cron] Job ${job.id} already claimed, skipping`)
      continue
    }

    try {
      if (job.job_type === 'renewal') {
        await processRenewalJob(job)
      } else if (job.job_type === 'expire') {
        await processExpireJob(job)
      }

      await completeJob(job.id)

      await insertJobLog({
        jobId: job.id,
        jobType: job.job_type,
        installationId: job.installation_id,
        status: 'success',
        details: `Job completed`,
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err)
      const attempts = job.attempts + 1
      console.error(`[cron] Job ${job.id} (${job.job_type}) failed:`, message)

      await failJob(job.id, message, attempts, MAX_ATTEMPTS)

      await insertJobLog({
        jobId: job.id,
        jobType: job.job_type,
        installationId: job.installation_id,
        status: 'failed',
        details: `Attempt ${attempts}/${MAX_ATTEMPTS}: ${message}`,
      }).catch(() => {})
    }
  }
}

async function processRenewalJob(job: RenewalJobRow): Promise<void> {
  const installationId = job.installation_id

  // Idempotency: skip if a pending invoice already exists
  if (await hasPendingInvoice(installationId)) {
    console.log(`[renewal] Skipping ${installationId} — pending invoice exists`)
    return
  }

  const subs = await getActiveSubscriptionsByInstallation(installationId)
  if (subs.length === 0) {
    console.log(`[renewal] No active subscriptions for ${installationId}, skipping`)
    return
  }

  // Fetch plans
  const planIds = [...new Set(subs.map((s) => s.plan_id))]
  const plans = await getPlansByIds(planIds)
  const planMap = new Map(plans.map((p) => [p.id, p]))

  // Compute total amount
  let baseFee = 0
  let currency = 'INR'
  let userTotal = 0

  for (const sub of subs) {
    const plan = planMap.get(sub.plan_id)
    if (!plan) {
      console.error(`[renewal] Plan not found: ${sub.plan_id}`)
      continue
    }
    baseFee = plan.base_fee
    currency = plan.currency
    userTotal += plan.amount_per_user * sub.quantity
  }

  const monthlyAmount = baseFee + userTotal
  const billingPeriod = subs[0].billing_period
  const discountPct = plans[0]?.annual_discount_pct ?? 20
  const totalAmount =
    billingPeriod === 'annual'
      ? Math.round(monthlyAmount * 12 * (1 - discountPct / 100))
      : monthlyAmount

  // Create Razorpay order with retry
  const razorpay = getRazorpay()
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const order = (await withRetry(
    () =>
      razorpay.orders.create({
        amount: totalAmount,
        currency,
        receipt: `renewal_${installationId.slice(0, 8)}_${Date.now()}`,
        notes: {
          installation_id: installationId,
          type: 'renewal',
          job_id: job.id,
        },
      }),
    `razorpay.orders.create(${installationId})`,
  )) as any

  // Compute next billing period (30 days for monthly, 365 for annual)
  const billingStart = subs[0].current_end ?? new Date().toISOString()
  const periodDays = subs[0].billing_period === 'annual' ? 365 : 30
  const billingEnd = new Date(
    new Date(billingStart).getTime() + periodDays * 24 * 60 * 60 * 1000,
  ).toISOString()

  // Insert invoice
  await withRetry(
    () =>
      insertInvoice({
        subscriptionId: subs[0].id,
        installationId,
        razorpayInvoiceId: order.id,
        amount: totalAmount,
        currency,
        billingStart,
        billingEnd,
      }),
    `insertInvoice(${installationId})`,
  )

  console.log(
    `[renewal] Created order ${order.id} for ${installationId} — ${totalAmount} ${currency}`,
  )
}

async function processExpireJob(job: RenewalJobRow): Promise<void> {
  const installationId = job.installation_id

  const subs = await getActiveSubscriptionsByInstallation(installationId)
  if (subs.length === 0) {
    console.log(`[expire] No active subscriptions for ${installationId}, skipping`)
    return
  }

  await withRetry(
    () => expireSubscriptionsByInstallation(installationId),
    `expireSubscriptions(${installationId})`,
  )
  await withRetry(
    () => expireIssuedInvoicesByInstallation(installationId),
    `expireInvoices(${installationId})`,
  )

  console.log(`[expire] Expired subscriptions for ${installationId}`)
}
