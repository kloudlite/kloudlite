'use server'

import crypto from 'crypto'
import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getRazorpay } from '@/lib/razorpay'
import type { RazorpayOrder } from '@/lib/razorpay-types'
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
  updateSubscriptionPeriod,
  upsertActiveSubscription,
  setScheduledBillingPeriod,
  clearScheduledBillingPeriod,
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
  const order = (await razorpay.orders.create({
    amount: totalAmount,
    currency,
    receipt: `inst_${installationId.slice(0, 8)}_${Date.now()}`,
    notes: {
      installation_id: installationId,
      allocations: JSON.stringify(allocations),
    },
  })) as unknown as RazorpayOrder

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
  const order = (await razorpay.orders.fetch(razorpayOrderId)) as unknown as RazorpayOrder
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

    // Check if a scheduled billing period switch needs to be applied at renewal
    const scheduledPeriod = subs.find((s) => s.status === 'active')?.scheduledBillingPeriod
    let effectivePeriod = billingPeriod
    let effectivePeriodEnd = periodEnd

    if (scheduledPeriod && scheduledPeriod !== billingPeriod) {
      // Apply the scheduled period switch (e.g., annual → monthly)
      const newPeriodDays = scheduledPeriod === 'annual' ? 365 : 30
      effectivePeriodEnd = new Date(Date.now() + newPeriodDays * 24 * 60 * 60 * 1000).toISOString()
      effectivePeriod = scheduledPeriod

      const activeSubs = subs.filter((s) => s.status === 'active')
      for (const sub of activeSubs) {
        await updateSubscriptionPeriod(sub.id, scheduledPeriod, effectivePeriodEnd)
      }
      await clearScheduledBillingPeriod(installationId)
    }

    // Post-activation bookkeeping — failures here must not invalidate the payment
    // (subscription is already extended, money was charged)
    try {
      await updateInvoiceStatus(razorpayOrderId, 'paid', razorpayPaymentId, now)
      await cancelRenewalJobs(installationId)
      await scheduleRenewalJobs(installationId, effectivePeriodEnd, effectivePeriod)
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
  newBillingPeriod?: 'monthly' | 'annual',
): { proratedAmount: number; remainingDays: number; oldMonthlyTotal: number; newMonthlyTotal: number; newCurrentEnd: string | null } {
  const now = Date.now()
  const currentEnd = activeSubs[0]?.currentEnd
  if (!currentEnd) {
    return { proratedAmount: 0, remainingDays: 0, oldMonthlyTotal: 0, newMonthlyTotal: 0, newCurrentEnd: null }
  }

  const endMs = new Date(currentEnd).getTime()
  const remainingDays = Math.max(0, Math.ceil((endMs - now) / (24 * 60 * 60 * 1000)))

  const planMap = new Map(plans.map((p) => [p.id, p]))
  const baseFee = plans[0]?.baseFee ?? 0
  const discountPct = planMap.values().next().value?.annualDiscountPct ?? 20

  const oldBillingPeriod = activeSubs[0]?.billingPeriod ?? 'monthly'
  const effectiveNewPeriod = newBillingPeriod ?? oldBillingPeriod

  const oldUserTotal = activeSubs.reduce((sum, sub) => {
    const plan = planMap.get(sub.planId)
    return sum + (plan ? plan.amountPerUser * sub.quantity : 0)
  }, 0)

  const newUserTotal = newAllocations.reduce((sum, alloc) => {
    const plan = planMap.get(alloc.planId)
    return sum + (plan ? plan.amountPerUser * alloc.quantity : 0)
  }, 0)

  // Compute period totals (baseFee + user costs), adjusted for annual discount
  const oldPeriodTotal = oldBillingPeriod === 'annual'
    ? (baseFee + oldUserTotal) * 12 * (1 - discountPct / 100)
    : baseFee + oldUserTotal
  const newPeriodTotal = effectiveNewPeriod === 'annual'
    ? (baseFee + newUserTotal) * 12 * (1 - discountPct / 100)
    : baseFee + newUserTotal

  const oldPeriodDays = oldBillingPeriod === 'annual' ? 365 : 30
  const newPeriodDays = effectiveNewPeriod === 'annual' ? 365 : 30

  const oldDailyRate = oldPeriodTotal / oldPeriodDays
  const newDailyRate = newPeriodTotal / newPeriodDays

  let proratedAmount: number
  let newCurrentEnd: string | null = null

  if (effectiveNewPeriod !== oldBillingPeriod) {
    // Period changing — "new full period cost minus remaining credit"
    const remainingCredit = oldDailyRate * remainingDays

    // Only Monthly → Annual is prorated mid-cycle; Annual → Monthly is scheduled for period end
    proratedAmount = Math.round(newPeriodTotal - remainingCredit)
    newCurrentEnd = new Date(now + 365 * 24 * 60 * 60 * 1000).toISOString()
  } else {
    // Same period — quantity-only change, daily rate diff (existing logic)
    proratedAmount = Math.round((newDailyRate - oldDailyRate) * remainingDays)
  }

  return { proratedAmount, remainingDays, oldMonthlyTotal: oldUserTotal, newMonthlyTotal: newUserTotal, newCurrentEnd }
}

