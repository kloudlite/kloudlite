import { createClient } from 'jsr:@supabase/supabase-js@2'

// --- Types ---

interface RenewalJobRow {
  id: string
  installation_id: string
  job_type: 'renewal' | 'expire'
  scheduled_at: string
  status: string
  attempts: number
  last_error: string | null
}

interface SubscriptionRow {
  id: string
  installation_id: string
  plan_id: string
  razorpay_subscription_id: string | null
  status: string
  quantity: number
  billing_period: 'monthly' | 'annual'
  current_start: string | null
  current_end: string | null
}

interface PlanRow {
  id: string
  name: string
  amount_per_user: number
  base_fee: number
  currency: string
  annual_discount_pct: number
}

// --- Supabase client ---

function getSupabase() {
  const url = Deno.env.get('SUPABASE_URL')
  const key = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')
  if (!url || !key) throw new Error('SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY must be set')
  return createClient(url, key, { auth: { persistSession: false } })
}

// --- Razorpay (raw fetch, no SDK) ---

async function createRazorpayOrder(params: {
  amount: number
  currency: string
  receipt: string
  notes: Record<string, string>
}): Promise<{ id: string }> {
  const keyId = Deno.env.get('RAZORPAY_KEY_ID')
  const keySecret = Deno.env.get('RAZORPAY_KEY_SECRET')
  if (!keyId || !keySecret) throw new Error('RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET must be set')

  const res = await fetch('https://api.razorpay.com/v1/orders', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: 'Basic ' + btoa(`${keyId}:${keySecret}`),
    },
    body: JSON.stringify(params),
  })

  if (!res.ok) {
    const body = await res.text()
    throw new Error(`Razorpay order creation failed (${res.status}): ${body}`)
  }

  return res.json()
}

// --- Retry with exponential backoff ---

async function withRetry<T>(
  fn: () => Promise<T>,
  label: string,
  maxAttempts = 3,
  baseDelayMs = 1000,
): Promise<T> {
  let lastError: unknown
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await fn()
    } catch (err) {
      lastError = err
      if (attempt < maxAttempts) {
        const delay = baseDelayMs * Math.pow(2, attempt - 1)
        console.warn(`[retry] ${label} — attempt ${attempt}/${maxAttempts} failed, retrying in ${delay}ms`)
        await new Promise((resolve) => setTimeout(resolve, delay))
      }
    }
  }
  throw lastError
}

// --- DB queries ---

const supabase = getSupabase()

async function getDueJobs(): Promise<RenewalJobRow[]> {
  const now = new Date().toISOString()
  const { data, error } = await supabase
    .from('renewal_jobs')
    .select('*')
    .eq('status', 'pending')
    .lte('scheduled_at', now)
    .order('scheduled_at', { ascending: true })

  if (error) {
    console.error('Failed to query due jobs:', error.message)
    return []
  }
  return data as RenewalJobRow[]
}

async function claimJob(jobId: string): Promise<boolean> {
  const { error, count } = await supabase
    .from('renewal_jobs')
    .update({ status: 'processing', attempts: 1 })
    .eq('id', jobId)
    .eq('status', 'pending')

  if (error) {
    console.error(`Failed to claim job ${jobId}:`, error.message)
    return false
  }
  return (count ?? 0) > 0
}

async function completeJob(jobId: string): Promise<void> {
  const { error } = await supabase
    .from('renewal_jobs')
    .update({ status: 'completed' })
    .eq('id', jobId)

  if (error) console.error(`Failed to complete job ${jobId}:`, error.message)
}

async function failJob(
  jobId: string,
  errorMessage: string,
  currentAttempts: number,
  maxAttempts: number,
): Promise<void> {
  const newStatus = currentAttempts < maxAttempts ? 'pending' : 'failed'
  const { error } = await supabase
    .from('renewal_jobs')
    .update({ status: newStatus, last_error: errorMessage, attempts: currentAttempts })
    .eq('id', jobId)

  if (error) console.error(`Failed to update job ${jobId} status:`, error.message)
}

async function getActiveSubscriptionsByInstallation(installationId: string): Promise<SubscriptionRow[]> {
  const { data, error } = await supabase
    .from('subscriptions')
    .select('*')
    .eq('installation_id', installationId)
    .eq('status', 'active')

  if (error) {
    console.error('Failed to query subscriptions:', error.message)
    return []
  }
  return data as SubscriptionRow[]
}

async function hasPendingInvoice(installationId: string): Promise<boolean> {
  const { count, error } = await supabase
    .from('invoices')
    .select('*', { count: 'exact', head: true })
    .eq('installation_id', installationId)
    .eq('status', 'issued')

  if (error) {
    console.error('Failed to check pending invoices:', error.message)
    return false
  }
  return (count ?? 0) > 0
}

