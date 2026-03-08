'use client'

import { useCallback, useState } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { createCheckoutSession, createPortalSession } from '@/app/actions/billing/checkout'
import { modifySubscription } from '@/app/actions/billing/subscriptions'

interface UseSubscriptionPaymentsOptions {
  installationId: string
}

export function useSubscriptionPayments({ installationId }: UseSubscriptionPaymentsOptions) {
  const router = useRouter()
  const [loading, setLoading] = useState(false)

  const handleSubscribe = useCallback(
    async (allocations: { priceId: string; quantity: number }[]) => {
      setLoading(true)
      try {
        const { url } = await createCheckoutSession(installationId, allocations)
        if (url) window.location.href = url
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to start checkout')
        setLoading(false)
      }
    },
    [installationId],
  )

  const handleModify = useCallback(
    async (modifications: { priceId: string; quantity: number }[]) => {
      setLoading(true)
      try {
        const result = await modifySubscription(installationId, modifications)

        if (result.url) {
          // Redirect to Stripe hosted invoice page for payment
          window.location.href = result.url
        } else {
          // No payment needed (e.g., downgrade with credit)
          toast.success('Subscription updated successfully.')
          router.refresh()
          setLoading(false)
        }
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to modify subscription')
        setLoading(false)
      }
    },
    [installationId, router],
  )

  const handleManageBilling = useCallback(async () => {
    setLoading(true)
    try {
      const { url } = await createPortalSession(installationId)
      if (url) window.location.href = url
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to open billing portal')
      setLoading(false)
    }
  }, [installationId])

  return {
    loading,
    handleSubscribe,
    handleModify,
    handleManageBilling,
  }
}
