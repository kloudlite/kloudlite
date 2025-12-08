'use client'

import { useEffect, useState, useCallback } from 'react'

export type VPNStatus = 'checking' | 'connected' | 'disconnected'

interface UseVPNStatusOptions {
  pollInterval?: number
  enabled?: boolean
}

export function useVPNStatus(options: UseVPNStatusOptions = {}) {
  const { pollInterval = 30000, enabled = true } = options
  const [status, setStatus] = useState<VPNStatus>('checking')
  const [mounted, setMounted] = useState(false)

  const checkVPNStatus = useCallback(async () => {
    try {
      // Extract subdomain from current hostname
      // Expected format: subdomain.khost.dev or *.subdomain.khost.dev
      const hostname = window.location.hostname
      const baseDomain = 'khost.dev' // This should match CLOUDFLARE_DNS_DOMAIN

      // Parse subdomain from hostname
      // Examples:
      // - "test.khost.dev" -> "test"
      // - "console.test.khost.dev" -> "test"
      const hostParts = hostname.split('.')
      const baseParts = baseDomain.split('.')

      let subdomain: string | null = null

      if (hostParts.length > baseParts.length) {
        // Get the part before the base domain
        // For "console.test.khost.dev" with base "khost.dev", we want "test"
        subdomain = hostParts[hostParts.length - baseParts.length - 1]
      }

      if (!subdomain) {
        setStatus('disconnected')
        return
      }

      // Construct VPN check URL and hit it directly from browser
      // The session cookie will be sent automatically with credentials: 'include'
      const vpnCheckUrl = `https://vpn-check.${subdomain}.${baseDomain}`

      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 5000)

      try {
        const response = await fetch(vpnCheckUrl, {
          method: 'GET',
          credentials: 'include', // Send cookies cross-origin
          signal: controller.signal,
          cache: 'no-cache',
        })
        clearTimeout(timeoutId)

        if (response.ok) {
          setStatus('connected')
        } else {
          setStatus('disconnected')
        }
      } catch (_fetchError) {
        clearTimeout(timeoutId)
        setStatus('disconnected')
      }
    } catch {
      setStatus('disconnected')
    }
  }, [])

  useEffect(() => {
    setMounted(true)

    if (!enabled) {
      setStatus('disconnected')
      return
    }

    checkVPNStatus()

    // Poll for status
    const interval = setInterval(() => {
      checkVPNStatus()
    }, pollInterval)

    return () => clearInterval(interval)
  }, [enabled, pollInterval, checkVPNStatus])

  return {
    status,
    isConnected: status === 'connected',
    isChecking: !mounted || status === 'checking',
    checkVPNStatus,
  }
}
