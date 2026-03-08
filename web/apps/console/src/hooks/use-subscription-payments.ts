'use client'

import { useCallback, useRef, useState } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { loadStripe } from '@stripe/stripe-js'
import type { Stripe } from '@stripe/stripe-js'
import { createCheckoutSession, createPortalSession } from '@/app/actions/billing/checkout'
import { modifySubscription } from '@/app/actions/billing/subscriptions'

interface UseSubscriptionPaymentsOptions {
  installationId: string
}

export function useSubscriptionPayments({ installationId }: UseSubscriptionPaymentsOptions) {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const stripeRef = useRef<Stripe | null>(null)

  const getStripeClient = useCallback(async () => {
    if (!stripeRef.current) {
      const key = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
      if (!key) throw new Error('Stripe publishable key not configured')
      stripeRef.current = await loadStripe(key)
    }
    return stripeRef.current
  }, [])

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

        if (result.clientSecret) {
          // Payment requires 3D Secure authentication
          const stripe = await getStripeClient()
          if (!stripe) throw new Error('Failed to load Stripe')

          toast.info('Payment requires authentication. Please complete the verification.')

          const { error } = await stripe.confirmPayment({
            clientSecret: result.clientSecret,
            confirmParams: {
              return_url: `${window.location.origin}/installations/${installationId}/billing?modified=success`,
            },
          })

          if (error) {
            // User cancelled or auth failed — the subscription items were already
            // updated but the invoice is unpaid. It will retry automatically.
            toast.error(error.message || 'Payment authentication failed')
          } else {
            // If confirmPayment doesn't redirect (e.g., immediate success), refresh
            toast.success('Subscription updated and payment confirmed.')
            router.refresh()
          }
        } else {
          // No 3DS needed — payment went through directly
          toast.success('Subscription updated. Prorated charges applied.')
          router.refresh()
        }
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to modify subscription')
      } finally {
        setLoading(false)
      }
    },
    [installationId, router, getStripeClient],
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
