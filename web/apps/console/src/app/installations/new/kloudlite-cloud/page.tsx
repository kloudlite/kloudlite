'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { Loader2, CheckCircle2, XCircle, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'

type DeployStatus = 'loading' | 'triggering' | 'pending' | 'running' | 'succeeded' | 'failed' | 'error'

export default function KloudliteCloudPage() {
  const router = useRouter()
  const [status, setStatus] = useState<DeployStatus>('loading')
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [installationId, setInstallationId] = useState<string | null>(null)
  const [currentStep, setCurrentStep] = useState(0)
  const [totalSteps, setTotalSteps] = useState(9)
  const [stepDescription, setStepDescription] = useState('')
  const initRef = useRef(false)

  const triggerDeploy = useCallback(async (instId: string) => {
    setStatus('triggering')
    setErrorMessage(null)

    try {
      const response = await fetch(`/api/installations/${instId}/trigger-managed-install`, {
        method: 'POST',
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Failed to trigger deployment')
      }

      setStatus('running')
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to trigger deployment'
      setErrorMessage(message)
      setStatus('failed')
    }
  }, [])

  // On mount: get session, verify key, trigger deploy
  useEffect(() => {
    if (initRef.current) return
    initRef.current = true

    const init = async () => {
      try {
        const sessionRes = await fetch('/api/installations/session')
        if (!sessionRes.ok) {
          router.push('/login')
          return
        }
        const sessionData = await sessionRes.json()

        if (!sessionData.installationKey) {
          router.push('/installations/new-kl-cloud')
          return
        }

        const verifyRes = await fetch('/api/installations/verify-key', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ installationKey: sessionData.installationKey }),
        })

        if (!verifyRes.ok) {
          toast.error('Failed to verify installation')
          router.push('/installations/new-kl-cloud')
          return
        }

        const verifyData = await verifyRes.json()
        const instId = verifyData.installationId
        setInstallationId(instId)

        // Verify the installation has an active subscription before deploying
        try {
          const subRes = await fetch(`/api/installations/${instId}/subscription`)
          if (!subRes.ok) {
            toast.error('Failed to verify subscription status. Please try again.')
            router.push(`/installations/new-kl-cloud?installation=${instId}`)
            return
          }
          const subData = await subRes.json()
          const hasActive = subData.subscriptions?.some(
            (s: { status: string }) => ['active', 'authenticated'].includes(s.status),
          )
          if (!hasActive) {
            toast.error('No active subscription found. Please complete payment first.')
            router.push(`/installations/new-kl-cloud?installation=${instId}`)
            return
          }
        } catch {
          toast.error('Failed to verify subscription. Please try again.')
          router.push(`/installations/new-kl-cloud?installation=${instId}`)
          return
        }

        // Check if a job is already running before triggering
        try {
          const statusRes = await fetch(`/api/installations/${instId}/job-status`)
          if (statusRes.ok) {
            const statusData = await statusRes.json()
            if (statusData.status === 'running' || statusData.status === 'pending') {
              setStatus(statusData.status)
              return
            }
            if (statusData.status === 'succeeded') {
              setStatus('succeeded')
              return
            }
          }
        } catch {
          // Ignore status check errors, proceed to trigger
        }

        await triggerDeploy(instId)
      } catch (err) {
        console.error('Error initializing:', err)
        setErrorMessage('Failed to initialize deployment')
        setStatus('error')
      }
    }

    init()
  }, [router, triggerDeploy])

  // Poll job status every 5s when running
  useEffect(() => {
    if (!installationId || (status !== 'running' && status !== 'pending')) return

    const pollStatus = async () => {
      try {
        const response = await fetch(`/api/installations/${installationId}/job-status`)
        if (!response.ok) return

        const data = await response.json()

        if (data.currentStep != null) setCurrentStep(data.currentStep)
        if (data.totalSteps != null) setTotalSteps(data.totalSteps)
        if (data.stepDescription) setStepDescription(data.stepDescription)

        if (data.status === 'succeeded') {
          setStatus('succeeded')
        } else if (data.status === 'failed' || data.status === 'unknown') {
          setErrorMessage(data.error || 'Deployment failed')
          setStatus('failed')
        } else if (data.status === 'running' || data.status === 'pending') {
          setStatus(data.status)
        }
      } catch (err) {
        console.error('Error polling status:', err)
      }
    }

    pollStatus()
    const interval = setInterval(pollStatus, 5000)
    return () => clearInterval(interval)
  }, [installationId, status])

  // Poll check-installation in parallel for deployment verification
  useEffect(() => {
    if (!installationId || status === 'succeeded') return

    const sessionCheck = async () => {
      try {
        const sessionRes = await fetch('/api/installations/session')
        if (!sessionRes.ok) return
        const sessionData = await sessionRes.json()
        if (!sessionData.installationKey) return

        const response = await fetch('/api/installations/check-installation', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ installationKey: sessionData.installationKey }),
        })
        const data = await response.json()

        if (data.verified && data.dnsConfigured) {
          setStatus('succeeded')
        }
      } catch {
        // Ignore errors in background check
      }
    }

    const interval = setInterval(sessionCheck, 5000)
    return () => clearInterval(interval)
  }, [installationId, status])

  // Auto-redirect on success
  useEffect(() => {
    if (status !== 'succeeded') return
    const timer = setTimeout(() => {
      router.push('/installations/new/complete')
    }, 2000)
    return () => clearTimeout(timer)
  }, [status, router])

  const handleRetry = () => {
    if (installationId) {
      triggerDeploy(installationId)
    }
  }

  if (status === 'loading') {
    return (
      <div className="flex items-center justify-center py-32">
        <Loader2 className="text-primary size-8 animate-spin" />
      </div>
    )
  }

  return (
    <div className="lg:flex lg:gap-12">
      {/* Left Column - Information */}
      <div className="hidden lg:block lg:w-[400px] lg:flex-shrink-0">
        <div className="sticky top-6 space-y-6">
          <div className="absolute -top-4 -right-4 -z-10 opacity-5">
            <svg width="300" height="300" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14h-2v-2h2v2zm0-4h-2V7h2v6z"/>
            </svg>
          </div>

          {/* What's Next Card */}
          <div className="border border-foreground/10 rounded-lg p-6 bg-muted/20">
            <h3 className="text-sm font-semibold text-foreground mb-3">What happens next?</h3>
            <div className="space-y-3">
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary/10 text-primary flex items-center justify-center text-xs font-semibold">
                  <CheckCircle2 className="w-3 h-3" />
                </div>
                <div>
                  <p className="text-sm font-medium text-foreground">Configure & subscribe</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Installation details, plan & payment</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-semibold">2</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Deploy to cloud</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Automatically setting up your infrastructure</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-semibold">3</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Verify & complete</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Confirm your installation is ready</p>
                </div>
              </div>
            </div>
          </div>

          {/* Tips Card */}
          <div className="border border-foreground/10 rounded-lg p-6 bg-background">
            <h3 className="text-sm font-semibold text-foreground mb-3">Quick Tips</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&bull;</span>
                <span>No CLI or cloud credentials needed</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&bull;</span>
                <span>Installation takes approximately 10-15 minutes</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&bull;</span>
                <span>Keep this window open while the deployment runs</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      {/* Right Column - Main Content */}
      <div className="space-y-6 lg:flex-1 lg:min-w-0">
        {/* Header */}
        <div>
          <h1 className="text-foreground text-2xl font-semibold tracking-tight">
            Deploying Kloudlite Cloud
          </h1>
          <p className="text-muted-foreground mt-1 text-sm">
            We&apos;re setting up your managed infrastructure automatically
          </p>
        </div>

        {/* Progress Card */}
        <div className="border border-foreground/10 rounded-lg bg-background">
          <div className="p-8">
            {(status === 'triggering' || status === 'pending' || status === 'running') && (
              <div className="text-center space-y-4">
                <div className="flex justify-center">
                  <div className="flex size-16 items-center justify-center bg-blue-100 dark:bg-blue-900/30 rounded-full">
                    <Loader2 className="size-8 text-blue-600 dark:text-blue-400 animate-spin" />
                  </div>
                </div>
                <div>
                  <h2 className="text-xl font-semibold text-foreground">
                    Setting up your Kloudlite Cloud installation...
                  </h2>
                  <p className="text-muted-foreground text-sm mt-2">
                    This process typically takes 10-15 minutes. You can safely keep this window open.
                  </p>
                </div>

                {/* Progress bar */}
                <div className="max-w-md mx-auto space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">
                      {currentStep > 0 ? `Step ${currentStep} of ${totalSteps}` : 'Starting...'}
                    </span>
                    <span className="text-muted-foreground font-medium">
                      {totalSteps > 0 ? `${Math.round((currentStep / totalSteps) * 100)}%` : '0%'}
                    </span>
                  </div>
                  <div className="h-2 bg-foreground/[0.06] rounded-full overflow-hidden">
                    <div
                      className="h-full bg-blue-600 dark:bg-blue-500 rounded-full transition-all duration-500 ease-out"
                      style={{ width: `${totalSteps > 0 ? (currentStep / totalSteps) * 100 : 0}%` }}
                    />
                  </div>
                  {stepDescription && (
                    <p className="text-xs text-muted-foreground truncate">
                      {stepDescription}
                    </p>
                  )}
                </div>
              </div>
            )}

            {status === 'succeeded' && (
              <div className="text-center space-y-4">
                <div className="flex justify-center">
                  <div className="flex size-16 items-center justify-center bg-green-100 dark:bg-green-900/30 rounded-full">
                    <CheckCircle2 className="size-8 text-green-600 dark:text-green-400" />
                  </div>
                </div>
                <div>
                  <h2 className="text-xl font-semibold text-foreground">
                    Installation complete!
                  </h2>
                  <p className="text-muted-foreground text-sm mt-2">
                    Redirecting you to the completion page...
                  </p>
                </div>
              </div>
            )}

            {(status === 'failed' || status === 'error') && (
              <div className="text-center space-y-4">
                <div className="flex justify-center">
                  <div className="flex size-16 items-center justify-center bg-red-100 dark:bg-red-900/30 rounded-full">
                    <XCircle className="size-8 text-red-600 dark:text-red-400" />
                  </div>
                </div>
                <div>
                  <h2 className="text-xl font-semibold text-foreground">
                    Deployment failed
                  </h2>
                  {errorMessage && (
                    <p className="text-muted-foreground text-sm mt-2">
                      {errorMessage}
                    </p>
                  )}
                </div>
                <Button onClick={handleRetry} variant="default">
                  <RefreshCw className="mr-2 size-4" />
                  Retry Deployment
                </Button>
              </div>
            )}
          </div>
        </div>

        {/* Status indicator */}
        <div className="flex items-center justify-center gap-3 text-base">
          {(status === 'triggering' || status === 'pending' || status === 'running') && (
            <>
              <Loader2 className="size-4 animate-spin text-blue-600" />
              <span className="text-muted-foreground">
                {currentStep > 0
                  ? `Step ${currentStep} of ${totalSteps} — ${stepDescription || 'In progress...'}`
                  : 'Deployment in progress...'}
              </span>
            </>
          )}
          {status === 'succeeded' && (
            <>
              <CheckCircle2 className="size-4 text-green-600" />
              <span className="text-green-600 font-medium">Installation complete! Redirecting...</span>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
