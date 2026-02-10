'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2 } from 'lucide-react'

interface JobStatus {
  status: string
  operation?: string
  currentStep?: number
  totalSteps?: number
  stepDescription?: string
  error?: string
}

interface InstallationJobProgressProps {
  installationId: string
  initialActive: boolean
}

export function InstallationJobProgress({ installationId, initialActive }: InstallationJobProgressProps) {
  const router = useRouter()
  const [jobStatus, setJobStatus] = useState<JobStatus | null>(null)
  const [active, setActive] = useState(initialActive)

  const fetchStatus = useCallback(async () => {
    try {
      const res = await fetch(`/api/installations/${installationId}/job-status`)
      if (!res.ok) {
        // Record may have been deleted (e.g. auto-delete after uninstall)
        setActive(false)
        router.refresh()
        return
      }
      const data: JobStatus = await res.json()
      setJobStatus(data)

      if (data.status === 'succeeded' || data.status === 'failed') {
        setActive(false)
        router.refresh()
      }
    } catch {
      // ignore
    }
  }, [installationId, router])

  useEffect(() => {
    if (!active) return
    fetchStatus()
    const interval = setInterval(fetchStatus, 5000)
    return () => clearInterval(interval)
  }, [active, fetchStatus])

  if (!active && !jobStatus) return null
  if (jobStatus && (jobStatus.status === 'succeeded' || jobStatus.status === 'failed')) return null

  const step = jobStatus?.currentStep || 0
  const total = jobStatus?.totalSteps || 1
  const percent = Math.round((step / total) * 100)
  const operation = jobStatus?.operation === 'uninstall' ? 'Uninstalling' : 'Installing'

  return (
    <div className="border border-blue-500/20 rounded-lg p-4 bg-blue-500/5">
      <div className="flex items-center gap-3 mb-3">
        <Loader2 className="h-4 w-4 animate-spin text-blue-600 dark:text-blue-400" />
        <span className="text-sm font-medium text-foreground">
          {operation}... {jobStatus?.stepDescription || ''}
        </span>
      </div>
      <div className="flex items-center gap-3">
        <div className="flex-1 bg-muted rounded-full h-1.5">
          <div
            className="bg-blue-600 dark:bg-blue-400 h-1.5 rounded-full transition-all duration-500"
            style={{ width: `${percent}%` }}
          />
        </div>
        <span className="text-xs text-muted-foreground whitespace-nowrap">
          {step}/{total}
        </span>
      </div>
    </div>
  )
}
