'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle, Button } from '@kloudlite/ui'
import { ExternalLink, Loader2, Copy, PartyPopper, Clock, AlertCircle } from 'lucide-react'
import { InstallationProgress } from '@/components/installation-progress'
import { toast } from 'sonner'

interface InstallationData {
  subdomain: string
  url: string
  installationId: string
}

type ActiveStatus = 'checking' | 'active' | 'waiting' | 'error'

export default function CompletePage() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [installationData, setInstallationData] = useState<InstallationData | null>(null)
  const [activeStatus, setActiveStatus] = useState<ActiveStatus>('checking')
  const [checkCount, setCheckCount] = useState(0)

  // Function to check if installation is active
  const checkActiveStatus = useCallback(async (installationId: string) => {
    try {
      const response = await fetch(`/api/installations/${installationId}/ping`)
      if (response.ok) {
        const data = await response.json()
        if (data.active) {
          setActiveStatus('active')
          return true
        } else {
          setActiveStatus('waiting')
          return false
        }
      }
      setActiveStatus('waiting')
      return false
    } catch {
      setActiveStatus('waiting')
      return false
    }
  }, [])

  useEffect(() => {
    const fetchInstallationData = async () => {
      try {
        const response = await fetch('/api/installations/session')
        if (response.ok) {
          const sessionData = await response.json()

          // Fetch installation details using verify-key
          if (sessionData.installationKey) {
            const verifyResponse = await fetch('/api/installations/verify-key', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ installationKey: sessionData.installationKey }),
            })
            if (verifyResponse.ok) {
              const verifyData = await verifyResponse.json()
              // Validate subdomain is not a placeholder/invalid value
              const isValidSubdomain = verifyData.subdomain &&
                verifyData.subdomain !== '0.0.0.0' &&
                !verifyData.subdomain.includes('0.0.0.0')

              if (isValidSubdomain) {
                const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'
                setInstallationData({
                  subdomain: verifyData.subdomain,
                  url: `https://${verifyData.subdomain}.${domain}`,
                  installationId: verifyData.installationId,
                })

                // Start checking active status
                if (verifyData.installationId) {
                  checkActiveStatus(verifyData.installationId)
                }
              }
            }
          }
        } else {
          router.push('/login')
        }
      } catch (error) {
        console.error('Error fetching installation data:', error)
      } finally {
        setLoading(false)
      }
    }
    fetchInstallationData()
  }, [router, checkActiveStatus])

  // Poll for active status every 5 seconds until active
  useEffect(() => {
    if (!installationData?.installationId || activeStatus === 'active') {
      return
    }

    const intervalId = setInterval(async () => {
      setCheckCount(c => c + 1)
      const isActive = await checkActiveStatus(installationData.installationId)
      if (isActive) {
        clearInterval(intervalId)
      }
    }, 5000)

    return () => clearInterval(intervalId)
  }, [installationData?.installationId, activeStatus, checkActiveStatus])

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} copied to clipboard`)
  }

  if (loading) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center">
        <Loader2 className="text-primary size-8 animate-spin" />
      </div>
    )
  }

  if (!installationData || !installationData.subdomain) {
    return null
  }

  // Render waiting state while installation is not yet active
  const renderWaitingState = () => (
    <div className="w-full">
      <div className="mb-8 text-center">
        <div className="mb-4 flex justify-center">
          <div className="flex size-16 items-center justify-center  bg-blue-100">
            <Clock className="size-8 text-blue-600 animate-pulse" />
          </div>
        </div>
        <h1 className="text-foreground mb-2 text-3xl font-semibold">Setting Up Your Installation</h1>
        <p className="text-muted-foreground">Please wait while your installation becomes active</p>
      </div>

      <InstallationProgress currentStep={3} />

      <div className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-xl flex items-center gap-2">
              <Loader2 className="size-5 animate-spin" />
              Waiting for Installation to Become Active
            </CardTitle>
            <CardDescription>
              Your installation is being set up. This usually takes 1-3 minutes.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="bg-muted  p-4">
              <p className="mb-2 text-sm font-medium">Installation Dashboard URL:</p>
              <div className="flex items-center justify-between gap-3">
                <span className="text-muted-foreground font-mono text-lg">
                  {installationData?.subdomain}.
                  {process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => installationData && copyToClipboard(installationData.url, 'URL')}
                >
                  <Copy className="mr-2 size-3" />
                  Copy
                </Button>
              </div>
            </div>

            <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-900  p-4">
              <div className="flex items-start gap-3">
                <AlertCircle className="size-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
                <div>
                  <p className="text-sm font-medium text-blue-900 dark:text-blue-200 mb-1">
                    Installation in Progress
                  </p>
                  <p className="text-sm text-blue-800 dark:text-blue-300">
                    We&apos;re checking if your installation is ready. Checked {checkCount} time{checkCount !== 1 ? 's' : ''}.
                    The page will automatically update when your installation is active.
                  </p>
                </div>
              </div>
            </div>

            <div className="flex gap-3">
              <Button
                variant="outline"
                size="lg"
                className="flex-1"
                onClick={() => router.push('/installations')}
              >
                View All Installations
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )

  // Render active/ready state
  const renderActiveState = () => (
    <div className="w-full">
      <div className="mb-8 text-center">
        <div className="mb-4 flex justify-center">
          <div className="flex size-16 items-center justify-center  bg-green-100">
            <PartyPopper className="size-8 text-green-600" />
          </div>
        </div>
        <h1 className="text-foreground mb-2 text-3xl font-semibold">Installation Complete!</h1>
        <p className="text-muted-foreground">Your Kloudlite installation is ready to use</p>
      </div>

      <InstallationProgress currentStep={3} />

      <div className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-xl">Your Installation is Ready</CardTitle>
            <CardDescription>
              Access your Kloudlite installation dashboard at the URL below
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {installationData?.subdomain ? (
              <div className="bg-muted  p-4">
                <p className="mb-2 text-sm font-medium">Installation Dashboard URL:</p>
                <div className="flex items-center justify-between gap-3">
                  <a
                    href={installationData.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary flex items-center gap-2 font-mono text-lg hover:underline"
                  >
                    {installationData.subdomain}.
                    {process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                    <ExternalLink className="size-4" />
                  </a>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => copyToClipboard(installationData.url, 'URL')}
                  >
                    <Copy className="mr-2 size-3" />
                    Copy
                  </Button>
                </div>
              </div>
            ) : (
              <div className="bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-900  p-4">
                <p className="text-sm font-medium text-amber-900 dark:text-amber-200 mb-1">
                  Domain Not Configured
                </p>
                <p className="text-sm text-amber-900 dark:text-amber-200">
                  Your installation key was generated, but no subdomain was configured. Please configure a domain for your installation from the installations list.
                </p>
              </div>
            )}

            <div className="flex gap-3">
              {installationData?.subdomain ? (
                <Button
                  className="flex-1"
                  size="lg"
                  onClick={() => window.open(installationData.url, '_blank')}
                >
                  <ExternalLink className="mr-2 size-4" />
                  Open Installation Dashboard
                </Button>
              ) : null}
              <Button
                variant={installationData?.subdomain ? 'outline' : 'default'}
                size="lg"
                className="flex-1"
                onClick={() => router.push('/installations')}
              >
                View All Installations
              </Button>
            </div>

            <div className="border-t pt-4">
              <p className="text-muted-foreground text-sm">
                <strong>What&apos;s next?</strong> You can now access your Kloudlite installation
                dashboard to create and manage workspaces, environments, and work machines. Your
                team members can log in using their own credentials at your installation URL.
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Need Help?</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <a
              href="https://docs.kloudlite.io"
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary flex items-center gap-2 text-sm hover:underline"
            >
              <ExternalLink className="size-4" />
              Read the Documentation
            </a>
            <a
              href="https://discord.gg/kloudlite"
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary flex items-center gap-2 text-sm hover:underline"
            >
              <ExternalLink className="size-4" />
              Join our Discord Community
            </a>
          </CardContent>
        </Card>
      </div>
    </div>
  )

  // Conditionally render based on active status
  if (activeStatus === 'active') {
    return renderActiveState()
  }

  return renderWaitingState()
}
