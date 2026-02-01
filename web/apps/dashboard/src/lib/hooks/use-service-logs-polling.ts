'use client'

import { useState, useEffect, useRef } from 'react'
import { getServiceLogs } from '@/app/actions/service-logs.actions'

export interface ServiceLogs {
  podName: string
  logs: string
  timestamp: string
}

interface UseServiceLogsPollingOptions {
  interval?: number
  enabled?: boolean
  tailLines?: number
  onError?: (error: string) => void
}

/**
 * Hook for polling service logs
 */
export function useServiceLogsPolling(
  namespace: string,
  serviceName: string,
  options: UseServiceLogsPollingOptions = {}
) {
  const { interval = 3000, enabled = true, tailLines = 100, onError } = options

  const [logs, setLogs] = useState<ServiceLogs | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<NodeJS.Timeout>()

  useEffect(() => {
    if (!enabled || !namespace || !serviceName) {
      return
    }

    const fetchLogs = async () => {
      try {
        const result = await getServiceLogs(namespace, serviceName, { tailLines })

        if (result.success && result.data) {
          setLogs(result.data as ServiceLogs)
          setError(null)
        } else {
          setError(result.error || 'Failed to fetch logs')
          onError?.(result.error || 'Failed to fetch logs')
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Unknown error'
        setError(errorMessage)
        onError?.(errorMessage)
      } finally {
        setIsLoading(false)
      }
    }

    fetchLogs()
    intervalRef.current = setInterval(fetchLogs, interval)

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [namespace, serviceName, interval, enabled, tailLines, onError])

  // Pause when tab not visible
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden && intervalRef.current) {
        clearInterval(intervalRef.current)
      } else if (!document.hidden && enabled) {
        intervalRef.current = setInterval(async () => {
          const result = await getServiceLogs(namespace, serviceName, { tailLines })
          if (result.success && result.data) {
            setLogs(result.data as ServiceLogs)
          }
        }, interval)
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange)
  }, [namespace, serviceName, interval, enabled, tailLines])

  return {
    logs,
    isLoading,
    error,
  }
}
