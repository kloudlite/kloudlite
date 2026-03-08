'use server'

import { redirect } from 'next/navigation'
import { requireOrgAccess } from '@/lib/console/authorization'
import {
  getBillingAccount,
  getSubscriptionItems,
} from '@/lib/console/storage'
import type { BillingAccount, SubscriptionItem } from '@/lib/console/storage'

export async function getStripePublishableKey(): Promise<string> {
  const key = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
  if (!key) throw new Error('NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY not configured')
  return key
}

export async function fetchBillingStatus(orgId: string): Promise<{
  customer: BillingAccount | null
  items: SubscriptionItem[]
}> {
  try {
    await requireOrgAccess(orgId)
  } catch {
    redirect('/login')
  }

  const [customer, items] = await Promise.all([
    getBillingAccount(orgId),
    getSubscriptionItems(orgId),
  ])

  return { customer, items }
}
