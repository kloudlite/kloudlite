'use server'

import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getStripeCustomer,
  getSubscriptionItems,
  getMemberRole,
  getInstallationById,
} from '@/lib/console/storage'
import type { StripeCustomer, SubscriptionItem } from '@/lib/console/storage'

export async function getStripePublishableKey(): Promise<string> {
  const key = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
  if (!key) throw new Error('NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY not configured')
  return key
}

export async function fetchBillingStatus(installationId: string): Promise<{
  customer: StripeCustomer | null
  items: SubscriptionItem[]
}> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const role = await getMemberRole(installationId, session.user.id)
  const installation = await getInstallationById(installationId)
  if (!role && installation?.userId !== session.user.id) {
    throw new Error('Forbidden')
  }

  const [customer, items] = await Promise.all([
    getStripeCustomer(installationId),
    getSubscriptionItems(installationId),
  ])

  return { customer, items }
}
