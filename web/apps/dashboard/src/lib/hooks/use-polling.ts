'use client'

import { useState, useEffect, useRef, useCallback } from 'react'

export interface UsePollingOptions<T> {
  /** Polling interval in milliseconds (default: 5000) */
  interval?: number
  /** Whether polling is enabled (default: true) */
  enabled?: boolean
  /** Callback when an error occurs */
  onError?: (error: string) => void
  /** Callback when data is successfully fetched */
  onSuccess?: (data: T) => void
  /** Whether to pause polling when tab is not visible (default: true) */
  pauseOnHidden?: boolean
}

export interface UsePollingResult<T> {
  data: T | null
  isLoading: boolean
  error: string | null
  /** Manually trigger a refresh */
  refresh: () => Promise<void>
}

/**
 * Generic polling hook with visibility-aware pause/resume
 *
 * @param fetcher - Async function that fetches the data
 * @param deps - Dependencies array that triggers refetch when changed
 * @param options - Polling options
 *
 * @example
 * ```tsx
 * const { data, isLoading, error } = usePolling(
 *   async () => {
 *     const result = await fetchData(id)
 *     if (!result.success) throw new Error(result.error)
 *     return result.data
 *   },
 *   [id],
 *   { interval: 5000 }
 * )
 * ```
 */
export function usePolling<T>(
  fetcher: () => Promise<T>,
  deps: React.DependencyList,
  options: UsePollingOptions<T> = {}
): UsePollingResult<T> {
  const {
    interval = 5000,
    enabled = true,
    onError,
    onSuccess,
    pauseOnHidden = true,
  } = options

  const [data, setData] = useState<T | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<NodeJS.Timeout | undefined>(undefined)
  const fetcherRef = useRef(fetcher)

  // Keep fetcher ref up to date
  useEffect(() => {
    fetcherRef.current = fetcher
  }, [fetcher])

  const fetchData = useCallback(async () => {
    try {
      const result = await fetcherRef.current()
      setData(result)
      setError(null)
      onSuccess?.(result)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error'
      setError(errorMessage)
      onError?.(errorMessage)
    } finally {
      setIsLoading(false)
    }
  }, [onError, onSuccess])

  // Main polling effect
  useEffect(() => {
    if (!enabled) {
      return
    }

    // Initial fetch
    fetchData()

    // Set up polling interval
    intervalRef.current = setInterval(fetchData, interval)

    // Cleanup
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [interval, enabled, fetchData, ...deps])

  // Visibility change handling
  useEffect(() => {
    if (!pauseOnHidden) {
      return
    }

    const handleVisibilityChange = () => {
      if (document.hidden) {
        // Pause polling when tab is hidden
        if (intervalRef.current) {
          clearInterval(intervalRef.current)
          intervalRef.current = undefined
        }
      } else if (enabled) {
        // Resume polling when tab becomes visible
        // Fetch immediately and restart interval
        fetchData()
        intervalRef.current = setInterval(fetchData, interval)
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [enabled, interval, pauseOnHidden, fetchData])

  const refresh = useCallback(async () => {
    setIsLoading(true)
    await fetchData()
  }, [fetchData])

  return {
    data,
    isLoading,
    error,
    refresh,
  }
}
