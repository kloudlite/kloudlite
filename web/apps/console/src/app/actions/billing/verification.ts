'use server'

import crypto from 'crypto'
import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getRazorpay } from '@/lib/razorpay'
import type { RazorpayOrder } from '@/lib/razorpay-types'
import {
  getSubscriptionsByInstallation,
  activateSubscriptionsByInstallation,
  extendSubscriptionPeriod,
  updateInvoiceStatus,
  upsertInvoice,
  scheduleRenewalJobs,
  cancelRenewalJobs,
  updateSubscriptionPeriod,
  clearScheduledBillingPeriod,
  setScheduledBillingPeriod,
} from '@/lib/console/storage'
import { getInstallationById, getMemberRole } from '@/lib/console/storage'
import { applySubscriptionChanges } from './proration'

function verifyRazorpaySignature(
  razorpayOrderId: string,
  razorpayPaymentId: string,
  razorpaySignature: string,
): void {
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

  verifyRazorpaySignature(razorpayOrderId, razorpayPaymentId, razorpaySignature)

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

  verifyRazorpaySignature(razorpayOrderId, razorpayPaymentId, razorpaySignature)

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
