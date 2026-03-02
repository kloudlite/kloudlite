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
  updateSubscriptionQuantity,
  upsertActiveSubscription,
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

  // Verify Razorpay payment signature (proves Razorpay signed order_id|payment_id with our secret)
  const keySecret = process.env.RAZORPAY_KEY_SECRET
  if (!keySecret) {
    throw new Error('RAZORPAY_KEY_SECRET not configured')
  }

  const expectedSignature = crypto
    .createHmac('sha256', keySecret)
    .update(`${razorpayOrderId}|${razorpayPaymentId}`)
    .digest('hex')

  const sigValid =
    expectedSignature.length === razorpaySignature.length &&
    crypto.timingSafeEqual(Buffer.from(expectedSignature), Buffer.from(razorpaySignature))

  if (!sigValid) {
    console.error('[Billing] Signature mismatch for order', razorpayOrderId)
    throw new Error('Payment verification failed — invalid signature')
  }

  // Verify the order belongs to this installation via Razorpay API
  const razorpay = getRazorpay()
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const order = (await razorpay.orders.fetch(razorpayOrderId)) as any
  if (order.notes?.installation_id !== installationId) {
    console.error('[Billing] Order installation mismatch:', order.notes?.installation_id, '!==', installationId)
    throw new Error('Payment verification failed — order does not belong to this installation')
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

    // Post-activation bookkeeping — failures here must not invalidate the payment
    // (subscription is already extended, money was charged)
    try {
      await updateInvoiceStatus(razorpayOrderId, 'paid', razorpayPaymentId, now)
      await cancelRenewalJobs(installationId)
      await scheduleRenewalJobs(installationId, periodEnd, billingPeriod)
    } catch (bookkeepingErr) {
      console.error('[Billing] Post-renewal bookkeeping failed (subscription still active):', bookkeepingErr)
    }
  } else {
    // First-time: activate 'created' subscriptions
    await activateSubscriptionsByInstallation(installationId, now, periodEnd)

    // Post-activation bookkeeping — failures here must not invalidate the payment
    // (subscription is already active, money was charged)
    try {
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
      await scheduleRenewalJobs(installationId, periodEnd, billingPeriod)
    } catch (bookkeepingErr) {
      console.error('[Billing] Post-activation bookkeeping failed (subscription still active):', bookkeepingErr)
    }
  }
}

// --- Subscription Modification Actions ---

function computeProration(
  activeSubs: { planId: string; quantity: number; billingPeriod: 'monthly' | 'annual'; currentEnd: string | null }[],
  newAllocations: { planId: string; quantity: number }[],
  plans: Plan[],
): { proratedAmount: number; remainingDays: number; oldMonthlyTotal: number; newMonthlyTotal: number } {
  const now = Date.now()
  const currentEnd = activeSubs[0]?.currentEnd
  if (!currentEnd) {
    return { proratedAmount: 0, remainingDays: 0, oldMonthlyTotal: 0, newMonthlyTotal: 0 }
  }

  const endMs = new Date(currentEnd).getTime()
  const remainingDays = Math.max(0, Math.ceil((endMs - now) / (24 * 60 * 60 * 1000)))

  const planMap = new Map(plans.map((p) => [p.id, p]))

  const oldUserTotal = activeSubs.reduce((sum, sub) => {
    const plan = planMap.get(sub.planId)
    return sum + (plan ? plan.amountPerUser * sub.quantity : 0)
  }, 0)

  const newUserTotal = newAllocations.reduce((sum, alloc) => {
    const plan = planMap.get(alloc.planId)
    return sum + (plan ? plan.amountPerUser * alloc.quantity : 0)
  }, 0)

  const monthlyDiff = newUserTotal - oldUserTotal
  const billingPeriod = activeSubs[0]?.billingPeriod ?? 'monthly'
  const discountPct = planMap.values().next().value?.annualDiscountPct ?? 20

  let proratedAmount: number
  if (billingPeriod === 'annual') {
    proratedAmount = Math.round(monthlyDiff * 12 * (1 - discountPct / 100) * remainingDays / 365)
  } else {
    proratedAmount = Math.round(monthlyDiff * remainingDays / 30)
  }

  return { proratedAmount, remainingDays, oldMonthlyTotal: oldUserTotal, newMonthlyTotal: newUserTotal }
}

async function applyQuantityChanges(
  installationId: string,
  activeSubs: Subscription[],
  newAllocations: { planId: string; quantity: number }[],
): Promise<void> {
  const allocMap = new Map(newAllocations.map((a) => [a.planId, a.quantity]))

  // Update existing active subs
  for (const sub of activeSubs) {
    const newQty = allocMap.get(sub.planId) ?? 0
    if (newQty !== sub.quantity) {
      await updateSubscriptionQuantity(sub.id, newQty)
    }
    allocMap.delete(sub.planId)
  }

  // Upsert new tiers not currently active
  for (const [planId, quantity] of allocMap) {
    if (quantity > 0) {
      const refSub = activeSubs[0]
      await upsertActiveSubscription({
        installationId,
        planId,
        quantity,
        billingPeriod: refSub.billingPeriod,
        currentStart: refSub.currentStart!,
        currentEnd: refSub.currentEnd!,
      })
    }
  }
}

