'use server'

import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getRazorpay } from '@/lib/razorpay'
import { syncPlansToRazorpay } from '@/lib/razorpay-plans'
import {
  getPlans,
  getPlanById,
  getSubscriptionByInstallation,
  createSubscription,
  getInvoicesByInstallation,
} from '@/lib/console/storage'
import type { Plan, Subscription, Invoice } from '@/lib/console/storage'
import { getInstallationById, getMemberRole } from '@/lib/console/storage'

// --- Read Actions ---

export async function fetchPlans(): Promise<Plan[]> {
  return getPlans()
}

export async function fetchSubscription(installationId: string): Promise<Subscription | null> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  return getSubscriptionByInstallation(installationId)
}

export async function fetchInvoices(installationId: string): Promise<Invoice[]> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  return getInvoicesByInstallation(installationId)
}

// --- Write Actions ---

export async function createNewSubscription(
  installationId: string,
  planId: string,
  quantity: number,
): Promise<{ subscriptionId: string; razorpaySubscriptionId: string }> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  // Verify owner permission
  const role = await getMemberRole(installationId, session.user.id)
  const installation = await getInstallationById(installationId)

  const isOwner = role === 'owner' || installation?.userId === session.user.id
  if (!isOwner) {
    throw new Error('Only the installation owner can manage billing')
  }

  // Check no existing active subscription
  const existing = await getSubscriptionByInstallation(installationId)
  if (existing && !['cancelled', 'expired'].includes(existing.status)) {
    throw new Error('Installation already has an active subscription')
  }

  // Ensure plans are synced to Razorpay
  await syncPlansToRazorpay()

  const plan = await getPlanById(planId)
  if (!plan || !plan.razorpayPlanId) {
    throw new Error('Plan not found or not synced to Razorpay')
  }

  const razorpay = getRazorpay()

  // Create Razorpay customer
  const customer = await razorpay.customers.create({
    name: session.user.name,
    email: session.user.email,
  })

  // Create Razorpay subscription
  const razorpaySub = await razorpay.subscriptions.create({
    plan_id: plan.razorpayPlanId,
    total_count: 120,
    quantity: quantity,
    customer_id: customer.id,
    notes: {
      installation_id: installationId,
      plan_tier: String(plan.tier),
    },
  })

  // Save locally
  const subscription = await createSubscription({
    installationId,
    planId,
    razorpaySubscriptionId: razorpaySub.id,
    razorpayCustomerId: customer.id,
    quantity,
  })

  return {
    subscriptionId: subscription.id,
    razorpaySubscriptionId: razorpaySub.id,
  }
}

export async function cancelExistingSubscription(installationId: string): Promise<void> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  // Verify owner permission
  const role = await getMemberRole(installationId, session.user.id)
  const installation = await getInstallationById(installationId)
  const isOwner = role === 'owner' || installation?.userId === session.user.id
  if (!isOwner) {
    throw new Error('Only the installation owner can manage billing')
  }

  const subscription = await getSubscriptionByInstallation(installationId)
  if (!subscription || !subscription.razorpaySubscriptionId) {
    throw new Error('No active subscription found')
  }

  if (['cancelled', 'expired'].includes(subscription.status)) {
    throw new Error('Subscription is already cancelled')
  }

  const razorpay = getRazorpay()
  await razorpay.subscriptions.cancel(subscription.razorpaySubscriptionId, false)
}