async function applySubscriptionChanges(
  installationId: string,
  activeSubs: Subscription[],
  newAllocations: { planId: string; quantity: number }[],
  newBillingPeriod?: 'monthly' | 'annual',
  newCurrentEnd?: string | null,
): Promise<void> {
  const allocMap = new Map(newAllocations.map((a) => [a.planId, a.quantity]))
  const oldBillingPeriod = activeSubs[0]?.billingPeriod ?? 'monthly'
  const periodChanged = newBillingPeriod && newBillingPeriod !== oldBillingPeriod

  // Update existing active subs
  for (const sub of activeSubs) {
    const newQty = allocMap.get(sub.planId) ?? 0
    if (newQty !== sub.quantity) {
      await updateSubscriptionQuantity(sub.id, newQty)
    }
    if (periodChanged && newCurrentEnd) {
      await updateSubscriptionPeriod(sub.id, newBillingPeriod, newCurrentEnd)
    }
    allocMap.delete(sub.planId)
  }

  // Upsert new tiers not currently active
  const effectivePeriod = newBillingPeriod ?? oldBillingPeriod
  const refSub = activeSubs[0]
  for (const [planId, quantity] of allocMap) {
    if (quantity > 0) {
      await upsertActiveSubscription({
        installationId,
        planId,
        quantity,
        billingPeriod: effectivePeriod,
        currentStart: refSub.currentStart!,
        currentEnd: (periodChanged && newCurrentEnd) ? newCurrentEnd : refSub.currentEnd!,
      })
    }
  }

  // Reschedule renewal jobs if period changed
  if (periodChanged && newCurrentEnd) {
    await cancelRenewalJobs(installationId)
    await scheduleRenewalJobs(installationId, newCurrentEnd, newBillingPeriod)
  }
}

export async function previewModification(
  installationId: string,
  newAllocations: { planId: string; quantity: number }[],
  newBillingPeriod?: 'monthly' | 'annual',
): Promise<{
  proratedAmount: number
  isUpgrade: boolean
  remainingDays: number
  oldMonthlyTotal: number
  newMonthlyTotal: number
  newCurrentEnd: string | null
  scheduledDowngrade: boolean
}> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const subs = await getSubscriptionsByInstallation(installationId)
  const activeSubs = subs.filter((s) => s.status === 'active')
  if (activeSubs.length === 0) {
    throw new Error('No active subscriptions found')
  }

  const oldBillingPeriod = activeSubs[0].billingPeriod
  const isAnnualToMonthly = newBillingPeriod === 'monthly' && oldBillingPeriod === 'annual'

  if (isAnnualToMonthly) {
    // Annual → Monthly is scheduled, not applied mid-cycle
    const plans = await getPlans()
    const planMap = new Map(plans.map((p) => [p.id, p]))
    const oldUserTotal = activeSubs.reduce((sum, sub) => {
      const plan = planMap.get(sub.planId)
      return sum + (plan ? plan.amountPerUser * sub.quantity : 0)
    }, 0)
    const newUserTotal = newAllocations.reduce((sum, alloc) => {
      const plan = planMap.get(alloc.planId)
      return sum + (plan ? plan.amountPerUser * alloc.quantity : 0)
    }, 0)
    return {
      proratedAmount: 0,
      isUpgrade: false,
      remainingDays: 0,
      oldMonthlyTotal: oldUserTotal,
      newMonthlyTotal: newUserTotal,
      newCurrentEnd: null,
      scheduledDowngrade: true,
    }
  }

  const plans = await getPlans()
  const result = computeProration(activeSubs, newAllocations, plans, newBillingPeriod)

  return {
    ...result,
    isUpgrade: result.proratedAmount > 0,
    scheduledDowngrade: false,
  }
}

export async function modifySubscriptionQuantities(
  installationId: string,
  newAllocations: { planId: string; quantity: number }[],
  newBillingPeriod?: 'monthly' | 'annual',
): Promise<
  | { applied: true; scheduled?: boolean }
  | { applied: false; razorpayOrderId: string; amount: number; currency: string }
