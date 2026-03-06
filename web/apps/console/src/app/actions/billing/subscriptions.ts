'use server'

import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getRazorpay } from '@/lib/razorpay'
import type { RazorpayOrder } from '@/lib/razorpay-types'
import {
  getPlans,
  getPlanById,
  getSubscriptionsByInstallation,
  createSubscription,
  deleteCreatedSubscriptions,
  cancelSubscriptionsByInstallation,
  cancelRenewalJobs,
  setScheduledBillingPeriod,
  clearScheduledBillingPeriod,
  getInstallationById,
  getMemberRole,
} from '@/lib/console/storage'
import { computeProration, applySubscriptionChanges } from './proration'

type Allocation = { planId: string; quantity: number }

function validateAllocations(
  allocations: Allocation[],
  options: { requirePositiveTotal: boolean } = { requirePositiveTotal: true },
): void {
  if (!Array.isArray(allocations) || allocations.length === 0) {
    throw new Error('At least one compute size must be provided')
  }

  const seen = new Set<string>()
  let total = 0

  for (const alloc of allocations) {
    if (!alloc?.planId || typeof alloc.planId !== 'string') {
      throw new Error('Each allocation must include a valid planId')
    }
    if (seen.has(alloc.planId)) {
      throw new Error(`Duplicate allocation for plan: ${alloc.planId}`)
    }
    seen.add(alloc.planId)

    if (!Number.isInteger(alloc.quantity) || alloc.quantity < 0) {
      throw new Error('Allocation quantity must be a non-negative integer')
    }
    total += alloc.quantity
  }

  if (options.requirePositiveTotal && total <= 0) {
    throw new Error('At least one user must be assigned')
  }
}

export async function createInstallationOrder(
  installationId: string,
  allocations: Allocation[],
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

  validateAllocations(allocations)

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

export async function previewModification(
  installationId: string,
  newAllocations: Allocation[],
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

  validateAllocations(newAllocations)

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
  newAllocations: Allocation[],
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

  validateAllocations(newAllocations)

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
