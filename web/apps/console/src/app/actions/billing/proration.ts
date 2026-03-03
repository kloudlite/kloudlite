import {
  updateSubscriptionQuantity,
  updateSubscriptionPeriod,
  upsertActiveSubscription,
  cancelRenewalJobs,
  scheduleRenewalJobs,
} from '@/lib/console/storage'
import type { Plan, Subscription } from '@/lib/console/storage'

export function computeProration(
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

export async function applySubscriptionChanges(
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
