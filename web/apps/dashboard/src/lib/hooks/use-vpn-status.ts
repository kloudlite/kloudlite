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
      const vpnCheckUrl = `https://vpn-check.${subdomain}.${baseDomain}`

      // Use fetch with mode: 'no-cors' to test connectivity without console errors
      // no-cors mode won't let us check response status, but we can detect if the request completes
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 3000)

      try {
        await fetch(vpnCheckUrl, {
          method: 'HEAD',
          mode: 'no-cors',
          signal: controller.signal,
          cache: 'no-cache',
        })
        clearTimeout(timeoutId)
        // If fetch completes without error, VPN is connected
        setStatus('connected')
      } catch (fetchError) {
        clearTimeout(timeoutId)
        // Distinguish between timeout/abort and actual network errors
        if (fetchError instanceof Error && fetchError.name === 'AbortError') {
          // Timeout - VPN might be slow or disconnected
          setStatus('disconnected')
        } else {
          // Network error - VPN is disconnected
          setStatus('disconnected')
        }
      }
    } catch (_error) {
      // Network error likely means VPN is not connected
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
