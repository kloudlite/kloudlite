export interface StripeCustomer {
  id: string
  installationId: string
  stripeCustomerId: string
  stripeSubscriptionId: string | null
  billingStatus: 'active' | 'past_due' | 'cancelled' | 'trialing' | 'incomplete'
  paymentIssue: boolean
  currentPeriodEnd: string | null
  createdAt: string
  updatedAt: string
}

export interface SubscriptionItem {
  id: string
  installationId: string
  stripeSubscriptionItemId: string
  stripePriceId: string
  tier: number
  productName: string
  quantity: number
  createdAt: string
  updatedAt: string
}
