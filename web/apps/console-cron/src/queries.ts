import { supabase } from './db'

export interface RenewalJobRow {
  id: string
  installation_id: string
  job_type: 'renewal' | 'expire'
  scheduled_at: string
  status: string
  attempts: number
  last_error: string | null
}

export interface SubscriptionRow {
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

export interface PlanRow {
  id: string
  name: string
  amount_per_user: number
  base_fee: number
  currency: string
  annual_discount_pct: number
}

// --- Job queue queries ---

export async function getDueJobs(): Promise<RenewalJobRow[]> {
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

export async function claimJob(jobId: string): Promise<boolean> {
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

export async function completeJob(jobId: string): Promise<void> {
  const { error } = await supabase
    .from('renewal_jobs')
    .update({ status: 'completed' })
    .eq('id', jobId)

  if (error) {
    console.error(`Failed to complete job ${jobId}:`, error.message)
  }
}

export async function failJob(
  jobId: string,
  errorMessage: string,
  currentAttempts: number,
  maxAttempts: number,
): Promise<void> {
  // If under max attempts, return to pending for retry
  const newStatus = currentAttempts < maxAttempts ? 'pending' : 'failed'

  const { error } = await supabase
    .from('renewal_jobs')
    .update({
      status: newStatus,
      last_error: errorMessage,
      attempts: currentAttempts,
    })
    .eq('id', jobId)

  if (error) {
    console.error(`Failed to update job ${jobId} status:`, error.message)
  }
}

// --- Subscription/plan queries ---

export async function getActiveSubscriptionsByInstallation(
  installationId: string,
): Promise<SubscriptionRow[]> {
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

export async function hasPendingInvoice(installationId: string): Promise<boolean> {
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

export async function getPlansByIds(planIds: string[]): Promise<PlanRow[]> {
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

export async function insertInvoice(data: {
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

  if (error) {
    throw new Error(`Failed to insert invoice: ${error.message}`)
  }
}

export async function expireSubscriptionsByInstallation(
  installationId: string,
): Promise<void> {
  const { error } = await supabase
    .from('subscriptions')
    .update({ status: 'expired' })
    .eq('installation_id', installationId)
    .eq('status', 'active')

  if (error) {
    throw new Error(`Failed to expire subscriptions: ${error.message}`)
  }
}

export async function expireIssuedInvoicesByInstallation(
  installationId: string,
): Promise<void> {
  const { error } = await supabase
    .from('invoices')
    .update({ status: 'expired' })
    .eq('installation_id', installationId)
    .eq('status', 'issued')

  if (error) {
    throw new Error(`Failed to expire invoices: ${error.message}`)
  }
}

// --- Job logs ---

export async function insertJobLog(data: {
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

  if (error) {
    console.error('Failed to insert job log:', error.message)
  }
}
