import { supabase } from '../supabase'
import type { Database } from '../supabase-types'
import type { Plan, PlanRow, Subscription, SubscriptionRow, Invoice, InvoiceRow } from './billing-types'

function mapToPlan(row: PlanRow): Plan {
  return {
    id: row.id,
    razorpayPlanId: row.razorpay_plan_id,
    tier: row.tier,
    name: row.name,
    amountPerUser: row.amount_per_user,
    baseFee: row.base_fee,
    currency: row.currency,
    monthlyHours: row.monthly_hours,
    overageRate: row.overage_rate,
    cpu: row.cpu,
    ram: row.ram,
    storage: row.storage,
    autoSuspend: row.auto_suspend,
    description: row.description,
    annualDiscountPct: row.annual_discount_pct ?? 20,
    createdAt: row.created_at,
  }
}

function mapToSubscription(row: SubscriptionRow): Subscription {
  return {
    id: row.id,
    installationId: row.installation_id,
    planId: row.plan_id,
    razorpaySubscriptionId: row.razorpay_subscription_id,
    razorpayCustomerId: row.razorpay_customer_id,
    status: row.status,
    quantity: row.quantity,
    billingPeriod: row.billing_period ?? 'monthly',
    currentStart: row.current_start,
    currentEnd: row.current_end,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  }
}

function mapToInvoice(row: InvoiceRow): Invoice {
  return {
    id: row.id,
    subscriptionId: row.subscription_id,
    installationId: row.installation_id,
    razorpayInvoiceId: row.razorpay_invoice_id,
    razorpayPaymentId: row.razorpay_payment_id,
    amount: row.amount,
    currency: row.currency,
    status: row.status,
    billingStart: row.billing_start,
    billingEnd: row.billing_end,
    paidAt: row.paid_at,
    createdAt: row.created_at,
  }
}

export async function getPlans(): Promise<Plan[]> {
  const { data, error } = await supabase
    .from('subscription_plans')
    .select('*')
    .order('tier', { ascending: true })
  if (error) {
    return []
  }
  return (data as PlanRow[]).map(mapToPlan)
}

export async function getPlanById(planId: string): Promise<Plan | null> {
  const { data, error } = await supabase
    .from('subscription_plans')
    .select('*')
    .eq('id', planId)
    .single()
  if (error) return null
  return mapToPlan(data as PlanRow)
}

export async function updatePlanRazorpayId(planId: string, razorpayPlanId: string): Promise<void> {
  const { error } = await supabase
    .from('subscription_plans')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update({ razorpay_plan_id: razorpayPlanId })
    .eq('id', planId)
  if (error) {
    throw new Error(`Failed to update plan razorpay_plan_id: ${error.message}`)
  }
}

export async function getSubscriptionByInstallation(installationId: string): Promise<Subscription | null> {
  const subs = await getSubscriptionsByInstallation(installationId)
  return subs[0] || null
}

export async function getSubscriptionsByInstallation(installationId: string): Promise<Subscription[]> {
  const { data, error } = await supabase
    .from('subscriptions')
    .select('*')
    .eq('installation_id', installationId)
    .order('created_at', { ascending: false })
  if (error) return []
  return (data as SubscriptionRow[]).map(mapToSubscription)
}

export async function getSubscriptionByRazorpayId(razorpaySubscriptionId: string): Promise<Subscription | null> {
  const { data, error } = await supabase
    .from('subscriptions')
    .select('*')
    .eq('razorpay_subscription_id', razorpaySubscriptionId)
    .single()
  if (error) return null
  return mapToSubscription(data as SubscriptionRow)
}

