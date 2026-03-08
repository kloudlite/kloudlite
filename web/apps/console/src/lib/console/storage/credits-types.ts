export interface CreditAccount {
  id: string
  orgId: string
  balance: number
  autoTopupEnabled: boolean
  autoTopupThreshold: number | null
  autoTopupAmount: number | null
  stripeCustomerId: string | null
  negativeBalanceFlagged: boolean
  lowBalanceWarning: boolean
  createdAt: string
  updatedAt: string
}

export interface CreditTransaction {
  id: string
  orgId: string
  type: 'topup' | 'usage_debit' | 'adjustment'
  amount: number
  description: string | null
  stripeInvoiceId: string | null
  usagePeriodId: string | null
  expiresAt: string | null
  createdAt: string
}

export interface UsageEvent {
  id: string
  installationId: string
  eventType: string
  resourceId: string
  resourceType: string | null
  metadata: Record<string, unknown>
  eventTimestamp: string
  createdAt: string
}

export interface UsagePeriod {
  id: string
  installationId: string
  orgId: string
  resourceId: string
  resourceType: string
  startedAt: string
  endedAt: string | null
  hourlyRate: number
  totalCost: number
  lastBilledAt: string
  createdAt: string
}

export interface PricingTier {
  id: string
  resourceType: string
  displayName: string
  hourlyRate: number
  unit: string
  category: 'compute' | 'storage'
  specs: Record<string, unknown>
  isActive: boolean
  createdAt: string
}
