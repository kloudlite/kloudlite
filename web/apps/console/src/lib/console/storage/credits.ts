import { supabase } from '../supabase'
import type {
  CreditAccount,
  CreditTransaction,
  UsageEvent,
  UsagePeriod,
  PricingTier,
} from './credits-types'

// --- Mapping helpers ---

function mapToCreditAccount(row: any): CreditAccount {
  return {
    id: row.id,
    orgId: row.org_id,
    balance: row.balance,
    autoTopupEnabled: row.auto_topup_enabled,
    autoTopupThreshold: row.auto_topup_threshold,
    autoTopupAmount: row.auto_topup_amount,
    stripeCustomerId: row.stripe_customer_id,
    negativeBalanceFlagged: row.negative_balance_flagged,
    lowBalanceWarning: row.low_balance_warning,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  }
}

function mapToCreditTransaction(row: any): CreditTransaction {
  return {
    id: row.id,
    orgId: row.org_id,
    type: row.type,
    amount: row.amount,
    description: row.description,
    stripeInvoiceId: row.stripe_invoice_id,
    usagePeriodId: row.usage_period_id,
    expiresAt: row.expires_at,
    createdAt: row.created_at,
  }
}

function mapToUsageEvent(row: any): UsageEvent {
  return {
    id: row.id,
    installationId: row.installation_id,
    eventType: row.event_type,
    resourceId: row.resource_id,
    resourceType: row.resource_type,
    metadata: row.metadata,
    eventTimestamp: row.event_timestamp,
    createdAt: row.created_at,
  }
}

function mapToUsagePeriod(row: any): UsagePeriod {
  return {
    id: row.id,
    installationId: row.installation_id,
    orgId: row.org_id,
    resourceId: row.resource_id,
    resourceType: row.resource_type,
    startedAt: row.started_at,
    endedAt: row.ended_at,
    hourlyRate: row.hourly_rate,
    totalCost: row.total_cost,
    lastBilledAt: row.last_billed_at,
    createdAt: row.created_at,
  }
}

function mapToPricingTier(row: any): PricingTier {
  return {
    id: row.id,
    resourceType: row.resource_type,
    displayName: row.display_name,
    hourlyRate: row.hourly_rate,
    unit: row.unit,
    category: row.category,
    specs: row.specs,
    isActive: row.is_active,
    createdAt: row.created_at,
  }
}

// --- Credit Accounts ---

export async function getCreditAccount(orgId: string): Promise<CreditAccount | null> {
  const { data, error } = await supabase
    .from('credit_accounts')
    .select('*')
    .eq('org_id', orgId)
    .single()
  if (error) return null
  return mapToCreditAccount(data)
}

export async function getCreditAccountByCustomerId(
  stripeCustomerId: string,
): Promise<CreditAccount | null> {
  const { data, error } = await supabase
    .from('credit_accounts')
    .select('*')
    .eq('stripe_customer_id', stripeCustomerId)
    .single()
  if (error) return null
  return mapToCreditAccount(data)
}

export async function ensureCreditAccount(
  orgId: string,
  stripeCustomerId?: string,
): Promise<CreditAccount> {
  const insertData: Record<string, unknown> = {
    org_id: orgId,
  }
  if (stripeCustomerId !== undefined) {
    insertData.stripe_customer_id = stripeCustomerId
  }

  const { data, error } = await supabase
    .from('credit_accounts')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .upsert(insertData, { onConflict: 'org_id' })
    .select('*')
    .single()
  if (error) {
    throw new Error(`Failed to ensure credit account: ${error.message}`)
  }
  return mapToCreditAccount(data)
}

export async function updateCreditAccountAutoTopup(
  orgId: string,
  enabled: boolean,
  threshold?: number,
  amount?: number,
): Promise<void> {
  const updateData: Record<string, unknown> = {
    auto_topup_enabled: enabled,
  }
  if (threshold !== undefined) updateData.auto_topup_threshold = threshold
  if (amount !== undefined) updateData.auto_topup_amount = amount

  const { error } = await supabase
    .from('credit_accounts')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update(updateData)
    .eq('org_id', orgId)
  if (error) {
    throw new Error(`Failed to update auto-topup settings: ${error.message}`)
  }
}

