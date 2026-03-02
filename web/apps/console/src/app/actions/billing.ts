'use server'

import crypto from 'crypto'
import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getRazorpay } from '@/lib/razorpay'
import {
  getPlans,
  getPlanById,
  getSubscriptionsByInstallation,
  createSubscription,
  activateSubscriptionsByInstallation,
  cancelSubscriptionsByInstallation,
  deleteCreatedSubscriptions,
  getInvoicesByInstallation,
  upsertInvoice,
  extendSubscriptionPeriod,
  updateInvoiceStatus,
  scheduleRenewalJobs,
  cancelRenewalJobs,
} from '@/lib/console/storage'
import type { Plan, Subscription, Invoice } from '@/lib/console/storage'
import { getInstallationById, getMemberRole } from '@/lib/console/storage'

// --- Read Actions ---

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

// --- Write Actions ---

export async function createInstallationOrder(
  installationId: string,
  allocations: { planId: string; quantity: number }[],
  billingPeriod: 'monthly' | 'annual' = 'monthly',
): Promise<{ razorpayOrderId: string; amount: number; currency: string }> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  // Verify owner permission
  const role = await getMemberRole(installationId, session.user.id)
  const installation = await getInstallationById(installationId)
  const isOwner = role === 'owner' || installation?.userId === session.user.id
  if (!isOwner) {
    throw new Error('Only the installation owner can manage billing')
  }

  if (allocations.length === 0) {
    throw new Error('At least one compute size must have users assigned')
  }

  // Check existing subscriptions
  const existing = await getSubscriptionsByInstallation(installationId)
  const activeOrPaid = existing.filter((s) => ['active', 'authenticated', 'paused'].includes(s.status))
  if (activeOrPaid.length > 0) {
    throw new Error('Installation already has active subscriptions')
  }

  // Delete stale 'created' subscriptions from previous incomplete checkout attempts
  // Uses DELETE (not soft-cancel) to free the UNIQUE(installation_id, plan_id) constraint
  const staleCreated = existing.filter((s) => s.status === 'created')
  if (staleCreated.length > 0) {
    await deleteCreatedSubscriptions(installationId)
  }

  // Resolve all plans and compute total amount
  const resolvedPlans = await Promise.all(
    allocations.map(async (alloc) => {
      const plan = await getPlanById(alloc.planId)
      if (!plan) {
        throw new Error(`Plan not found: ${alloc.planId}`)
      }
      return { plan, quantity: alloc.quantity }
    }),
  )

  // Compute total: baseFee (once) + sum(amountPerUser * quantity)
  const baseFee = resolvedPlans[0].plan.baseFee
  const currency = resolvedPlans[0].plan.currency
  const discountPct = resolvedPlans[0].plan.annualDiscountPct
  const userTotal = resolvedPlans.reduce((sum, { plan, quantity }) => {
    return sum + plan.amountPerUser * quantity
  }, 0)
  const monthlyTotal = baseFee + userTotal

  // Annual: monthly * 12 * (1 - discount/100)
  const totalAmount =
    billingPeriod === 'annual'
      ? Math.round(monthlyTotal * 12 * (1 - discountPct / 100))
      : monthlyTotal

  const razorpay = getRazorpay()

  // Create Razorpay Order
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const order = (await razorpay.orders.create({
    amount: totalAmount,
    currency,
    receipt: `inst_${installationId.slice(0, 8)}_${Date.now()}`,
    notes: {
      installation_id: installationId,
      allocations: JSON.stringify(allocations),
    },
  })) as any

  // Create local subscription records (status: 'created')
  // Only store order_id on the first record (UNIQUE constraint on razorpay_subscription_id)
  let isFirst = true
  for (const { plan, quantity } of resolvedPlans) {
    await createSubscription({
      installationId,
      planId: plan.id,
      razorpaySubscriptionId: isFirst ? order.id : null,
      razorpayCustomerId: null,
      quantity,
      billingPeriod,
    })
    isFirst = false
  }

  return {
    razorpayOrderId: order.id,
    amount: totalAmount,
    currency,
  }
}

