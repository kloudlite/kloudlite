'use server'

import { redirect } from 'next/navigation'
import { requireOrgAccess } from '@/lib/console/authorization'
import { getBillingAccount } from '@/lib/console/storage'
import { getCreditAccount } from '@/lib/console/storage/credits'
import type { BillingAccount } from '@/lib/console/storage'
import type { CreditAccount } from '@/lib/console/storage/credits-types'

export async function getStripePublishableKey(): Promise<string> {
  const key = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
  if (!key) throw new Error('NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY not configured')
  return key
}

export async function fetchBillingStatus(orgId: string): Promise<{
  customer: BillingAccount | null
  creditAccount: CreditAccount | null
}> {
  try {
    await requireOrgAccess(orgId)
  } catch {
    redirect('/login')
  }

  const [customer, creditAccount] = await Promise.all([
    getBillingAccount(orgId),
    getCreditAccount(orgId),
  ])

  return { customer, creditAccount }
}
