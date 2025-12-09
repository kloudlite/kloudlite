'use client'

import { useState, useEffect } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@kloudlite/ui'
import { Button } from '@kloudlite/ui'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@kloudlite/ui'
import { Copy, Check, Loader2, AlertCircle } from 'lucide-react'
import { toast } from 'sonner'

interface KltunSetupDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

interface TokenResponse {
  temporary_token: string
  expires_at: string
  expires_in: number
  server_url: string
}

type Platform = 'unix' | 'windows'

export function KltunSetupDialog({ open, onOpenChange }: KltunSetupDialogProps) {
  const [tokenData, setTokenData] = useState<TokenResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null)
  const [timeRemaining, setTimeRemaining] = useState<number>(0)
  const [platform, setPlatform] = useState<Platform>('unix')

  // Generate token when dialog opens
  useEffect(() => {
    if (open && !tokenData) {
      generateToken()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  // Countdown timer
  useEffect(() => {
    if (tokenData) {
      const expiryTime = new Date(tokenData.expires_at).getTime()
      const updateTimer = () => {
        const now = Date.now()
        const remaining = Math.max(0, Math.floor((expiryTime - now) / 1000))
        setTimeRemaining(remaining)
      }

      updateTimer()
      const interval = setInterval(updateTimer, 1000)
      return () => clearInterval(interval)
    }
    return undefined
  }, [tokenData])

  const generateToken = async () => {
    setLoading(true)
    setError(null)

    try {
      const response = await fetch('/api/vpn/generate-token', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })

      if (!response.ok) {
        throw new Error('Failed to generate token')
      }

      const data = await response.json()
      setTokenData(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate token')
      toast.error('Failed to generate token. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const copyToClipboard = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedCommand(label)
      toast.success('Copied to clipboard')
      setTimeout(() => setCopiedCommand(null), 2000)
    } catch {
      toast.error('Failed to copy to clipboard')
    }
  }

  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const serverUrl = tokenData?.server_url || ''
  const token = tokenData?.temporary_token || ''

  const getUnixInstallScript = () => {
    const scriptUrl = `${serverUrl}/kltun`
    return `curl -fsSL "${scriptUrl}" | sh -s -- --token "${token}"`
  }

  const getWindowsInstallScript = () => {
    const scriptUrl = `${serverUrl}/kltun.ps1?token=${encodeURIComponent(token)}`
    return `iwr "${scriptUrl}" -UseBasicParsing | iex`
  }

  const CodeBlock = ({ command, label }: { command: string; label: string }) => {
    const isCopied = copiedCommand === label
    return (
      <div className="flex items-start gap-2">
        <div className="flex-1 bg-muted rounded-md p-3">
          <code className="font-mono text-sm break-all">{command}</code>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => copyToClipboard(command, label)}
          disabled={!token}
          className="flex-shrink-0"
        >
          {isCopied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
        </Button>
      </div>
    )
  }

  const InstallInstructions = () => {
    const unixScript = getUnixInstallScript()
    const windowsScript = getWindowsInstallScript()

    return (
      <div className="space-y-4">
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold">Run this command to install and connect kltun:</h3>
            {timeRemaining > 0 && (
              <div className="flex items-center gap-2 text-sm text-orange-600 dark:text-orange-400">
                <AlertCircle className="h-4 w-4" />
                <span>Token expires in {formatTime(timeRemaining)}</span>
              </div>
            )}
          </div>

          <Tabs value={platform} onValueChange={(v) => setPlatform(v as Platform)} className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="unix">Linux / macOS</TabsTrigger>
              <TabsTrigger value="windows">Windows (PowerShell)</TabsTrigger>
            </TabsList>
            <TabsContent value="unix" className="mt-3">
              <CodeBlock command={unixScript} label="install-kltun-unix" />
              <p className="text-sm text-muted-foreground mt-2">
                Run this in Terminal (macOS) or any shell (Linux). Requires curl.
              </p>
            </TabsContent>
            <TabsContent value="windows" className="mt-3">
              <CodeBlock command={windowsScript} label="install-kltun-windows" />
              <p className="text-sm text-muted-foreground mt-2">
                Run this in PowerShell (as Administrator for best results).
              </p>
            </TabsContent>
          </Tabs>

          {timeRemaining === 0 && token && (
            <div className="flex items-center gap-2 text-sm text-red-600 dark:text-red-400">
              <AlertCircle className="h-4 w-4" />
              <span>Token expired. Please close and reopen this dialog to generate a new token.</span>
            </div>
          )}
          <p className="text-sm text-muted-foreground">
            This script will download kltun, install it to your system, and connect to your WorkMachine.
          </p>
        </div>
      </div>
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Setup kltun</DialogTitle>
          <DialogDescription>
            Install kltun and connect to your WorkMachine
          </DialogDescription>
        </DialogHeader>

        {loading && (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            <span className="ml-3 text-sm text-muted-foreground">Generating token...</span>
          </div>
        )}

        {error && !loading && (
          <div className="space-y-4 py-4">
            <div className="flex items-center gap-2 text-sm text-red-600 dark:text-red-400">
              <AlertCircle className="h-4 w-4" />
              <span>{error}</span>
            </div>
            <Button onClick={generateToken} variant="outline">
              Try Again
            </Button>
          </div>
        )}

        {tokenData && !loading && !error && (
          <InstallInstructions />
        )}
      </DialogContent>
    </Dialog>
  )
}
