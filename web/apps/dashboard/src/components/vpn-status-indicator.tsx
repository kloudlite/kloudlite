'use client'

import { useEffect, useState } from 'react'
import { Shield, ShieldOff, Loader2, Copy, Check } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@kloudlite/ui'

type VPNStatus = 'checking' | 'connected' | 'disconnected'

export function VPNStatusIndicator() {
  const [status, setStatus] = useState<VPNStatus>('checking')
  const [token, setToken] = useState<string>('')
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
    checkVPNStatus()

    // Poll for status every 30 seconds
    const interval = setInterval(() => {
      checkVPNStatus()
    }, 30000)

    return () => clearInterval(interval)
  }, [])

  const checkVPNStatus = async () => {
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
  }

  const generateToken = async () => {
    try {
      const response = await fetch('/api/vpn/generate-token', { method: 'POST' })
      const data = await response.json()

      if (data.temporary_token) {
        setToken(data.temporary_token)
      }
    } catch (_error) {
      console.error('Failed to generate VPN token:', _error)
    }
  }

  const handleDropdownOpen = (open: boolean) => {
    if (open && !token) {
      generateToken()
    }
  }

  const copyToClipboard = async (text: string, commandType: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedCommand(commandType)
      setTimeout(() => setCopiedCommand(null), 2000)
    } catch (_error) {
      console.error('Failed to copy:', _error)
    }
  }

  const getIcon = () => {
    if (!mounted || status === 'checking') {
      return <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
    }
    if (status === 'connected') {
      return <Shield className="h-4 w-4 text-green-500" />
    }
    return <ShieldOff className="h-4 w-4 text-red-500" />
  }

  const getStatusMessage = () => {
    if (status === 'checking') return 'Checking VPN status...'
    if (status === 'connected') return 'VPN Connected'
    return 'VPN Not Connected'
  }

  const serverUrl = typeof window !== 'undefined' ? window.location.origin : ''

  const curlCommand = `curl -fsSL "${serverUrl}/kltun" | sh -s -- --token ${token || 'GENERATING...'}`
  const wgetCommand = `wget -qO- "${serverUrl}/kltun" | sh -s -- --token ${token || 'GENERATING...'}`
  const powershellCommand = `Invoke-WebRequest -Uri "${serverUrl}/kltun" -UseBasicParsing | Invoke-Expression; kltun connect --token ${token || 'GENERATING...'} --server ${serverUrl}`

  const CopyButton = ({ text, commandType }: { text: string; commandType: string }) => (
    <button
      onClick={() => copyToClipboard(text, commandType)}
      className="ml-2 p-1 hover:bg-accent rounded"
      disabled={!token}
    >
      {copiedCommand === commandType ? (
        <Check className="h-4 w-4 text-green-500" />
      ) : (
        <Copy className="h-4 w-4" />
      )}
    </button>
  )

  return (
    <DropdownMenu onOpenChange={handleDropdownOpen}>
      <DropdownMenuTrigger asChild>
        <button
          className="text-muted-foreground hover:text-foreground transition-colors"
          title={getStatusMessage()}
        >
          {getIcon()}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[500px]">
        <DropdownMenuLabel className="flex items-center gap-2">
          {getIcon()}
          {getStatusMessage()}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />

        {status === 'disconnected' && (
          <>
            <div className="px-2 py-2">
              <p className="text-sm text-muted-foreground mb-3">
                Connect to Kloudlite VPN to access your workspaces and services.
              </p>

              <Tabs defaultValue="curl" className="w-full">
                <TabsList className="grid w-full grid-cols-3">
                  <TabsTrigger value="curl">curl</TabsTrigger>
                  <TabsTrigger value="wget">wget</TabsTrigger>
                  <TabsTrigger value="powershell">PowerShell</TabsTrigger>
                </TabsList>

                <TabsContent value="curl" className="mt-3">
                  <div className="space-y-2">
                    <label className="text-xs font-medium">Unix/Linux (curl)</label>
                    <div className="flex items-start gap-2">
                      <code className="flex-1 block p-2 bg-muted rounded text-xs break-all">
                        {curlCommand}
                      </code>
                      <CopyButton text={curlCommand} commandType="curl" />
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="wget" className="mt-3">
                  <div className="space-y-2">
                    <label className="text-xs font-medium">Unix/Linux (wget)</label>
                    <div className="flex items-start gap-2">
                      <code className="flex-1 block p-2 bg-muted rounded text-xs break-all">
                        {wgetCommand}
                      </code>
                      <CopyButton text={wgetCommand} commandType="wget" />
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="powershell" className="mt-3">
                  <div className="space-y-2">
                    <label className="text-xs font-medium">Windows (PowerShell)</label>
                    <div className="flex items-start gap-2">
                      <code className="flex-1 block p-2 bg-muted rounded text-xs break-all">
                        {powershellCommand}
                      </code>
                      <CopyButton text={powershellCommand} commandType="powershell" />
                    </div>
                  </div>
                </TabsContent>
              </Tabs>

              {!token && (
                <p className="text-xs text-muted-foreground mt-2">
                  Generating token...
                </p>
              )}
            </div>
          </>
        )}

        {status === 'connected' && (
          <div className="px-2 py-2">
            <p className="text-sm text-green-600 dark:text-green-400">
              Your VPN connection is active. You can access your workspaces and cluster services.
            </p>
          </div>
        )}

        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={checkVPNStatus}>
          <Loader2 className="mr-2 h-4 w-4" />
          <span>Refresh Status</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
