'use client'

import { useState, useCallback } from 'react'

const SUBDOMAIN_REGEX = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/

interface UseSubdomainCheckOptions {
  endpoint?: string
}

export function useSubdomainCheck({ endpoint = '/api/installations/check-subdomain' }: UseSubdomainCheckOptions = {}) {
  const [checking, setChecking] = useState(false)
  const [available, setAvailable] = useState<boolean | null>(null)

  const check = useCallback(async (subdomain: string) => {
    if (!subdomain || subdomain.length < 3) {
      setAvailable(null)
      return
    }

    if (!SUBDOMAIN_REGEX.test(subdomain)) {
      setAvailable(null)
      return
    }

    setChecking(true)
    try {
      const response = await fetch(`${endpoint}?subdomain=${subdomain}`)
      const data = await response.json()
      setAvailable(data.available)
    } catch (err) {
      console.error('Error checking subdomain:', err)
      setAvailable(false)
    } finally {
      setChecking(false)
    }
  }, [endpoint])

  return { checking, available, check }
}
