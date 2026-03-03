'use server'

import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getPlans,
  getSubscriptionsByInstallation,
  getInvoicesByInstallation,
} from '@/lib/console/storage'
import type { Plan, Subscription, Invoice } from '@/lib/console/storage'

export async function getRazorpayKey(): Promise<string> {
  const keyId = process.env.RAZORPAY_KEY_ID
  if (!keyId) throw new Error('RAZORPAY_KEY_ID not configured')
  return keyId
}

export async function fetchPlans(): Promise<Plan[]> {
  return getPlans()
}

export async function fetchSubscriptions(installationId: string): Promise<Subscription[]> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  return getSubscriptionsByInstallation(installationId)
}

export async function fetchInvoices(installationId: string): Promise<Invoice[]> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  return getInvoicesByInstallation(installationId)
}