export async function updateCreditAccountWarnings(
  orgId: string,
  lowBalanceWarning: boolean,
  negativeBalanceFlagged?: boolean,
): Promise<void> {
  const updateData: Record<string, unknown> = {
    low_balance_warning: lowBalanceWarning,
  }
  if (negativeBalanceFlagged !== undefined) {
    updateData.negative_balance_flagged = negativeBalanceFlagged
  }

  const { error } = await supabase
    .from('credit_accounts')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update(updateData)
    .eq('org_id', orgId)
  if (error) {
    throw new Error(`Failed to update credit account warnings: ${error.message}`)
  }
}

export async function updateCreditAccountStripeCustomer(
  orgId: string,
  stripeCustomerId: string,
): Promise<void> {
  const { error } = await supabase
    .from('credit_accounts')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update({ stripe_customer_id: stripeCustomerId })
    .eq('org_id', orgId)
  if (error) {
    throw new Error(`Failed to update stripe customer: ${error.message}`)
  }
}

// --- Credit Transactions ---

export async function getCreditTransactions(
  orgId: string,
  limit: number = 50,
  offset: number = 0,
): Promise<CreditTransaction[]> {
  const { data, error } = await supabase
    .from('credit_transactions')
    .select('*')
    .eq('org_id', orgId)
    .order('created_at', { ascending: false })
    .range(offset, offset + limit - 1)
  if (error) return []
  return (data as any[]).map(mapToCreditTransaction)
}

export async function getCreditTransactionsByPeriod(
  orgId: string,
  startDate: string,
  endDate: string,
): Promise<CreditTransaction[]> {
  const { data, error } = await supabase
    .from('credit_transactions')
    .select('*')
    .eq('org_id', orgId)
    .gte('created_at', startDate)
    .lte('created_at', endDate)
    .order('created_at', { ascending: false })
  if (error) return []
  return (data as any[]).map(mapToCreditTransaction)
}

export async function topupCredits(
  orgId: string,
  amount: number,
  description: string,
  stripeInvoiceId?: string,
): Promise<number> {
  const { data, error } = await (supabase as any).rpc('topup_credits', {
    p_org_id: orgId,
    p_amount: amount,
    p_description: description,
    p_stripe_invoice_id: stripeInvoiceId ?? null,
  })
  if (error) {
    throw new Error(`Failed to topup credits: ${error.message}`)
  }
  return data as number
}

export async function debitCredits(
  orgId: string,
  amount: number,
  description: string,
  usagePeriodId?: string,
): Promise<number> {
  const { data, error } = await (supabase as any).rpc('debit_credits', {
    p_org_id: orgId,
    p_amount: amount,
    p_description: description,
    p_usage_period_id: usagePeriodId ?? null,
  })
  if (error) {
    throw new Error(`Failed to debit credits: ${error.message}`)
  }
  return data as number
}

// --- Usage Events ---

export async function insertUsageEvent(event: {
  installationId: string
  eventType: string
  resourceId: string
  resourceType?: string
  metadata?: Record<string, unknown>
  eventTimestamp: string
}): Promise<UsageEvent | null> {
  const insertData: Record<string, unknown> = {
    installation_id: event.installationId,
    event_type: event.eventType,
    resource_id: event.resourceId,
    resource_type: event.resourceType ?? null,
    metadata: event.metadata ?? {},
    event_timestamp: event.eventTimestamp,
  }

  const { data, error } = await supabase
    .from('usage_events')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .insert(insertData)
    .select('*')
    .single()
  if (error) {
    // Handle unique constraint violations gracefully
    if (error.code === '23505') return null
    throw new Error(`Failed to insert usage event: ${error.message}`)
  }
  return mapToUsageEvent(data)
}

// --- Usage Periods ---

export async function getActiveUsagePeriods(orgId: string): Promise<UsagePeriod[]> {
  const { data, error } = await supabase
    .from('usage_periods')
    .select('*')
    .eq('org_id', orgId)
    .is('ended_at', null)
  if (error) return []
  return (data as any[]).map(mapToUsagePeriod)
}

