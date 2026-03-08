'use client'

import { useEffect, useState, useCallback, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { ExternalLink, Loader2, Copy, CheckCircle2, Clock, AlertCircle, XCircle, RotateCcw } from 'lucide-react'
import { toast } from 'sonner'

interface JobStatus {
  status: string
  error?: string
  operation?: string
  currentStep?: number
  totalSteps?: number
  stepDescription?: string
}

type ActiveStatus = 'checking' | 'provisioning' | 'waiting' | 'active' | 'error'

export interface CompletionStatusProps {
  subdomain: string
  url: string
  installationId: string
  cloudProvider?: string
}

export function CompletionStatus({ subdomain, url, installationId, cloudProvider }: CompletionStatusProps) {
  const router = useRouter()
  const [activeStatus, setActiveStatus] = useState<ActiveStatus>('checking')
  const [checkCount, setCheckCount] = useState(0)
  const [jobStatus, setJobStatus] = useState<JobStatus | null>(null)
  const [errorMessage, setErrorMessage] = useState<string>('')
  const [retrying, setRetrying] = useState(false)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const checkJobStatus = useCallback(async (): Promise<JobStatus | null> => {
    try {
      const response = await fetch(`/api/installations/${installationId}/job-status`)
      if (response.status === 404) {
        return null // No job exists
      }
      if (response.ok) {
        return await response.json()
      }
      return null
    } catch {
      return null
    }
  }, [installationId])

  const checkActiveStatus = useCallback(async (): Promise<boolean> => {
    try {
      const response = await fetch(`/api/installations/${installationId}/ping`)
      if (response.ok) {
        const data = await response.json()
        if (data.active) {
          setActiveStatus('active')
          return true
        }
      }
      return false
    } catch {
      return false
    }
  }, [installationId])

  // Combined polling: check job status + ping
  const pollStatus = useCallback(async () => {
    setCheckCount(c => c + 1)

    // Check job status first
    const job = await checkJobStatus()
    setJobStatus(job)

    if (job) {
      if (job.status === 'failed') {
        setActiveStatus('error')
        setErrorMessage(job.error || 'Installation job failed')
        return 'stop'
      }
      if (job.status === 'running' || job.status === 'pending') {
        setActiveStatus('provisioning')
        return 'continue'
      }
    }

    // Job succeeded or no job — check if installation is reachable
    const isActive = await checkActiveStatus()
    if (isActive) return 'stop'

    // If no job ever existed and ping fails, it's an error
    if (!job) {
      setActiveStatus('error')
      setErrorMessage('No installation job was found. The installation may not have been provisioned.')
      return 'stop'
    }

    // Job succeeded but not reachable yet — waiting for DNS
    setActiveStatus('waiting')
    return 'continue'
  }, [checkJobStatus, checkActiveStatus])

  // Initial status check
  useEffect(() => {
    pollStatus()
  }, [pollStatus])

  // Poll every 5 seconds until active or error
  useEffect(() => {
    if (activeStatus === 'active' || activeStatus === 'error') {
      return
    }

    intervalRef.current = setInterval(async () => {
      const result = await pollStatus()
      if (result === 'stop' && intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }, 5000)

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [activeStatus, pollStatus])

  const handleRetry = async () => {
    setRetrying(true)
    try {
      const response = await fetch(`/api/installations/${installationId}/trigger-managed-install`, {
        method: 'POST',
      })
      if (response.ok) {
        setActiveStatus('provisioning')
        setErrorMessage('')
        setJobStatus(null)
        setCheckCount(0)
        toast.success('Installation job triggered')
      } else {
        const data = await response.json()
        toast.error(data.error || 'Failed to retry installation')
      }
    } catch {
      toast.error('Failed to retry installation')
    } finally {
      setRetrying(false)
    }
  }

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`${label} copied to clipboard`)
  }

  const isActive = activeStatus === 'active'
  const isError = activeStatus === 'error'
  const isProvisioning = activeStatus === 'provisioning'
  const progressPercent = jobStatus?.currentStep && jobStatus?.totalSteps
    ? Math.round((jobStatus.currentStep / jobStatus.totalSteps) * 100)
    : 0

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
                <div className={`flex-shrink-0 w-5 h-5 rounded-full flex items-center justify-center text-xs font-semibold ${
                  isError ? 'bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400' :
                  isProvisioning ? 'bg-primary text-primary-foreground' :
                  'bg-primary/10 text-primary'
                }`}>
                  {isError ? <XCircle className="w-3 h-3" /> :
                   isProvisioning ? '2' :
                   <CheckCircle2 className="w-3 h-3" />}
                </div>
                <div>
                  <p className="text-sm font-medium text-foreground">Deploy to cloud</p>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {isError ? 'Deployment failed' :
                     isProvisioning && jobStatus?.stepDescription ? jobStatus.stepDescription :
                     'Install Kloudlite in your infrastructure'}
                  </p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className={`flex-shrink-0 w-5 h-5 rounded-full flex items-center justify-center text-xs font-semibold ${
                  isActive ? 'bg-primary/10 text-primary' :
                  isError ? 'bg-muted text-muted-foreground' :
                  (!isProvisioning && !isError) ? 'bg-primary text-primary-foreground' :
                  'bg-muted text-muted-foreground'
                }`}>
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
            {isActive ? 'Installation Complete!' :
             isError ? 'Installation Failed' :
             'Setting Up Your Installation'}
          </h1>
          <p className="text-muted-foreground mt-1 text-sm">
            {isActive ? 'Your Kloudlite installation is ready to use' :
             isError ? 'Something went wrong during installation' :
             'Please wait while your installation becomes active'}
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
                        href={url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary flex items-center gap-2 font-mono text-lg hover:underline"
                      >
                        {subdomain}.
                        {process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                        <ExternalLink className="size-4" />
                      </a>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => copyToClipboard(url, 'URL')}
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
                      onClick={() => router.push(`/installations/${installationId}`)}
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
            ) : isError ? (
              <>
                <div className="mb-6">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="flex size-10 items-center justify-center bg-red-100 dark:bg-red-900/30 rounded-full">
                      <XCircle className="size-5 text-red-600 dark:text-red-400" />
                    </div>
                    <h2 className="text-xl font-semibold text-foreground">Installation Failed</h2>
                  </div>
                  <p className="text-muted-foreground text-sm">
                    The installation could not be completed. You can retry or go back to your installations.
                  </p>
                </div>

                <div className="space-y-4">
                  <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-900 p-4 rounded-sm">
                    <div className="flex items-start gap-3">
                      <AlertCircle className="size-5 text-red-600 dark:text-red-400 mt-0.5 flex-shrink-0" />
                      <div>
                        <p className="text-sm font-medium text-red-900 dark:text-red-200 mb-1">
                          Error Details
                        </p>
                        <p className="text-xs text-red-800 dark:text-red-300 font-mono whitespace-pre-wrap">
                          {errorMessage}
                        </p>
                      </div>
                    </div>
                  </div>

                  {cloudProvider === 'oci' && (
                    <Button
                      size="lg"
                      className="w-full"
                      onClick={handleRetry}
                      disabled={retrying}
                    >
                      {retrying ? (
                        <Loader2 className="mr-2 size-4 animate-spin" />
                      ) : (
                        <RotateCcw className="mr-2 size-4" />
                      )}
                      {retrying ? 'Retrying...' : 'Retry Installation'}
                    </Button>
                  )}

                  <div className="flex gap-3">
                    <Button
                      variant="outline"
                      size="lg"
                      className="flex-1"
                      onClick={() => router.push(`/installations/${installationId}`)}
                    >
                      View Installation Details
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
                </div>
              </>
            ) : (
              <>
                <div className="mb-6">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="flex size-10 items-center justify-center bg-blue-100 dark:bg-blue-900/30 rounded-full">
                      <Clock className="size-5 text-blue-600 dark:text-blue-400 animate-pulse" />
                    </div>
                    <h2 className="text-xl font-semibold text-foreground">
                      {isProvisioning ? 'Provisioning Your Installation' : 'Waiting for Installation to Become Active'}
                    </h2>
                  </div>
                  <p className="text-muted-foreground text-sm">
                    {isProvisioning
                      ? 'Your cloud infrastructure is being set up.'
                      : 'Your installation is being set up. This usually takes 1-3 minutes.'}
                  </p>
                </div>

                <div className="space-y-4">
                  {/* Progress bar for provisioning */}
                  {isProvisioning && jobStatus?.totalSteps && (
                    <div className="space-y-2">
                      <div className="flex justify-between text-xs text-muted-foreground">
                        <span>Step {jobStatus.currentStep || 0} of {jobStatus.totalSteps}</span>
                        <span>{progressPercent}%</span>
                      </div>
                      <div className="w-full bg-muted rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full transition-all duration-500"
                          style={{ width: `${progressPercent}%` }}
                        />
                      </div>
                      {jobStatus.stepDescription && (
                        <p className="text-sm text-muted-foreground">{jobStatus.stepDescription}</p>
                      )}
                    </div>
                  )}

                  <div className="bg-foreground/[0.02] p-4 rounded-sm border border-border/60">
                    <p className="mb-2 text-sm font-medium text-foreground">Installation Dashboard URL:</p>
                    <div className="flex items-center justify-between gap-3">
                      <span className="text-muted-foreground font-mono text-lg">
                        {subdomain}.
                        {process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                      </span>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => copyToClipboard(url, 'URL')}
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
                          {isProvisioning ? 'Provisioning in Progress' : 'Installation in Progress'}
                        </p>
                        <p className="text-xs text-blue-800 dark:text-blue-300">
                          {isProvisioning
                            ? 'Cloud resources are being created. This may take a few minutes.'
                            : `We're checking if your installation is ready. Checked ${checkCount} time${checkCount !== 1 ? 's' : ''}. The page will automatically update when your installation is active.`}
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
          {isProvisioning && (
            <>
              <Loader2 className="size-4 animate-spin text-blue-600" />
              <span className="text-muted-foreground">
                Provisioning{jobStatus?.stepDescription ? `: ${jobStatus.stepDescription}` : '...'}
              </span>
            </>
          )}
          {activeStatus === 'waiting' && (
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
          {isError && (
            <>
              <XCircle className="size-4 text-red-600" />
              <span className="text-red-600 font-medium">Installation failed</span>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
