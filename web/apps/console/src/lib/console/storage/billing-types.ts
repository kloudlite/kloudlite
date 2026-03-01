import type { Database } from '../supabase-types'

export type PlanRow = Database['public']['Tables']['subscription_plans']['Row']
export type SubscriptionRow = Database['public']['Tables']['subscriptions']['Row']
export type InvoiceRow = Database['public']['Tables']['invoices']['Row']

export interface Plan {
  id: string
  razorpayPlanId: string | null
  tier: number
  name: string
  amountPerUser: number
  baseFee: number
  currency: string
  monthlyHours: number
  overageRate: number
  cpu: number
  ram: string
  storage: string
  autoSuspend: string
  description: string | null
  createdAt: string
}

export interface Subscription {
  id: string
  installationId: string
  planId: string
  razorpaySubscriptionId: string | null
  razorpayCustomerId: string | null
  status: 'created' | 'authenticated' | 'active' | 'paused' | 'cancelled' | 'expired'
  quantity: number
  currentStart: string | null
  currentEnd: string | null
  createdAt: string
  updatedAt: string
}

export interface Invoice {
  id: string
  subscriptionId: string
  installationId: string
  razorpayInvoiceId: string | null
  razorpayPaymentId: string | null
  amount: number
  currency: string
  status: 'issued' | 'paid' | 'expired' | 'cancelled'
  billingStart: string | null
  billingEnd: string | null
  paidAt: string | null
  createdAt: string
}
