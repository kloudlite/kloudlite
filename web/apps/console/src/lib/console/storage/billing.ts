import { supabase } from '../supabase'
import type { Database } from '../supabase-types'
import type { BillingAccount, SubscriptionItem } from './billing-types'

// --- Mapping helpers ---

type BillingAccountRow = Database['public']['Tables']['billing_accounts']['Row']
type SubscriptionItemRow = Database['public']['Tables']['subscription_items']['Row']

function mapToBillingAccount(row: BillingAccountRow): BillingAccount {
  return {
    id: row.id,
    orgId: row.org_id,
    stripeCustomerId: row.stripe_customer_id,
    stripeSubscriptionId: row.stripe_subscription_id,
    billingStatus: row.billing_status,
    hasPaymentIssue: row.has_payment_issue,
    currentPeriodEnd: row.current_period_end,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  }
}

function mapToSubscriptionItem(row: SubscriptionItemRow): SubscriptionItem {
  return {
    id: row.id,
    orgId: row.org_id,
    installationId: row.installation_id,
    stripeItemId: row.stripe_item_id,
    stripePriceId: row.stripe_price_id,
    tier: row.tier,
    productName: row.product_name,
    quantity: row.quantity,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  }
}

// --- Billing Account CRUD ---

export async function getBillingAccount(orgId: string): Promise<BillingAccount | null> {
  const { data, error } = await supabase
    .from('billing_accounts')
    .select('*')
    .eq('org_id', orgId)
    .single()
  if (error) return null
  return mapToBillingAccount(data as BillingAccountRow)
}

export async function getBillingAccountByCustomerId(
  stripeCustomerId: string,
): Promise<BillingAccount | null> {
  const { data, error } = await supabase
    .from('billing_accounts')
    .select('*')
    .eq('stripe_customer_id', stripeCustomerId)
    .single()
  if (error) return null
  return mapToBillingAccount(data as BillingAccountRow)
}

export async function upsertBillingAccount(data: {
  orgId: string
  stripeCustomerId: string
  stripeSubscriptionId?: string | null
  billingStatus: BillingAccount['billingStatus']
  currentPeriodEnd?: string | null
}): Promise<void> {
  type Insert = Database['public']['Tables']['billing_accounts']['Insert']
  const insertData: Insert = {
    org_id: data.orgId,
    stripe_customer_id: data.stripeCustomerId,
    stripe_subscription_id: data.stripeSubscriptionId ?? null,
    billing_status: data.billingStatus,
    current_period_end: data.currentPeriodEnd ?? null,
  }
  const { error } = await supabase
    .from('billing_accounts')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .upsert(insertData, { onConflict: 'org_id' })
  if (error) {
    throw new Error(`Failed to upsert billing account: ${error.message}`)
  }
}

export async function updateBillingStatus(
  stripeCustomerId: string,
  billingStatus: BillingAccount['billingStatus'],
  currentPeriodEnd?: string | null,
  paymentIssue?: boolean,
): Promise<void> {
  type Update = Database['public']['Tables']['billing_accounts']['Update']
  const updateData: Update = { billing_status: billingStatus }
  if (currentPeriodEnd !== undefined) updateData.current_period_end = currentPeriodEnd
  if (paymentIssue !== undefined) updateData.has_payment_issue = paymentIssue
  const { error } = await supabase
    .from('billing_accounts')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update(updateData)
    .eq('stripe_customer_id', stripeCustomerId)
  if (error) {
    throw new Error(`Failed to update billing status: ${error.message}`)
  }
}

// --- Subscription Items CRUD ---

export async function getSubscriptionItems(orgId: string): Promise<SubscriptionItem[]> {
  const { data, error } = await supabase
    .from('subscription_items')
    .select('*')
    .eq('org_id', orgId)
    .order('tier', { ascending: true })
  if (error) return []
  return (data as SubscriptionItemRow[]).map(mapToSubscriptionItem)
}