export async function previewModification(
  installationId: string,
  newAllocations: { planId: string; quantity: number }[],
): Promise<{
  proratedAmount: number
  isUpgrade: boolean
  remainingDays: number
  oldMonthlyTotal: number
  newMonthlyTotal: number
}> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const subs = await getSubscriptionsByInstallation(installationId)
  const activeSubs = subs.filter((s) => s.status === 'active')
  if (activeSubs.length === 0) {
    throw new Error('No active subscriptions found')
  }

  const plans = await getPlans()
  const result = computeProration(activeSubs, newAllocations, plans)

  return {
    ...result,
    isUpgrade: result.proratedAmount > 0,
  }
}

export async function modifySubscriptionQuantities(
  installationId: string,
  newAllocations: { planId: string; quantity: number }[],
): Promise<{ applied: true } | { applied: false; razorpayOrderId: string; amount: number; currency: string }> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  // Verify owner permission
  const role = await getMemberRole(installationId, session.user.id)
  const installation = await getInstallationById(installationId)
  const isOwner = role === 'owner' || installation?.userId === session.user.id
  if (!isOwner) {
    throw new Error('Only the installation owner can manage billing')
  }

  const totalNewUsers = newAllocations.reduce((sum, a) => sum + a.quantity, 0)
  if (totalNewUsers === 0) {
    throw new Error('At least one user must be assigned')
  }

  const subs = await getSubscriptionsByInstallation(installationId)
  const activeSubs = subs.filter((s) => s.status === 'active')
  if (activeSubs.length === 0) {
    throw new Error('No active subscriptions to modify')
  }

  const plans = await getPlans()
  const { proratedAmount } = computeProration(activeSubs, newAllocations, plans)

  // Downgrade or no cost change — apply immediately
  if (proratedAmount <= 0 || proratedAmount < 100) {
    await applyQuantityChanges(installationId, activeSubs, newAllocations)
    return { applied: true }
  }

  // Upgrade — create Razorpay order for prorated amount
  const currency = plans[0]?.currency ?? 'INR'
  const razorpay = getRazorpay()

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const order = (await razorpay.orders.create({
    amount: proratedAmount,
    currency,
    receipt: `mod_${installationId.slice(0, 8)}_${Date.now()}`,
    notes: {
      installation_id: installationId,
      modification: 'true',
      allocations: JSON.stringify(newAllocations),
    },
  })) as any

  return {
    applied: false,
    razorpayOrderId: order.id,
    amount: proratedAmount,
    currency,
  }
}

export async function verifyModificationAndApply(
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

  const sigValid =
    expectedSignature.length === razorpaySignature.length &&
    crypto.timingSafeEqual(Buffer.from(expectedSignature), Buffer.from(razorpaySignature))

  if (!sigValid) {
    console.error('[Billing] Modification signature mismatch for order', razorpayOrderId)
    throw new Error('Payment verification failed — invalid signature')
  }

  // Verify the order belongs to this installation and is a modification
  const razorpay = getRazorpay()
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const order = (await razorpay.orders.fetch(razorpayOrderId)) as any
  if (order.notes?.installation_id !== installationId) {
    throw new Error('Payment verification failed — order does not belong to this installation')
  }
  if (order.notes?.modification !== 'true') {
    throw new Error('Payment verification failed — order is not a modification')
  }

  // Parse allocations from order notes
  const newAllocations = JSON.parse(order.notes.allocations) as { planId: string; quantity: number }[]

  const subs = await getSubscriptionsByInstallation(installationId)
  const activeSubs = subs.filter((s) => s.status === 'active')
  if (activeSubs.length === 0) {
    throw new Error('No active subscriptions to modify')
  }

  await applyQuantityChanges(installationId, activeSubs, newAllocations)

  // Create invoice for prorated amount (bookkeeping — don't fail the modification)
  try {
    const firstSub = activeSubs[0]
    await upsertInvoice({
      subscriptionId: firstSub.id,
      installationId,
      razorpayInvoiceId: razorpayOrderId,
      razorpayPaymentId,
      amount: order.amount,
      currency: order.currency,
      status: 'paid',
      billingStart: new Date().toISOString(),
      billingEnd: firstSub.currentEnd ?? undefined,
      paidAt: new Date().toISOString(),
    })
  } catch (bookkeepingErr) {
    console.error('[Billing] Post-modification bookkeeping failed:', bookkeepingErr)
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