> {
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

  const oldBillingPeriod = activeSubs[0].billingPeriod
  const isAnnualToMonthly = newBillingPeriod === 'monthly' && oldBillingPeriod === 'annual'

  if (isAnnualToMonthly) {
    // Annual → Monthly: schedule for end of current period, don't apply now
    // If there are also quantity changes, apply those within the current annual period
    const hasQuantityChanges = newAllocations.some((alloc) => {
      const existing = activeSubs.find((s) => s.planId === alloc.planId)
      return !existing || existing.quantity !== alloc.quantity
    })

    if (hasQuantityChanges) {
      // Apply quantity changes within current annual period (no period change)
      const plans = await getPlans()
      const { proratedAmount, newCurrentEnd } = computeProration(activeSubs, newAllocations, plans)
      if (proratedAmount > 0 && proratedAmount >= 100) {
        // Quantity upgrade within annual — needs payment
        const currency = plans[0]?.currency ?? 'INR'
        const razorpay = getRazorpay()
        const order = (await razorpay.orders.create({
          amount: proratedAmount,
          currency,
          receipt: `mod_${installationId.slice(0, 8)}_${Date.now()}`,
          notes: {
            installation_id: installationId,
            modification: 'true',
            allocations: JSON.stringify(newAllocations),
            schedule_monthly: 'true',
            ...(newCurrentEnd ? { new_current_end: newCurrentEnd } : {}),
          },
        })) as unknown as RazorpayOrder

        return {
          applied: false,
          razorpayOrderId: order.id,
          amount: proratedAmount,
          currency,
        }
      }
      await applySubscriptionChanges(installationId, activeSubs, newAllocations, undefined, newCurrentEnd)
    }

    // Schedule the billing period switch for end of current period
    for (const sub of activeSubs) {
      await setScheduledBillingPeriod(sub.id, 'monthly')
    }

    return { applied: true, scheduled: true }
  }

  const plans = await getPlans()
  const { proratedAmount, newCurrentEnd } = computeProration(activeSubs, newAllocations, plans, newBillingPeriod)

  // Downgrade or no cost change — apply immediately
  if (proratedAmount <= 0 || proratedAmount < 100) {
    await applySubscriptionChanges(installationId, activeSubs, newAllocations, newBillingPeriod, newCurrentEnd)
    return { applied: true }
  }

  // Upgrade — create Razorpay order for prorated amount
  const currency = plans[0]?.currency ?? 'INR'
  const razorpay = getRazorpay()

  const order = (await razorpay.orders.create({
    amount: proratedAmount,
    currency,
    receipt: `mod_${installationId.slice(0, 8)}_${Date.now()}`,
    notes: {
      installation_id: installationId,
      modification: 'true',
      allocations: JSON.stringify(newAllocations),
      ...(newBillingPeriod ? { billing_period: newBillingPeriod } : {}),
      ...(newCurrentEnd ? { new_current_end: newCurrentEnd } : {}),
    },
  })) as unknown as RazorpayOrder

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
  const order = (await razorpay.orders.fetch(razorpayOrderId)) as unknown as RazorpayOrder
  if (order.notes?.installation_id !== installationId) {
    throw new Error('Payment verification failed — order does not belong to this installation')
  }
  if (order.notes?.modification !== 'true') {
    throw new Error('Payment verification failed — order is not a modification')
  }

  // Parse allocations and optional period change from order notes
  const newAllocations = JSON.parse(order.notes.allocations) as { planId: string; quantity: number }[]
  const newBillingPeriod = order.notes.billing_period as 'monthly' | 'annual' | undefined
  const newCurrentEnd = order.notes.new_current_end as string | undefined

  const subs = await getSubscriptionsByInstallation(installationId)
  const activeSubs = subs.filter((s) => s.status === 'active')
  if (activeSubs.length === 0) {
    throw new Error('No active subscriptions to modify')
  }

  await applySubscriptionChanges(installationId, activeSubs, newAllocations, newBillingPeriod, newCurrentEnd ?? null)

  // If this modification also schedules a monthly downgrade, apply it now
  if (order.notes?.schedule_monthly === 'true') {
    const updatedSubs = await getSubscriptionsByInstallation(installationId)
    const updatedActive = updatedSubs.filter((s) => s.status === 'active')
    for (const sub of updatedActive) {
      await setScheduledBillingPeriod(sub.id, 'monthly')
    }
  }

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

export async function cancelScheduledDowngrade(installationId: string): Promise<void> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  // Verify owner permission
  const role = await getMemberRole(installationId, session.user.id)
  const installation = await getInstallationById(installationId)
  const isOwner = role === 'owner' || installation?.userId === session.user.id
  if (!isOwner) {
    throw new Error('Only the installation owner can manage billing')
  }

  await clearScheduledBillingPeriod(installationId)
}