export async function createSubscription(data: {
  installationId: string
  planId: string
  razorpaySubscriptionId: string | null
  razorpayCustomerId: string | null
  quantity: number
  billingPeriod?: 'monthly' | 'annual'
}): Promise<Subscription> {
  type Insert = Database['public']['Tables']['subscriptions']['Insert']
  const insertData: Insert = {
    installation_id: data.installationId,
    plan_id: data.planId,
    razorpay_subscription_id: data.razorpaySubscriptionId,
    razorpay_customer_id: data.razorpayCustomerId,
    quantity: data.quantity,
    status: 'created',
    billing_period: data.billingPeriod ?? 'monthly',
  }
  const result = await supabase
    .from('subscriptions')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .insert(insertData)
    .select()
    .single()
  if (result.error) {
    throw new Error(`Failed to create subscription: ${result.error.message}`)
  }
  return mapToSubscription(result.data as SubscriptionRow)
}

export async function updateSubscriptionStatus(
  razorpaySubscriptionId: string,
  status: Subscription['status'],
  periodStart?: string,
  periodEnd?: string,
): Promise<void> {
  type Update = Database['public']['Tables']['subscriptions']['Update']
  const updateData: Update = { status }
  if (periodStart) updateData.current_start = periodStart
  if (periodEnd) updateData.current_end = periodEnd
  const { error } = await supabase
    .from('subscriptions')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update(updateData)
    .eq('razorpay_subscription_id', razorpaySubscriptionId)
  if (error) {
    throw new Error(`Failed to update subscription: ${error.message}`)
  }
}

export async function activateSubscriptionsByInstallation(
  installationId: string,
  periodStart: string,
  periodEnd: string,
): Promise<void> {
  type Update = Database['public']['Tables']['subscriptions']['Update']
  const updateData: Update = {
    status: 'active',
    current_start: periodStart,
    current_end: periodEnd,
  }
  const { error } = await supabase
    .from('subscriptions')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update(updateData)
    .eq('installation_id', installationId)
    .eq('status', 'created')
  if (error) {
    throw new Error(`Failed to activate subscriptions: ${error.message}`)
  }
}

export async function cancelSubscriptionsByInstallation(
  installationId: string,
): Promise<void> {
  type Update = Database['public']['Tables']['subscriptions']['Update']
  const updateData: Update = { status: 'cancelled' }
  const { error } = await supabase
    .from('subscriptions')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update(updateData)
    .eq('installation_id', installationId)
    .in('status', ['created', 'authenticated', 'active', 'paused'])
  if (error) {
    throw new Error(`Failed to cancel subscriptions: ${error.message}`)
  }
}

export async function getActiveSubscriptionsByInstallationIds(
  installationIds: string[],
): Promise<Record<string, Subscription>> {
  if (installationIds.length === 0) return {}

  const { data, error } = await supabase
    .from('subscriptions')
    .select('*')
    .in('installation_id', installationIds)
    .eq('status', 'active')

  if (error) return {}

  const result: Record<string, Subscription> = {}
  for (const row of data as SubscriptionRow[]) {
    // Keep only the first active subscription per installation
    if (!result[row.installation_id]) {
      result[row.installation_id] = mapToSubscription(row)
    }
  }
  return result
}

export async function getInvoicesByInstallation(installationId: string): Promise<Invoice[]> {
  const { data, error } = await supabase
    .from('invoices')
    .select('*')
    .eq('installation_id', installationId)
    .order('created_at', { ascending: false })
  if (error) return []
  return (data as InvoiceRow[]).map(mapToInvoice)
}

export async function getPendingInvoicesByInstallationIds(
  installationIds: string[],
): Promise<Record<string, Invoice>> {
  if (installationIds.length === 0) return {}

  const { data, error } = await supabase
    .from('invoices')
    .select('*')
    .in('installation_id', installationIds)
    .eq('status', 'issued')

  if (error) return {}

  const result: Record<string, Invoice> = {}
  for (const row of data as InvoiceRow[]) {
    // Keep only the first (most recent) pending invoice per installation
    if (!result[row.installation_id]) {
      result[row.installation_id] = mapToInvoice(row)
    }
  }
  return result
}

export async function getPendingInvoiceByInstallation(
  installationId: string,
): Promise<Invoice | null> {
  const { data, error } = await supabase
    .from('invoices')
    .select('*')
    .eq('installation_id', installationId)
    .eq('status', 'issued')
    .order('created_at', { ascending: false })
    .limit(1)
    .maybeSingle()

  if (error || !data) return null
  return mapToInvoice(data as InvoiceRow)
}

