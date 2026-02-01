'use client'

import { useState, useEffect, useRef } from 'react'
import { getWorkMachineMetrics } from '@/app/actions/workmachine-metrics.actions'

export interface WorkMachineMetrics {
  name: string
  cpu: number
  memory: number
  network: {
    rx: number
    tx: number
  }
  disk: number
  timestamp: string
}

interface UseWorkMachineMetricsPollingOptions {
  interval?: number
  enabled?: boolean
  onError?: (error: string) => void
}

/**
 * Hook for polling WorkMachine metrics (CPU, memory, network, disk)
 */
export function useWorkMachineMetricsPolling(
  workMachineName: string,
  options: UseWorkMachineMetricsPollingOptions = {}
) {
  const { interval = 10000, enabled = true, onError } = options // Default 10s for metrics

  const [metrics, setMetrics] = useState<WorkMachineMetrics | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<NodeJS.Timeout>()

  useEffect(() => {
    if (!enabled || !workMachineName) {
      return
    }

    const fetchMetrics = async () => {
      try {
        const result = await getWorkMachineMetrics(workMachineName)

        if (result.success && result.data) {
          setMetrics(result.data as WorkMachineMetrics)
          setError(null)
        } else {
          setError(result.error || 'Failed to fetch metrics')
          onError?.(result.error || 'Failed to fetch metrics')
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Unknown error'
        setError(errorMessage)
        onError?.(errorMessage)
      } finally {
        setIsLoading(false)
      }
    }

    fetchMetrics()
    intervalRef.current = setInterval(fetchMetrics, interval)

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [workMachineName, interval, enabled, onError])

  // Pause when tab not visible
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden && intervalRef.current) {
        clearInterval(intervalRef.current)
      } else if (!document.hidden && enabled) {
        intervalRef.current = setInterval(async () => {
          const result = await getWorkMachineMetrics(workMachineName)
          if (result.success && result.data) {
            setMetrics(result.data as WorkMachineMetrics)
          }
        }, interval)
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange)
  }, [workMachineName, interval, enabled])

  return {
    metrics,
    isLoading,
    error,
  }
}