export async function syncSubscriptionItems(
  orgId: string,
  items: Array<{
    stripeItemId: string
    stripePriceId: string
    tier: number
    productName: string
    quantity: number
    installationId?: string | null
  }>,
): Promise<void> {
  const jsonItems = items.map((item) => ({
    installation_id: item.installationId || null,
    stripe_item_id: item.stripeItemId,
    stripe_price_id: item.stripePriceId,
    tier: item.tier ?? null,
    product_name: item.productName,
    quantity: item.quantity,
  }))

  const { error } = await (supabase as any).rpc('sync_subscription_items', {
    p_org_id: orgId,
    p_items: jsonItems,
  })

  if (!error) return

  // Fall back to two-step delete+insert if RPC doesn't exist yet
  if (error.code === '42883' || error.message?.includes('function') || error.code === 'PGRST202') {
    await supabase
      .from('subscription_items')
      .delete()
      .eq('org_id', orgId)

    if (items.length > 0) {
      const rows = items.map((item) => ({
        org_id: orgId,
        installation_id: item.installationId || null,
        stripe_item_id: item.stripeItemId,
        stripe_price_id: item.stripePriceId,
        tier: item.tier ?? null,
        product_name: item.productName,
        quantity: item.quantity,
      }))

      const { error: insertError } = await supabase
        .from('subscription_items')
        // @ts-expect-error — Supabase generic inference resolves mutations to never
        .insert(rows)

      if (insertError) {
        throw new Error(`Failed to insert subscription items: ${insertError.message}`)
      }
    }
    return
  }

  throw new Error(`Failed to sync subscription items: ${error.message}`)
}

// --- Sync Items from Stripe ---

/**
 * Fetch subscription items directly from Stripe and sync to DB.
 * Useful when the webhook hasn't fired yet (e.g., after checkout redirect).
 */
export async function syncSubscriptionItemsFromStripe(
  orgId: string,
  stripeSubscriptionId: string,
): Promise<void> {
  try {
    const { getStripe } = await import('@/lib/stripe')
    const stripe = getStripe()
    const subscription = await stripe.subscriptions.retrieve(stripeSubscriptionId, {
      expand: ['items.data.price.product'],
    })

    const items = subscription.items.data.map((item) => {
      const product = item.price.product as { name: string; metadata?: Record<string, string> }
      const tier = product.metadata?.tier ? parseInt(product.metadata.tier, 10) : 0
      return {
        stripeItemId: item.id,
        stripePriceId: item.price.id,
        tier,
        productName: product.name,
        quantity: item.quantity ?? 1,
      }
    })

    await syncSubscriptionItems(orgId, items)
    console.log(`[billing] Synced ${items.length} subscription items from Stripe for org ${orgId}`)
  } catch (err) {
    console.error(`[billing] Failed to sync subscription items from Stripe:`, err)
  }
}

// --- Subscription Cancellation ---

/**
 * Cancel the Stripe subscription for an organization.
 */
export async function cancelSubscriptionForOrg(orgId: string): Promise<void> {
  try {
    const account = await getBillingAccount(orgId)
    if (!account?.stripeSubscriptionId) return

    const { getStripe } = await import('@/lib/stripe')
    const stripe = getStripe()
    await stripe.subscriptions.cancel(account.stripeSubscriptionId)
    console.log(
      `[billing] Cancelled Stripe subscription ${account.stripeSubscriptionId} for org ${orgId}`,
    )
  } catch (err) {
    console.error(
      `[billing] Failed to cancel Stripe subscription for org ${orgId}:`,
      err,
    )
  }
}

// --- Webhook Idempotency ---

export async function isWebhookEventProcessed(stripeEventId: string): Promise<boolean> {
  const { count, error } = await supabase
    .from('processed_webhook_events')
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
  type Insert = Database['public']['Tables']['processed_webhook_events']['Insert']
  const insertData: Insert = {
    stripe_event_id: stripeEventId,
    event_type: eventType,
  }
  const { error } = await supabase
    .from('processed_webhook_events')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .insert(insertData)
  if (error) {
    console.error('Failed to mark webhook event as processed:', error.message)
  }
}
