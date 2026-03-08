export interface BillingAccount {
  id: string
  orgId: string
  stripeCustomerId: string
  stripeSubscriptionId: string | null
  billingStatus: 'active' | 'past_due' | 'cancelled' | 'trialing' | 'incomplete'
  hasPaymentIssue: boolean
  currentPeriodEnd: string | null
  createdAt: string
  updatedAt: string
}

export interface SubscriptionItem {
  id: string
  orgId: string
  installationId: string | null
  stripeItemId: string
  stripePriceId: string
  tier: number
  productName: string
  quantity: number
  createdAt: string
  updatedAt: string
}