async function getPlansByIds(planIds: string[]): Promise<PlanRow[]> {
  const { data, error } = await supabase
    .from('subscription_plans')
    .select('id, name, amount_per_user, base_fee, currency, annual_discount_pct')
    .in('id', planIds)

  if (error) {
    console.error('Failed to fetch plans:', error.message)
    return []
  }
  return data as PlanRow[]
}

async function insertInvoice(data: {
  subscriptionId: string
  installationId: string
  razorpayInvoiceId: string
  amount: number
  currency: string
  billingStart: string
  billingEnd: string
}): Promise<void> {
  const { error } = await supabase.from('invoices').insert({
    subscription_id: data.subscriptionId,
    installation_id: data.installationId,
    razorpay_invoice_id: data.razorpayInvoiceId,
    amount: data.amount,
    currency: data.currency,
    status: 'issued',
    billing_start: data.billingStart,
    billing_end: data.billingEnd,
  })

  if (error) throw new Error(`Failed to insert invoice: ${error.message}`)
}

async function expireSubscriptionsByInstallation(installationId: string): Promise<void> {
  const { error } = await supabase
    .from('subscriptions')
    .update({ status: 'expired' })
    .eq('installation_id', installationId)
    .eq('status', 'active')

  if (error) throw new Error(`Failed to expire subscriptions: ${error.message}`)
}

async function expireIssuedInvoicesByInstallation(installationId: string): Promise<void> {
  const { error } = await supabase
    .from('invoices')
    .update({ status: 'expired' })
    .eq('installation_id', installationId)
    .eq('status', 'issued')

  if (error) throw new Error(`Failed to expire invoices: ${error.message}`)
}

async function insertJobLog(data: {
  jobId: string | null
  jobType: string
  installationId: string
  status: 'success' | 'failed'
  details: string
}): Promise<void> {
  const { error } = await supabase.from('cron_job_logs').insert({
    job_id: data.jobId,
    job_type: data.jobType,
    installation_id: data.installationId,
    status: data.status,
    details: data.details,
  })

  if (error) console.error('Failed to insert job log:', error.message)
}

// --- Job processors ---

const MAX_ATTEMPTS = 3

async function processRenewalJob(job: RenewalJobRow): Promise<void> {
  const installationId = job.installation_id

  if (await hasPendingInvoice(installationId)) {
    console.log(`[renewal] Skipping ${installationId} — pending invoice exists`)
    return
  }

  const subs = await getActiveSubscriptionsByInstallation(installationId)
  if (subs.length === 0) {
    console.log(`[renewal] No active subscriptions for ${installationId}, skipping`)
    return
  }

  const planIds = [...new Set(subs.map((s) => s.plan_id))]
  const plans = await getPlansByIds(planIds)
  const planMap = new Map(plans.map((p) => [p.id, p]))

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

  const order = await withRetry(
    () =>
      createRazorpayOrder({
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
  )

  const billingStart = subs[0].current_end ?? new Date().toISOString()
  const periodDays = subs[0].billing_period === 'annual' ? 365 : 30
  const billingEnd = new Date(
    new Date(billingStart).getTime() + periodDays * 24 * 60 * 60 * 1000,
  ).toISOString()

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

  console.log(`[renewal] Created order ${order.id} for ${installationId} — ${totalAmount} ${currency}`)
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

async function processDueJobs(): Promise<{ processed: number; errors: number }> {
  const jobs = await getDueJobs()
  if (jobs.length === 0) return { processed: 0, errors: 0 }

  console.log(`[cron] Found ${jobs.length} due job(s)`)
  let processed = 0
  let errors = 0

  for (const job of jobs) {
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
        details: 'Job completed',
      })
      processed++
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
      errors++
    }
  }

  return { processed, errors }
}

// --- Edge Function handler ---

Deno.serve(async (req) => {
  // Only allow POST (from pg_cron via pg_net) or GET (for manual health checks)
  if (req.method !== 'POST' && req.method !== 'GET') {
    return new Response('Method not allowed', { status: 405 })
  }

  const start = Date.now()
  console.log(`[cron] Invoked at ${new Date().toISOString()}`)

  try {
    const result = await processDueJobs()
    const elapsed = Date.now() - start

    return new Response(
      JSON.stringify({ ok: true, ...result, elapsed_ms: elapsed }),
      { headers: { 'Content-Type': 'application/json' } },
    )
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    console.error('[cron] Fatal error:', message)
    return new Response(
      JSON.stringify({ ok: false, error: message }),
      { status: 500, headers: { 'Content-Type': 'application/json' } },
    )
  }
})