export async function verifyPaymentAndActivate(
  installationId: string,
  razorpayOrderId: string,
  razorpayPaymentId: string,
  razorpaySignature: string,
): Promise<void> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  // Verify owner permission
  const role = await getMemberRole(installationId, session.user.id)
  const installation = await getInstallationById(installationId)
  const isOwner = role === 'owner' || installation?.userId === session.user.id
  if (!isOwner) {
    throw new Error('Only the installation owner can verify payments')
  }

  // Verify Razorpay payment signature
  const keySecret = process.env.RAZORPAY_KEY_SECRET
  if (!keySecret) {
    throw new Error('RAZORPAY_KEY_SECRET not configured')
  }

  const expectedSignature = crypto
    .createHmac('sha256', keySecret)
    .update(`${razorpayOrderId}|${razorpayPaymentId}`)
    .digest('hex')

  try {
    const isValid = crypto.timingSafeEqual(
      Buffer.from(expectedSignature, 'hex'),
      Buffer.from(razorpaySignature, 'hex'),
    )
    if (!isValid) {
      throw new Error('Payment verification failed — invalid signature')
    }
  } catch {
    throw new Error('Payment verification failed — invalid signature')
  }

  // Verify the order belongs to this installation via Razorpay API
  const razorpay = getRazorpay()
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const order = (await razorpay.orders.fetch(razorpayOrderId)) as any
  if (order.notes?.installation_id !== installationId) {
    throw new Error('Payment verification failed — order does not belong to this installation')
  }
  if (order.status !== 'paid') {
    throw new Error('Payment verification failed — order is not paid')
  }

  // Check if this order was already verified (idempotency guard)
  const subs = await getSubscriptionsByInstallation(installationId)
  const alreadyLinkedSub = subs.find((s) => s.razorpaySubscriptionId === razorpayOrderId)
  if (alreadyLinkedSub && alreadyLinkedSub.status === 'active') {
    return // Already activated — idempotent no-op
  }

  const hasActiveSubs = subs.some((s) => s.status === 'active')

  // Determine billing period from existing subscriptions
  const billingPeriod = subs[0]?.billingPeriod ?? 'monthly'
  const periodDays = billingPeriod === 'annual' ? 365 : 30

  const now = new Date().toISOString()
  const periodEnd = new Date(Date.now() + periodDays * 24 * 60 * 60 * 1000).toISOString()

  if (hasActiveSubs) {
    // Renewal: extend existing active subscriptions
    await extendSubscriptionPeriod(installationId, now, periodEnd)

    // Update the issued invoice to paid
    await updateInvoiceStatus(razorpayOrderId, 'paid', razorpayPaymentId, now)

    // Cancel old pending jobs and schedule new ones for the extended period
    await cancelRenewalJobs(installationId)
    await scheduleRenewalJobs(installationId, periodEnd, billingPeriod)
  } else {
    // First-time: activate 'created' subscriptions
    await activateSubscriptionsByInstallation(installationId, now, periodEnd)

    // Create invoice record (order already fetched above)
    const firstSub = subs[0]
    if (firstSub) {
      await upsertInvoice({
        subscriptionId: firstSub.id,
        installationId,
        razorpayInvoiceId: razorpayOrderId,
        razorpayPaymentId,
        amount: order.amount,
        currency: order.currency,
        status: 'paid',
        billingStart: now,
        billingEnd: periodEnd,
        paidAt: now,
      })
    }

    // Schedule renewal and expiry jobs for this subscription
    await scheduleRenewalJobs(installationId, periodEnd, billingPeriod)
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

  const subs = await getSubscriptionsByInstallation(installationId)
  const activeSubs = subs.filter(
    (s) => !['cancelled', 'expired'].includes(s.status),
  )

  if (activeSubs.length === 0) {
    throw new Error('No active subscriptions found')
  }

  // Cancel locally (Orders don't have Razorpay-side subscriptions to cancel)
  await cancelSubscriptionsByInstallation(installationId)

  // Cancel any pending renewal/expiry jobs
  await cancelRenewalJobs(installationId)
}
