'use client'

import { useEffect, useState, useCallback, useRef } from 'react'

export type VPNStatus = 'checking' | 'connected' | 'disconnected'

interface VPNStatusResponse {
  status: string
  vpn_ip: string
  session_id: string
  uptime_seconds: number
  tunnel_endpoint: string
  timestamp: string
}

interface VPNStatusInfo {
  status: VPNStatus
  vpnIP?: string
  sessionID?: string
  uptimeSeconds?: number
  tunnelEndpoint?: string
  lastChecked?: Date
}

interface UseVPNStatusOptions {
  pollInterval?: number
  enabled?: boolean
}

export function useVPNStatus(options: UseVPNStatusOptions = {}) {
  const { pollInterval = 30000, enabled = true } = options
  const [statusInfo, setStatusInfo] = useState<VPNStatusInfo>({
    status: 'checking',
  })
  const [mounted, setMounted] = useState(false)
  const lastStatusRef = useRef<VPNStatus>('checking')

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
        if (lastStatusRef.current !== 'disconnected') {
          lastStatusRef.current = 'disconnected'
          setStatusInfo({ status: 'disconnected' })
        }
        return
      }

      // Construct VPN check URL and hit the /status endpoint
      const vpnCheckUrl = `https://vpn-check.${subdomain}.${baseDomain}/status`

      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 5000)

      try {
        const response = await fetch(vpnCheckUrl, {
          method: 'GET',
          signal: controller.signal,
          cache: 'no-cache',
        })
        clearTimeout(timeoutId)

        if (response.ok) {
          const data: VPNStatusResponse = await response.json()
          lastStatusRef.current = 'connected'
          setStatusInfo({
            status: 'connected',
            vpnIP: data.vpn_ip,
            sessionID: data.session_id,
            uptimeSeconds: data.uptime_seconds,
            tunnelEndpoint: data.tunnel_endpoint,
            lastChecked: new Date(),
          })
        } else {
          if (lastStatusRef.current !== 'disconnected') {
            lastStatusRef.current = 'disconnected'
            setStatusInfo({ status: 'disconnected', lastChecked: new Date() })
          }
        }
      } catch (_fetchError) {
        clearTimeout(timeoutId)
        if (lastStatusRef.current !== 'disconnected') {
          lastStatusRef.current = 'disconnected'
          setStatusInfo({ status: 'disconnected', lastChecked: new Date() })
        }
      }
    } catch {
      if (lastStatusRef.current !== 'disconnected') {
        lastStatusRef.current = 'disconnected'
        setStatusInfo({ status: 'disconnected' })
      }
    }
  }, [])

  useEffect(() => {
    setMounted(true)

    if (!enabled) {
      setStatusInfo({ status: 'disconnected' })
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
    status: statusInfo.status,
    isConnected: statusInfo.status === 'connected',
    isChecking: !mounted || statusInfo.status === 'checking',
    vpnIP: statusInfo.vpnIP,
    sessionID: statusInfo.sessionID,
    uptimeSeconds: statusInfo.uptimeSeconds,
    tunnelEndpoint: statusInfo.tunnelEndpoint,
    lastChecked: statusInfo.lastChecked,
    checkVPNStatus,
  }
}
