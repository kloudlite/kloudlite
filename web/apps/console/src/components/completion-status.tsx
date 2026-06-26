'use client'

import { useEffect, useState, useCallback, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { ExternalLink, Loader2, CheckCircle2, Clock, XCircle, RotateCcw } from 'lucide-react'
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
  installationId: string
  cloudProvider?: string
}

export function CompletionStatus({ installationId, cloudProvider: _cloudProvider }: CompletionStatusProps) {
  const router = useRouter()
  const [activeStatus, setActiveStatus] = useState<ActiveStatus>('checking')
  const [jobStatus, setJobStatus] = useState<JobStatus | null>(null)
  const [errorMessage, setErrorMessage] = useState<string>('')
  const [retrying, setRetrying] = useState(false)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const checkJobStatus = useCallback(async (): Promise<JobStatus | null> => {
    try {
      const response = await fetch(`/api/installations/${installationId}/job-status`)
      if (response.status === 404) return null
      if (!response.ok) return null
      return response.json()
    } catch {
      return null
    }
  }, [installationId])

  // ... rest of the polling logic remains the same
  useEffect(() => {
    async function poll() {
      const status = await checkJobStatus()
      if (!status) return

      setJobStatus(status)

      if (status.status === 'succeeded') {
        setActiveStatus('active')
        clearInterval(intervalRef.current!)
        return
      }
      if (status.status === 'failed') {
        setActiveStatus('error')
        setErrorMessage(status.error || 'Installation failed')
        clearInterval(intervalRef.current!)
        return
      }
      if (status.status === 'running' || status.status === 'pending') {
        setActiveStatus('provisioning')
        if (status.operation === 'uninstall') {
          setActiveStatus('waiting')
        }
      }
    }

    poll()
    intervalRef.current = setInterval(poll, 5000)
    return () => clearInterval(intervalRef.current!)
  }, [checkJobStatus])

  async function handleRetry() {
    setRetrying(true)
    try {
      const response = await fetch(`/api/installations/${installationId}/trigger-managed-install`, {
        method: 'POST',
      })
      const data = await response.json()
      if (!response.ok) {
        setErrorMessage(data.error || 'Failed to retry')
        return
      }
      setErrorMessage('')
      setActiveStatus('provisioning')
      toast.success('Installation retrying...')
    } catch {
      setErrorMessage('Failed to retry installation')
    } finally {
      setRetrying(false)
    }
  }

  return (
    <div className="border border-foreground/10 rounded-lg p-8 bg-background">
      <div className="max-w-lg mx-auto text-center space-y-8">
        {/* Check mark / spinner */}
        <div className="flex justify-center">
          <div className="size-16 rounded-full bg-primary/10 flex items-center justify-center">
            {activeStatus === 'active' ? (
              <CheckCircle2 className="size-8 text-green-600" />
            ) : activeStatus === 'error' ? (
              <XCircle className="size-8 text-destructive" />
            ) : activeStatus === 'waiting' ? (
              <Clock className="size-8 text-yellow-600" />
            ) : (
              <Loader2 className="size-8 text-primary animate-spin" />
            )}
          </div>
        </div>

        <div className="space-y-2">
          <h2 className="text-foreground text-xl font-semibold">
            {activeStatus === 'active' && 'Installation Complete'}
            {activeStatus === 'error' && 'Installation Failed'}
            {activeStatus === 'provisioning' && 'Provisioning Your Installation'}
            {activeStatus === 'waiting' && 'Waiting for Uninstall'}
            {activeStatus === 'checking' && 'Checking Status...'}
          </h2>
          <p className="text-muted-foreground text-sm">
            {activeStatus === 'provisioning' && jobStatus?.stepDescription
              ? `Step ${jobStatus.currentStep} of ${jobStatus.totalSteps}: ${jobStatus.stepDescription}`
              : activeStatus === 'provisioning'
              ? 'Setting up your infrastructure on OCI...'
              : activeStatus === 'active'
              ? 'Your installation is ready.'
              : activeStatus === 'error'
              ? errorMessage || 'An error occurred during installation.'
              : ''}
          </p>
        </div>

        {/* Steps progress */}
        {activeStatus === 'provisioning' && jobStatus && (
          <div className="space-y-3">
            <div className="h-2 bg-foreground/[0.06] rounded-full overflow-hidden">
              <div
                className="h-full bg-primary rounded-full transition-all duration-500"
                style={{ width: `${((jobStatus.currentStep || 0) / (jobStatus.totalSteps || 9)) * 100}%` }}
              />
            </div>
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Progress</span>
              <span>{Math.round(((jobStatus.currentStep || 0) / (jobStatus.totalSteps || 9)) * 100)}%</span>
            </div>
          </div>
        )}

        {/* View button */}
        {activeStatus === 'active' && (
          <Button onClick={() => router.push(`/installations/${installationId}`)} size="lg">
            View Installation
          </Button>
        )}

        {/* Error actions */}
        {activeStatus === 'error' && (
          <div className="space-y-3">
            <Button onClick={handleRetry} disabled={retrying} className="w-full" size="lg">
              {retrying ? <Loader2 className="size-4 animate-spin mr-2" /> : <RotateCcw className="size-4 mr-2" />}
              Retry Installation
            </Button>
            <Button variant="outline" onClick={() => router.push(`/installations/${installationId}`)} className="w-full" size="lg">
              Go to Installation
            </Button>
          </div>
        )}

        {/* Uninstall waiting */}
        {activeStatus === 'waiting' && (
          <div className="space-y-3">
            <Button variant="outline" onClick={() => router.push(`/installations/${installationId}`)} className="w-full" size="lg">
              <ExternalLink className="size-4 mr-2" />
              View Installation
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
