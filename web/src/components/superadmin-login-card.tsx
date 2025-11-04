'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Shield, Copy, Check, Loader2, Clock, ExternalLink } from 'lucide-react'

interface SuperAdminLoginCardProps {
  installationId: string
  isActive: boolean
}

export function SuperAdminLoginCard({ installationId, isActive }: SuperAdminLoginCardProps) {
  const [loading, setLoading] = useState(false)
  const [loginUrl, setLoginUrl] = useState<string | null>(null)
  const [_expiresAt, setExpiresAt] = useState<string | null>(null)
  const [_validForSeconds, setValidForSeconds] = useState<number | null>(null)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [timeRemaining, setTimeRemaining] = useState<number | null>(null)

  const generateLoginUrl = async () => {
    setLoading(true)
    setError(null)
    setLoginUrl(null)

    try {
      const response = await fetch(`/api/installations/${installationId}/superadmin-login-token`, {
        method: 'POST',
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'Failed to generate login URL')
      }

      const data = await response.json()
      setLoginUrl(data.loginUrl)
      setExpiresAt(data.expiresAt)
      setValidForSeconds(data.validForSeconds)
      setTimeRemaining(data.validForSeconds)

      // Start countdown timer
      const interval = setInterval(() => {
        setTimeRemaining((prev) => {
          if (prev === null || prev <= 1) {
            clearInterval(interval)
            return 0
          }
          return prev - 1
        })
      }, 1000)

      // Clear URL after expiry
      setTimeout(() => {
        setLoginUrl(null)
        setExpiresAt(null)
        setValidForSeconds(null)
        setTimeRemaining(null)
      }, data.validForSeconds * 1000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate login URL')
    } finally {
      setLoading(false)
    }
  }

  const copyToClipboard = async () => {
    if (!loginUrl) return

    try {
      await navigator.clipboard.writeText(loginUrl)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  if (!isActive) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Shield className="text-muted-foreground h-5 w-5" />
            <CardTitle>Super Admin Access</CardTitle>
          </div>
          <CardDescription>Quick login URL for super admin access</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert>
            <AlertDescription>
              Super admin login is only available for active installations. Please ensure your
              installation is fully configured and active.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Shield className="text-primary h-5 w-5" />
          <CardTitle>Super Admin Access</CardTitle>
        </div>
        <CardDescription>Generate a one-time login URL for super admin access</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <p className="text-muted-foreground text-sm">
            Generate a temporary login URL that provides instant super admin access to your
            installation dashboard. The URL is valid for 5 minutes.
          </p>
        </div>

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {!loginUrl && (
          <Button onClick={generateLoginUrl} disabled={loading} className="w-full sm:w-auto">
            {loading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Generating...
              </>
            ) : (
              <>
                <Shield className="mr-2 h-4 w-4" />
                Generate Login URL
              </>
            )}
          </Button>
        )}

        {loginUrl && (
          <div className="space-y-3">
            <div className="bg-muted rounded-lg border p-4">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-foreground text-sm font-medium">Login URL</span>
                {timeRemaining !== null && (
                  <div className="text-muted-foreground flex items-center gap-1 text-xs">
                    <Clock className="h-3 w-3" />
                    <span>Expires in {formatTime(timeRemaining)}</span>
                  </div>
                )}
              </div>
              <div className="bg-background flex items-center gap-2 rounded border p-2">
                <code className="text-foreground flex-1 overflow-x-auto text-xs">{loginUrl}</code>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={copyToClipboard}
                  className="shrink-0"
                >
                  {copied ? (
                    <Check className="h-4 w-4 text-green-600" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>

            <div className="flex flex-col gap-2 sm:flex-row">
              <Button asChild variant="default" className="flex-1">
                <a href={loginUrl} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="mr-2 h-4 w-4" />
                  Open in New Tab
                </a>
              </Button>
              <Button variant="outline" onClick={generateLoginUrl} disabled={loading}>
                Generate New URL
              </Button>
            </div>

            <Alert>
              <AlertDescription className="text-xs">
                This URL provides full super admin access. Keep it secure and do not share it with
                unauthorized users.
              </AlertDescription>
            </Alert>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
