'use client'

import { useCallback, useState } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'

interface UseSubscriptionPaymentsOptions {
  orgId: string
}

export function useSubscriptionPayments({ orgId }: UseSubscriptionPaymentsOptions) {
  const router = useRouter()
  const [loading, setLoading] = useState(false)

  const handleSubscribe = useCallback(
    async (allocations: { priceId: string; quantity: number }[]) => {
      setLoading(true)
      try {
        const res = await fetch(`/api/orgs/${orgId}/billing/checkout`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ allocations }),
        })

        if (!res.ok) {
          const data = await res.json().catch(() => ({}))
          throw new Error(data.error || `Failed to start checkout (${res.status})`)
        }

        const { url } = await res.json()
        if (url) window.location.href = url
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to start checkout')
        setLoading(false)
      }
    },
    [orgId],
  )

  const handleModify = useCallback(
    async (modifications: { priceId: string; quantity: number }[]) => {
      setLoading(true)
      try {
        const res = await fetch(`/api/orgs/${orgId}/billing/subscription`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ modifications }),
        })

        if (!res.ok) {
          const data = await res.json().catch(() => ({}))
          throw new Error(data.error || `Failed to modify subscription (${res.status})`)
        }

        toast.success('Subscription updated successfully. Proration applied to your next invoice.')
        router.refresh()
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to modify subscription')
      } finally {
        setLoading(false)
      }
    },
    [orgId, router],
  )

  const handleCancel = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetch(`/api/orgs/${orgId}/billing/subscription`, {
        method: 'DELETE',
      })

      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.error || `Failed to cancel subscription (${res.status})`)
      }

      toast.success('Subscription will be cancelled at the end of the billing period.')
      router.refresh()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to cancel subscription')
    } finally {
      setLoading(false)
    }
  }, [orgId, router])

  const handleManageBilling = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetch(`/api/orgs/${orgId}/billing/portal`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
      })

      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.error || `Failed to open billing portal (${res.status})`)
      }

      const { url } = await res.json()
      if (url) window.location.href = url
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to open billing portal')
      setLoading(false)
    }
  }, [orgId])

  return {
    loading,
    handleSubscribe,
    handleModify,
    handleCancel,
    handleManageBilling,
  }
}
