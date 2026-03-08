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
          // Payment needed — use stripe.js to confirm and redirect
          const stripe = await getStripeClient()
          if (!stripe) throw new Error('Failed to load Stripe')

          const returnUrl = `${window.location.origin}/installations/${installationId}/billing?modified=success`

          const { error } = await stripe.confirmPayment({
            clientSecret: result.clientSecret,
            confirmParams: {
              return_url: returnUrl,
            },
            redirect: 'always',
          })

          // confirmPayment only returns here if there's an error
          // (on success it redirects to return_url)
          if (error) {
            toast.error(error.message || 'Payment failed. Please try again.')
            setLoading(false)
          }
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
