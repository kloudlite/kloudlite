'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { ExternalLink, Loader2, Copy, CheckCircle2, Clock, AlertCircle } from 'lucide-react'
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
      <div className="flex items-center justify-center py-32">
        <Loader2 className="text-primary size-8 animate-spin" />
      </div>
    )
  }

  if (!installationData || !installationData.subdomain) {
    return null
  }

  const isActive = activeStatus === 'active'

  return (
    <div className="lg:flex lg:gap-12">
      {/* Left Column - Information */}
      <div className="hidden lg:block lg:w-[400px] lg:flex-shrink-0">
        <div className="sticky top-6 space-y-6">
          {/* What's Next Card */}
          <div className="border border-foreground/10 rounded-lg p-6 bg-muted/20">
            <h3 className="text-sm font-semibold text-foreground mb-3">Installation Progress</h3>
            <div className="space-y-3">
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary/10 text-primary flex items-center justify-center text-xs font-semibold">
                  <CheckCircle2 className="w-3 h-3" />
                </div>
                <div>
                  <p className="text-sm font-medium text-foreground">Create installation</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Set up your installation details</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary/10 text-primary flex items-center justify-center text-xs font-semibold">
                  <CheckCircle2 className="w-3 h-3" />
                </div>
                <div>
                  <p className="text-sm font-medium text-foreground">Deploy to cloud</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Install Kloudlite in your infrastructure</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className={`flex-shrink-0 w-5 h-5 rounded-full flex items-center justify-center text-xs font-semibold ${isActive ? 'bg-primary/10 text-primary' : 'bg-primary text-primary-foreground'}`}>
                  {isActive ? <CheckCircle2 className="w-3 h-3" /> : '3'}
                </div>
                <div>
                  <p className="text-sm font-medium text-foreground">Complete</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Your installation is ready</p>
                </div>
              </div>
            </div>
          </div>

          {/* Help Card */}
          <div className="border border-foreground/10 rounded-lg p-6 bg-background">
            <h3 className="text-sm font-semibold text-foreground mb-3">Need Help?</h3>
            <div className="space-y-2">
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
            </div>
          </div>
        </div>
      </div>

      {/* Right Column - Main Content */}
      <div className="space-y-6 lg:flex-1 lg:min-w-0">
        {/* Header */}
        <div>
          <h1 className="text-foreground text-2xl font-semibold tracking-tight">
            {isActive ? 'Installation Complete!' : 'Setting Up Your Installation'}
          </h1>
          <p className="text-muted-foreground mt-1 text-sm">
            {isActive
              ? 'Your Kloudlite installation is ready to use'
              : 'Please wait while your installation becomes active'}
          </p>
        </div>

        {/* Main Content Card */}
        <div className="border border-foreground/10 rounded-lg bg-background">
          <div className="p-8">
            {isActive ? (
              <>
                <div className="mb-6">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="flex size-10 items-center justify-center bg-green-100 dark:bg-green-900/30 rounded-full">
                      <CheckCircle2 className="size-5 text-green-600 dark:text-green-400" />
                    </div>
                    <h2 className="text-xl font-semibold text-foreground">Your Installation is Ready</h2>
                  </div>
                  <p className="text-muted-foreground text-sm">
                    Access your Kloudlite installation dashboard at the URL below
                  </p>
                </div>

                <div className="space-y-4">
                  <div className="bg-foreground/[0.02] p-4 rounded-sm border border-border/60">
                    <p className="mb-2 text-sm font-medium text-foreground">Installation Dashboard URL:</p>
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

                  <div className="flex gap-3">
                    <Button
                      className="flex-1"
                      size="lg"
                      onClick={() => router.push(`/installations/${installationData.installationId}`)}
                    >
                      Open Installation Settings
                    </Button>
                    <Button
                      variant="outline"
                      size="lg"
                      className="flex-1"
                      onClick={() => router.push('/installations')}
                    >
                      View All Installations
                    </Button>
                  </div>

                  <div className="border-t border-border/60 pt-4">
                    <p className="text-muted-foreground text-sm">
                      <strong>What&apos;s next?</strong> Go to Installation Settings to manage your team by adding admins and members.
                      To administrate your installation&apos;s dashboard, generate a Super Admin Login from the settings page.
                    </p>
                  </div>
                </div>
              </>
            ) : (
              <>
                <div className="mb-6">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="flex size-10 items-center justify-center bg-blue-100 dark:bg-blue-900/30 rounded-full">
                      <Clock className="size-5 text-blue-600 dark:text-blue-400 animate-pulse" />
                    </div>
                    <h2 className="text-xl font-semibold text-foreground">Waiting for Installation to Become Active</h2>
                  </div>
                  <p className="text-muted-foreground text-sm">
                    Your installation is being set up. This usually takes 1-3 minutes.
                  </p>
                </div>

                <div className="space-y-4">
                  <div className="bg-foreground/[0.02] p-4 rounded-sm border border-border/60">
                    <p className="mb-2 text-sm font-medium text-foreground">Installation Dashboard URL:</p>
                    <div className="flex items-center justify-between gap-3">
                      <span className="text-muted-foreground font-mono text-lg">
                        {installationData.subdomain}.
                        {process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                      </span>
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

                  <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-900 p-4 rounded-sm">
                    <div className="flex items-start gap-3">
                      <AlertCircle className="size-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
                      <div>
                        <p className="text-sm font-medium text-blue-900 dark:text-blue-200 mb-1">
                          Installation in Progress
                        </p>
                        <p className="text-xs text-blue-800 dark:text-blue-300">
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
                </div>
              </>
            )}
          </div>
        </div>

        {/* Status indicator */}
        <div className="flex items-center justify-center gap-3 text-base">
          {!isActive && (
            <>
              <Loader2 className="size-4 animate-spin text-blue-600" />
              <span className="text-muted-foreground">Waiting for installation to become active...</span>
            </>
          )}
          {isActive && (
            <>
              <CheckCircle2 className="size-4 text-green-600" />
              <span className="text-green-600 font-medium">Installation is active and ready!</span>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
