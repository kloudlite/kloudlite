'use client'

import { useCallback, useState } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import {
  getRazorpayKey,
  createInstallationOrder,
  verifyPaymentAndActivate,
  modifySubscriptionQuantities,
  verifyModificationAndApply,
  cancelScheduledDowngrade,
} from '@/app/actions/billing'
import { useRazorpay } from '@/components/razorpay-provider'
import type { Subscription, Invoice } from '@/lib/console/storage'

interface UseSubscriptionPaymentsOptions {
  installationId: string
  userEmail: string
  userName: string
  activeSubs: Subscription[]
  pendingInvoice: Invoice | undefined
}

export function useSubscriptionPayments({
  installationId,
  userEmail,
  userName,
  activeSubs,
  pendingInvoice,
}: UseSubscriptionPaymentsOptions) {
  const router = useRouter()
  const { openCheckout } = useRazorpay()
  const [paying, setPaying] = useState(false)
  const [editing, setEditing] = useState(false)

  const handleCancelScheduledDowngrade = useCallback(async () => {
    try {
      await cancelScheduledDowngrade(installationId)
      toast.success('Scheduled downgrade cancelled.')
      router.refresh()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to cancel scheduled downgrade')
    }
  }, [installationId, router])

  const handleModify = useCallback(
    async (allocations: { planId: string; quantity: number }[], billingPeriod: 'monthly' | 'annual') => {
      try {
        const currentPeriod = activeSubs[0]?.billingPeriod ?? 'monthly'
        const newPeriod = billingPeriod !== currentPeriod ? billingPeriod : undefined
        const result = await modifySubscriptionQuantities(installationId, allocations, newPeriod)

        if (result.applied) {
          const message = 'scheduled' in result && result.scheduled
            ? 'Monthly billing scheduled for end of current period.'
            : 'Subscription updated successfully.'
          toast.success(message)
          setEditing(false)
          router.refresh()
          return
        }

        const key = await getRazorpayKey()
        const options = {
          key,
          order_id: result.razorpayOrderId,
          amount: result.amount,
          currency: result.currency,
          name: 'Kloudlite',
          description: 'Subscription Upgrade (prorated)',
          prefill: { email: userEmail, name: userName },
          theme: { color: '#3B82F6' },
          handler: async (response: Record<string, string>) => {
            try {
              await verifyModificationAndApply(
                installationId,
                response.razorpay_order_id,
                response.razorpay_payment_id,
                response.razorpay_signature,
              )
              toast.success('Upgrade successful! Quantities updated.')
              setEditing(false)
              router.refresh()
            } catch {
              toast.error('Payment verification failed. Please contact support.')
            }
          },
          modal: {
            ondismiss: () => {
              toast.info('Payment cancelled. No changes were made.')
            },
          },
        }
        openCheckout(options)
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to modify subscription')
      }
    },
    [installationId, userEmail, userName, router, openCheckout, activeSubs],
  )

  const handlePayNow = useCallback(async () => {
    if (!pendingInvoice?.razorpayInvoiceId || paying) return
    setPaying(true)

    const key = await getRazorpayKey()
    const options = {
      key,
      order_id: pendingInvoice.razorpayInvoiceId,
      amount: pendingInvoice.amount,
      currency: pendingInvoice.currency,
      name: 'Kloudlite',
      description: 'Subscription Renewal',
      prefill: { email: userEmail, name: userName },
      theme: { color: '#3B82F6' },
      handler: async (response: Record<string, string>) => {
        try {
          await verifyPaymentAndActivate(
            installationId,
            response.razorpay_order_id,
            response.razorpay_payment_id,
            response.razorpay_signature,
          )
          toast.success('Payment successful! Subscription renewed.')
          router.refresh()
        } catch {
          toast.error('Payment verification failed. Please contact support.')
        }
      },
      modal: {
        ondismiss: () => {
          setPaying(false)
          toast.info('Payment cancelled. You can try again anytime.')
        },
      },
    }
    openCheckout(options)
  }, [pendingInvoice, installationId, userEmail, userName, router, openCheckout, paying])

  const handleSubscribe = useCallback(
    async (
      allocations: { planId: string; quantity: number }[],
      billingPeriod: 'monthly' | 'annual',
    ) => {
      try {
        const order = await createInstallationOrder(installationId, allocations, billingPeriod)
        const key = await getRazorpayKey()

        const totalUsers = allocations.reduce((sum, a) => sum + a.quantity, 0)
        const periodLabel = billingPeriod === 'annual' ? 'Annual' : 'Monthly'
        const options = {
          key,
          order_id: order.razorpayOrderId,
          amount: order.amount,
          currency: order.currency,
          name: 'Kloudlite',
          description: `${totalUsers} ${totalUsers === 1 ? 'user' : 'users'} — ${periodLabel} — Kloudlite Cloud`,
          prefill: { email: userEmail, name: userName },
          theme: { color: '#3B82F6' },
          handler: async (response: Record<string, string>) => {
            try {
              await verifyPaymentAndActivate(
                installationId,
                response.razorpay_order_id,
                response.razorpay_payment_id,
                response.razorpay_signature,
              )
              toast.success('Payment successful!')
              router.refresh()
            } catch {
              toast.error('Payment verification failed. Please contact support.')
            }
          },
          modal: {
            ondismiss: () => {
              toast.info('Payment cancelled. You can try again anytime.')
            },
          },
        }
        openCheckout(options)
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to create order')
      }
    },
    [installationId, userEmail, userName, router, openCheckout],
  )

  return {
    paying,
    editing,
    setEditing,
    handleSubscribe,
    handleModify,
    handlePayNow,
    handleCancelScheduledDowngrade,
  }
}
