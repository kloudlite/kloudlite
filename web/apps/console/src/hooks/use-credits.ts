'use client'

import { useState, useEffect, useCallback } from 'react'
import { toast } from 'sonner'
import type {
  CreditAccount,
  CreditTransaction,
  UsagePeriod,
  PricingTier,
} from '@/lib/console/storage/credits-types'

interface CreditsData {
  account: CreditAccount | null
  transactions: CreditTransaction[]
  activePeriods: UsagePeriod[]
  pricingTiers: PricingTier[]
}

interface UseCreditsReturn {
  loading: boolean
  data: CreditsData | null
  refresh: () => Promise<void>
  handleTopup: (amount: number, returnUrl?: string) => Promise<void>
  handleManageBilling: () => Promise<void>
  handleUpdateAutoTopup: (
    enabled: boolean,
    threshold?: number,
    amount?: number,
  ) => Promise<void>
}

export function useCredits(orgId: string, options?: { skipInitialFetch?: boolean }): UseCreditsReturn {
  const [loading, setLoading] = useState(true)
  const [data, setData] = useState<CreditsData | null>(null)

  const fetchCredits = useCallback(async () => {
    try {
      setLoading(true)
      const res = await fetch(`/api/orgs/${orgId}/billing/credits`)
      if (!res.ok) throw new Error('Failed to fetch credits')
      const json = await res.json()
      setData(json)
    } catch (err) {
      console.error('Failed to fetch credits:', err)
      toast.error('Failed to load billing information')
    } finally {
      setLoading(false)
    }
  }, [orgId])

  useEffect(() => {
    if (!options?.skipInitialFetch) {
      fetchCredits()
    }
  }, [fetchCredits, options?.skipInitialFetch])

  const handleTopup = useCallback(
    async (amount: number, returnUrl?: string) => {
      try {
        setLoading(true)
        const res = await fetch(`/api/orgs/${orgId}/billing/topup`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ amount, returnUrl }),
        })
        if (!res.ok) {
          const error = await res.json()
          throw new Error(error.error || 'Failed to create top-up')
        }
        const { url } = await res.json()
        if (url) {
          window.location.href = url
        }
      } catch (err) {
        toast.error(
          err instanceof Error ? err.message : 'Failed to process top-up',
        )
        setLoading(false)
      }
    },
    [orgId],
  )

  const handleManageBilling = useCallback(async () => {
    try {
      setLoading(true)
      const res = await fetch(`/api/orgs/${orgId}/billing/portal`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
      })
      if (!res.ok) throw new Error('Failed to open billing portal')
      const { url } = await res.json()
      if (url) {
        window.location.href = url
      }
    } catch (err) {
      toast.error('Failed to open billing portal')
      setLoading(false)
    }
  }, [orgId])

  const handleUpdateAutoTopup = useCallback(
    async (enabled: boolean, threshold?: number, amount?: number) => {
      try {
        const res = await fetch(`/api/orgs/${orgId}/billing/credits`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            autoTopupEnabled: enabled,
            autoTopupThreshold: threshold,
            autoTopupAmount: amount,
          }),
        })
        if (!res.ok)
          throw new Error('Failed to update auto top-up settings')
        toast.success('Auto top-up settings updated')
        await fetchCredits()
      } catch (err) {
        toast.error(
          err instanceof Error
            ? err.message
            : 'Failed to update settings',
        )
      }
    },
    [orgId, fetchCredits],
  )

  return {
    loading,
    data,
    refresh: fetchCredits,
    handleTopup,
    handleManageBilling,
    handleUpdateAutoTopup,
  }
}