export async function getActiveUsagePeriodsForInstallation(
  installationId: string,
): Promise<UsagePeriod[]> {
  const { data, error } = await supabase
    .from('usage_periods')
    .select('*')
    .eq('installation_id', installationId)
    .is('ended_at', null)
  if (error) return []
  return (data as any[]).map(mapToUsagePeriod)
}

export async function openUsagePeriod(params: {
  installationId: string
  orgId: string
  resourceId: string
  resourceType: string
  hourlyRate: number
}): Promise<UsagePeriod> {
  const insertData: Record<string, unknown> = {
    installation_id: params.installationId,
    org_id: params.orgId,
    resource_id: params.resourceId,
    resource_type: params.resourceType,
    hourly_rate: params.hourlyRate,
    started_at: new Date().toISOString(),
    total_cost: 0,
    last_billed_at: new Date().toISOString(),
  }

  const { data, error } = await supabase
    .from('usage_periods')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .insert(insertData)
    .select('*')
    .single()
  if (error) {
    throw new Error(`Failed to open usage period: ${error.message}`)
  }
  return mapToUsagePeriod(data)
}

export async function closeUsagePeriod(
  resourceId: string,
  installationId: string,
): Promise<UsagePeriod | null> {
  // Find the open period
  const { data: existing, error: findError } = await supabase
    .from('usage_periods')
    .select('*')
    .eq('resource_id', resourceId)
    .eq('installation_id', installationId)
    .is('ended_at', null)
    .single()
  if (findError || !existing) return null

  const period = mapToUsagePeriod(existing)
  const now = new Date()
  const hoursElapsed = (now.getTime() - Date.parse(period.startedAt)) / (1000 * 60 * 60)
  const totalCost = hoursElapsed * period.hourlyRate

  const { data, error } = await supabase
    .from('usage_periods')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update({
      ended_at: now.toISOString(),
      total_cost: totalCost,
    })
    .eq('id', period.id)
    .select('*')
    .single()
  if (error) {
    throw new Error(`Failed to close usage period: ${error.message}`)
  }
  return mapToUsagePeriod(data)
}

export async function closeAllUsagePeriodsForInstallation(
  installationId: string,
): Promise<void> {
  const periods = await getActiveUsagePeriodsForInstallation(installationId)

  for (const period of periods) {
    await closeUsagePeriod(period.resourceId, installationId)
  }
}

export async function updateLastBilledAt(periodIds: string[]): Promise<void> {
  if (periodIds.length === 0) return

  const { error } = await supabase
    .from('usage_periods')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update({ last_billed_at: new Date().toISOString() })
    .in('id', periodIds)
  if (error) {
    throw new Error(`Failed to update last billed at: ${error.message}`)
  }
}

export async function getUsagePeriodsByDateRange(
  orgId: string,
  startDate: string,
  endDate: string,
): Promise<UsagePeriod[]> {
  const { data, error } = await supabase
    .from('usage_periods')
    .select('*')
    .eq('org_id', orgId)
    .gte('started_at', startDate)
    .lte('started_at', endDate)
    .order('started_at', { ascending: false })
  if (error) return []
  return (data as any[]).map(mapToUsagePeriod)
}

// --- Pricing Tiers ---

export async function getPricingTiers(): Promise<PricingTier[]> {
  const { data, error } = await supabase
    .from('pricing_tiers')
    .select('*')
    .eq('is_active', true)
    .order('category', { ascending: true })
    .order('resource_type', { ascending: true })
  if (error) return []
  return (data as any[]).map(mapToPricingTier)
}

export async function getPricingTierByType(
  resourceType: string,
): Promise<PricingTier | null> {
  const { data, error } = await supabase
    .from('pricing_tiers')
    .select('*')
    .eq('resource_type', resourceType)
    .eq('is_active', true)
    .single()
  if (error) return null
  return mapToPricingTier(data)
}

export async function getHourlyRate(resourceType: string): Promise<number> {
  const tier = await getPricingTierByType(resourceType)
  if (!tier) {
    throw new Error(`No active pricing tier found for resource type: ${resourceType}`)
  }
  return tier.hourlyRate
}
