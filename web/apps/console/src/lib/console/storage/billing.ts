import { supabase } from '../supabase'
import type { Database } from '../supabase-types'
import type { StripeCustomer, SubscriptionItem } from './billing-types'

// --- Mapping helpers ---

type StripeCustomerRow = Database['public']['Tables']['stripe_customers']['Row']
type SubscriptionItemRow = Database['public']['Tables']['subscription_items']['Row']

function mapToStripeCustomer(row: StripeCustomerRow): StripeCustomer {
  return {
    id: row.id,
    installationId: row.installation_id,
    stripeCustomerId: row.stripe_customer_id,
    stripeSubscriptionId: row.stripe_subscription_id,
    billingStatus: row.billing_status,
    paymentIssue: row.payment_issue,
    currentPeriodEnd: row.current_period_end,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  }
}

function mapToSubscriptionItem(row: SubscriptionItemRow): SubscriptionItem {
  return {
    id: row.id,
    installationId: row.installation_id,
    stripeSubscriptionItemId: row.stripe_subscription_item_id,
    stripePriceId: row.stripe_price_id,
    tier: row.tier,
    productName: row.product_name,
    quantity: row.quantity,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  }
}

// --- Stripe Customer CRUD ---

export async function getStripeCustomer(installationId: string): Promise<StripeCustomer | null> {
  const { data, error } = await supabase
    .from('stripe_customers')
    .select('*')
    .eq('installation_id', installationId)
    .single()
  if (error) return null
  return mapToStripeCustomer(data as StripeCustomerRow)
}

export async function getStripeCustomerByCustomerId(
  stripeCustomerId: string,
): Promise<StripeCustomer | null> {
  const { data, error } = await supabase
    .from('stripe_customers')
    .select('*')
    .eq('stripe_customer_id', stripeCustomerId)
    .single()
  if (error) return null
  return mapToStripeCustomer(data as StripeCustomerRow)
}

export async function upsertStripeCustomer(data: {
  installationId: string
  stripeCustomerId: string
  stripeSubscriptionId?: string | null
  billingStatus: StripeCustomer['billingStatus']
  currentPeriodEnd?: string | null
}): Promise<void> {
  type Insert = Database['public']['Tables']['stripe_customers']['Insert']
  const insertData: Insert = {
    installation_id: data.installationId,
    stripe_customer_id: data.stripeCustomerId,
    stripe_subscription_id: data.stripeSubscriptionId ?? null,
    billing_status: data.billingStatus,
    current_period_end: data.currentPeriodEnd ?? null,
  }
  const { error } = await supabase
    .from('stripe_customers')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .upsert(insertData, { onConflict: 'installation_id' })
  if (error) {
    throw new Error(`Failed to upsert stripe customer: ${error.message}`)
  }
}

export async function updateBillingStatus(
  stripeCustomerId: string,
  billingStatus: StripeCustomer['billingStatus'],
  currentPeriodEnd?: string | null,
  paymentIssue?: boolean,
): Promise<void> {
  type Update = Database['public']['Tables']['stripe_customers']['Update']
  const updateData: Update = { billing_status: billingStatus }
  if (currentPeriodEnd !== undefined) updateData.current_period_end = currentPeriodEnd
  if (paymentIssue !== undefined) updateData.payment_issue = paymentIssue
  const { error } = await supabase
    .from('stripe_customers')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update(updateData)
    .eq('stripe_customer_id', stripeCustomerId)
  if (error) {
    throw new Error(`Failed to update billing status: ${error.message}`)
  }
}

export async function getActiveSubscriptionsByInstallationIds(
  installationIds: string[],
): Promise<Record<string, StripeCustomer>> {
  if (installationIds.length === 0) return {}

  const { data, error } = await supabase
    .from('stripe_customers')
    .select('*')
    .in('installation_id', installationIds)
    .eq('billing_status', 'active')

  if (error) return {}

  const result: Record<string, StripeCustomer> = {}
  for (const row of data as StripeCustomerRow[]) {
    // Keep only the first active customer per installation
    if (!result[row.installation_id]) {
      result[row.installation_id] = mapToStripeCustomer(row)
    }
  }
  return result
}

// --- Subscription Items CRUD ---

export async function getSubscriptionItems(installationId: string): Promise<SubscriptionItem[]> {
  const { data, error } = await supabase
    .from('subscription_items')
    .select('*')
    .eq('installation_id', installationId)
    .order('tier', { ascending: true })
  if (error) return []
  return (data as SubscriptionItemRow[]).map(mapToSubscriptionItem)
}

export async function syncSubscriptionItems(
  installationId: string,
  items: Array<{
    stripeSubscriptionItemId: string
    stripePriceId: string
    tier: number
    productName: string
    quantity: number
  }>,
): Promise<void> {
  // Delete existing items for this installation
  const { error: deleteError } = await supabase
    .from('subscription_items')
    .delete()
    .eq('installation_id', installationId)
  if (deleteError) {
    throw new Error(`Failed to clear subscription items: ${deleteError.message}`)
  }

  if (items.length === 0) return

  type Insert = Database['public']['Tables']['subscription_items']['Insert']
  const insertData: Insert[] = items.map((item) => ({
    installation_id: installationId,
    stripe_subscription_item_id: item.stripeSubscriptionItemId,
    stripe_price_id: item.stripePriceId,
    tier: item.tier,
    product_name: item.productName,
    quantity: item.quantity,
  }))

  const { error } = await supabase
    .from('subscription_items')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .insert(insertData)
  if (error) {
    throw new Error(`Failed to sync subscription items: ${error.message}`)
  }
}

// --- Webhook Idempotency ---

export async function isWebhookEventProcessed(stripeEventId: string): Promise<boolean> {
  const { count, error } = await supabase
    .from('stripe_webhook_events')
    .select('*', { count: 'exact', head: true })
    .eq('stripe_event_id', stripeEventId)

  if (error) {
    console.error('Failed to check webhook event:', error.message)
    return false
  }
  return (count ?? 0) > 0
}

export async function markWebhookEventProcessed(
  stripeEventId: string,
  eventType: string,
): Promise<void> {
  type Insert = Database['public']['Tables']['stripe_webhook_events']['Insert']
  const insertData: Insert = {
    stripe_event_id: stripeEventId,
    event_type: eventType,
  }
  const { error } = await supabase
    .from('stripe_webhook_events')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .insert(insertData)
  if (error) {
    console.error('Failed to mark webhook event as processed:', error.message)
  }
}
