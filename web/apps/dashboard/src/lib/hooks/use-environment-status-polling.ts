'use client'

import { useState, useEffect, useRef } from 'react'
import { getEnvironmentStatus } from '@/app/actions/environment-status.actions'

export interface EnvironmentStatus {
  name: string
  state: string
  activated: boolean
  message?: string
  conditions: Array<{
    type: string
    status: string
    message?: string
  }>
  resourceCount: {
    deployments: number
    services: number
    configmaps: number
    secrets: number
  }
  lastUpdated: string
}

interface UseEnvironmentStatusPollingOptions {
  interval?: number
  enabled?: boolean
  onError?: (error: string) => void
}

/**
 * Hook for polling environment status
 */
export function useEnvironmentStatusPolling(
  envName: string,
  options: UseEnvironmentStatusPollingOptions = {}
) {
  const { interval = 5000, enabled = true, onError } = options

  const [status, setStatus] = useState<EnvironmentStatus | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<NodeJS.Timeout | undefined>(undefined)

  useEffect(() => {
    if (!enabled || !envName) {
      return
    }

    const fetchStatus = async () => {
      try {
        const result = await getEnvironmentStatus(envName)

        if (result.success && result.data) {
          setStatus(result.data)
          setError(null)
        } else {
          setError(result.error || 'Failed to fetch environment status')
          onError?.(result.error || 'Failed to fetch environment status')
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Unknown error'
        setError(errorMessage)
        onError?.(errorMessage)
      } finally {
        setIsLoading(false)
      }
    }

    fetchStatus()
    intervalRef.current = setInterval(fetchStatus, interval)

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [envName, interval, enabled, onError])

  // Pause when tab not visible
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden && intervalRef.current) {
        clearInterval(intervalRef.current)
      } else if (!document.hidden && enabled) {
        intervalRef.current = setInterval(async () => {
          const result = await getEnvironmentStatus(envName)
          if (result.success && result.data) {
            setStatus(result.data as EnvironmentStatus)
          }
        }, interval)
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange)
  }, [envName, interval, enabled])

  return {
    status,
    isLoading,
    error,
  }
}
