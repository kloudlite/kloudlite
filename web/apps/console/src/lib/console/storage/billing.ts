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
    console.error('Error fetching plans:', error)
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
  if (error) {
    if (error.code === 'PGRST116') return null
    console.error('Error fetching plan:', error)
    return null
  }
  return mapToPlan(data as PlanRow)
}

export async function updatePlanRazorpayId(planId: string, razorpayPlanId: string): Promise<void> {
  const { error } = await supabase
    .from('subscription_plans')
    .update({ razorpay_plan_id: razorpayPlanId })
    .eq('id', planId)
  if (error) {
    throw new Error(`Failed to update plan razorpay_plan_id: ${error.message}`)
  }
}

export async function getSubscriptionByInstallation(installationId: string): Promise<Subscription | null> {
  const { data, error } = await supabase
    .from('subscriptions')
    .select('*')
    .eq('installation_id', installationId)
    .single()
  if (error) {
    if (error.code === 'PGRST116') return null
    console.error('Error fetching subscription:', error)
    return null
  }
  return mapToSubscription(data as SubscriptionRow)
}

export async function getSubscriptionByRazorpayId(razorpaySubscriptionId: string): Promise<Subscription | null> {
  const { data, error } = await supabase
    .from('subscriptions')
    .select('*')
    .eq('razorpay_subscription_id', razorpaySubscriptionId)
    .single()
  if (error) {
    if (error.code === 'PGRST116') return null
    console.error('Error fetching subscription by razorpay ID:', error)
    return null
  }
  return mapToSubscription(data as SubscriptionRow)
}

export async function createSubscription(data: {
  installationId: string
  planId: string
  razorpaySubscriptionId: string
  razorpayCustomerId: string
  quantity: number
}): Promise<Subscription> {
  type Insert = Database['public']['Tables']['subscriptions']['Insert']
  const insertData: Insert = {
    installation_id: data.installationId,
    plan_id: data.planId,
    razorpay_subscription_id: data.razorpaySubscriptionId,
    razorpay_customer_id: data.razorpayCustomerId,
    quantity: data.quantity,
    status: 'created',
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
    .update(updateData)
    .eq('razorpay_subscription_id', razorpaySubscriptionId)
  if (error) {
    throw new Error(`Failed to update subscription: ${error.message}`)
  }
}

export async function getInvoicesByInstallation(installationId: string): Promise<Invoice[]> {
  const { data, error } = await supabase
    .from('invoices')
    .select('*')
    .eq('installation_id', installationId)
    .order('created_at', { ascending: false })
  if (error) {
    console.error('Error fetching invoices:', error)
    return []
  }
  return (data as InvoiceRow[]).map(mapToInvoice)
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
