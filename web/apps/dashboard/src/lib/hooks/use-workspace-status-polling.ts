'use client'

import { useState, useEffect, useRef } from 'react'
import { getWorkspaceStatus } from '@/app/actions/workspace-status.actions'

export interface WorkspaceStatus {
  name: string
  namespace: string
  phase: string
  state: string
  isReady: boolean
  message?: string
  podName?: string
  conditions: Array<{
    type: string
    status: string
    message?: string
  }>
  lastUpdated: string
}

interface UseWorkspaceStatusPollingOptions {
  interval?: number // Polling interval in milliseconds (default: 5000)
  enabled?: boolean // Whether polling is enabled (default: true)
  onError?: (error: string) => void
}

/**
 * Hook for polling workspace status
 * Replaces WebSocket-based status streaming with polling
 */
export function useWorkspaceStatusPolling(
  namespace: string,
  workspaceName: string,
  options: UseWorkspaceStatusPollingOptions = {}
) {
  const { interval = 5000, enabled = true, onError } = options

  const [status, setStatus] = useState<WorkspaceStatus | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<NodeJS.Timeout>()

  useEffect(() => {
    if (!enabled || !namespace || !workspaceName) {
      return
    }

    // Fetch immediately on mount
    const fetchStatus = async () => {
      try {
        const result = await getWorkspaceStatus(namespace, workspaceName)

        if (result.success && result.data) {
          setStatus(result.data as WorkspaceStatus)
          setError(null)
        } else {
          setError(result.error || 'Failed to fetch workspace status')
          onError?.(result.error || 'Failed to fetch workspace status')
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Unknown error'
        setError(errorMessage)
        onError?.(errorMessage)
      } finally {
        setIsLoading(false)
      }
    }

    // Initial fetch
    fetchStatus()

    // Set up polling interval
    intervalRef.current = setInterval(fetchStatus, interval)

    // Cleanup
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [namespace, workspaceName, interval, enabled, onError])

  // Pause polling when tab is not visible
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden && intervalRef.current) {
        clearInterval(intervalRef.current)
      } else if (!document.hidden && enabled) {
        // Resume polling when tab becomes visible
        intervalRef.current = setInterval(async () => {
          const result = await getWorkspaceStatus(namespace, workspaceName)
          if (result.success && result.data) {
            setStatus(result.data as WorkspaceStatus)
            setError(null)
          }
        }, interval)
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [namespace, workspaceName, interval, enabled])

  return {
    status,
    isLoading,
    error,
  }
}