export async function extendSubscriptionPeriod(
  installationId: string,
  newStart: string,
  newEnd: string,
): Promise<void> {
  type Update = Database['public']['Tables']['subscriptions']['Update']
  const updateData: Update = {
    current_start: newStart,
    current_end: newEnd,
  }
  const { error } = await supabase
    .from('subscriptions')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update(updateData)
    .eq('installation_id', installationId)
    .eq('status', 'active')
  if (error) {
    throw new Error(`Failed to extend subscription period: ${error.message}`)
  }
}

export async function updateInvoiceStatus(
  razorpayInvoiceId: string,
  status: Invoice['status'],
  razorpayPaymentId?: string,
  paidAt?: string,
): Promise<void> {
  type Update = Database['public']['Tables']['invoices']['Update']
  const updateData: Update = { status }
  if (razorpayPaymentId) updateData.razorpay_payment_id = razorpayPaymentId
  if (paidAt) updateData.paid_at = paidAt
  const { error } = await supabase
    .from('invoices')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update(updateData)
    .eq('razorpay_invoice_id', razorpayInvoiceId)
  if (error) {
    throw new Error(`Failed to update invoice status: ${error.message}`)
  }
}

// --- Renewal Job Scheduling ---

export async function scheduleRenewalJobs(
  installationId: string,
  currentEnd: string,
  billingPeriod: 'monthly' | 'annual' = 'monthly',
  graceDays = 7,
): Promise<void> {
  const endDate = new Date(currentEnd)

  // Annual subs: send renewal notice 7 days before end; Monthly: 24h before
  const renewalLeadMs =
    billingPeriod === 'annual' ? 7 * 24 * 60 * 60 * 1000 : 24 * 60 * 60 * 1000
  const renewalAt = new Date(endDate.getTime() - renewalLeadMs).toISOString()

  // Schedule expiry after grace period
  const expireAt = new Date(
    endDate.getTime() + graceDays * 24 * 60 * 60 * 1000,
  ).toISOString()

  const { error } = await supabase
    .from('renewal_jobs')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .insert([
      {
        installation_id: installationId,
        job_type: 'renewal',
        scheduled_at: renewalAt,
        status: 'pending',
      },
      {
        installation_id: installationId,
        job_type: 'expire',
        scheduled_at: expireAt,
        status: 'pending',
      },
    ])

  if (error) {
    throw new Error(`Failed to schedule renewal jobs: ${error.message}`)
  }
}

export async function cancelRenewalJobs(installationId: string): Promise<void> {
  const { error } = await supabase
    .from('renewal_jobs')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update({ status: 'cancelled' })
    .eq('installation_id', installationId)
    .eq('status', 'pending')

  if (error) {
    throw new Error(`Failed to cancel renewal jobs: ${error.message}`)
  }
}

export async function upsertInvoice(data: {
  subscriptionId: string
  installationId: string
  razorpayInvoiceId: string
  razorpayPaymentId?: string
  amount: number
  currency: string
  status: Invoice['status']
  billingStart?: string
  billingEnd?: string
  paidAt?: string
}): Promise<void> {
  type Insert = Database['public']['Tables']['invoices']['Insert']
  const insertData: Insert = {
    subscription_id: data.subscriptionId,
    installation_id: data.installationId,
    razorpay_invoice_id: data.razorpayInvoiceId,
    razorpay_payment_id: data.razorpayPaymentId ?? null,
    amount: data.amount,
    currency: data.currency,
    status: data.status,
    billing_start: data.billingStart ?? null,
    billing_end: data.billingEnd ?? null,
    paid_at: data.paidAt ?? null,
  }
  const { error } = await supabase
    .from('invoices')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .upsert(insertData, { onConflict: 'razorpay_invoice_id' })
  if (error) {
    throw new Error(`Failed to upsert invoice: ${error.message}`)
  }
}
