'use client'

import { useEffect, useState, useCallback, useRef } from 'react'

export type VPNStatus = 'checking' | 'connected' | 'disconnected' | 'reconnecting' | 'idle'

// New response format from kltun daemon
interface DaemonStatusResponse {
  running: boolean
  started_at: string
  uptime_seconds: number
}

interface VPNStatusResponse {
  status: string
  status_message: string
  vpn_ip?: string
  session_id?: string
  connection_uptime_seconds: number
  tunnel_endpoint?: string
  dashboard_server?: string
}

interface StatusResponse {
  daemon: DaemonStatusResponse
  vpn: VPNStatusResponse
  timestamp: string
}

interface VPNStatusInfo {
  status: VPNStatus
  statusMessage?: string
  vpnIP?: string
  sessionID?: string
  connectionUptimeSeconds?: number
  tunnelEndpoint?: string
  dashboardServer?: string
  daemonRunning?: boolean
  daemonUptimeSeconds?: number
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
          mode: 'cors',
          credentials: 'omit',
          signal: controller.signal,
          cache: 'no-cache',
        })
        clearTimeout(timeoutId)

        if (response.ok) {
          const data: StatusResponse = await response.json()

          // Map VPN status from daemon response
          let mappedStatus: VPNStatus = 'disconnected'
          switch (data.vpn.status) {
            case 'connected':
              mappedStatus = 'connected'
              break
            case 'reconnecting':
              mappedStatus = 'reconnecting'
              break
            case 'idle':
              mappedStatus = 'idle'
              break
            default:
              mappedStatus = 'disconnected'
          }

          lastStatusRef.current = mappedStatus
          setStatusInfo({
            status: mappedStatus,
            statusMessage: data.vpn.status_message,
            vpnIP: data.vpn.vpn_ip,
            sessionID: data.vpn.session_id,
            connectionUptimeSeconds: data.vpn.connection_uptime_seconds,
            tunnelEndpoint: data.vpn.tunnel_endpoint,
            dashboardServer: data.vpn.dashboard_server,
            daemonRunning: data.daemon.running,
            daemonUptimeSeconds: data.daemon.uptime_seconds,
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
    statusMessage: statusInfo.statusMessage,
    isConnected: statusInfo.status === 'connected',
    isReconnecting: statusInfo.status === 'reconnecting',
    isIdle: statusInfo.status === 'idle',
    isChecking: !mounted || statusInfo.status === 'checking',
    vpnIP: statusInfo.vpnIP,
    sessionID: statusInfo.sessionID,
    connectionUptimeSeconds: statusInfo.connectionUptimeSeconds,
    tunnelEndpoint: statusInfo.tunnelEndpoint,
    dashboardServer: statusInfo.dashboardServer,
    daemonRunning: statusInfo.daemonRunning,
    daemonUptimeSeconds: statusInfo.daemonUptimeSeconds,
    lastChecked: statusInfo.lastChecked,
    checkVPNStatus,
  }
}
