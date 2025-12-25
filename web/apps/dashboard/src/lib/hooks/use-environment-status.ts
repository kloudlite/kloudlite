'use client'

import { useEffect, useState, useRef, useCallback } from 'react'

interface CloningStatus {
  phase: string
  message?: string
  totalPVCs?: number
  clonedPVCs?: number
  currentPVC?: string
  bytesTransferred?: number
  startTime?: string
  completionTime?: string
  failedPVCs?: string[]
}

interface EnvironmentStatusEvent {
  state: string
  message: string
  activated: boolean
  cloningStatus?: CloningStatus
  timestamp: string
}

interface UseEnvironmentStatusOptions {
  enabled?: boolean
  onStateChange?: (state: string) => void
  onDeleted?: () => void
}

// Maximum reconnection attempts before giving up
const MAX_RECONNECT_ATTEMPTS = 10
// Base delay for exponential backoff (ms)
const BASE_RECONNECT_DELAY = 1000
// Maximum delay between reconnections (ms)
const MAX_RECONNECT_DELAY = 30000

export function useEnvironmentStatus(
  envName: string | undefined,
  options: UseEnvironmentStatusOptions = {}
) {
  const { enabled = true, onStateChange, onDeleted } = options
  const [status, setStatus] = useState<EnvironmentStatusEvent | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const eventSourceRef = useRef<EventSource | null>(null)
  const previousStateRef = useRef<string | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isCleaningUpRef = useRef(false)

  const cleanup = useCallback(() => {
    isCleaningUpRef.current = true
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
  }, [])

  const connect = useCallback(() => {
    if (!envName || !enabled) return
    if (isCleaningUpRef.current) return

    // Close existing connection if any
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    // Create SSE connection
    const eventSource = new EventSource(
      `/api/v1/environments/${encodeURIComponent(envName)}/status-stream`
    )
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setIsConnected(true)
      setError(null)
      // Reset reconnect attempts on successful connection
      reconnectAttemptsRef.current = 0
    }

    eventSource.addEventListener('status', (event) => {
      try {
        const data: EnvironmentStatusEvent = JSON.parse(event.data)
        setStatus(data)

        // Call onStateChange if state changed
        if (onStateChange && data.state !== previousStateRef.current) {
          previousStateRef.current = data.state
          onStateChange(data.state)
        }
      } catch (err) {
        console.error('Failed to parse environment status event:', err)
      }
    })

    eventSource.addEventListener('deleted', () => {
      setIsConnected(false)
      eventSource.close()
      eventSourceRef.current = null
      if (onDeleted) {
        onDeleted()
      }
    })

    eventSource.onerror = () => {
      // SSE error - likely Cloudflare timeout or network issue
      setIsConnected(false)
      eventSource.close()
      eventSourceRef.current = null

      // Don't reconnect if cleaning up
      if (isCleaningUpRef.current) return

      // Attempt reconnection with exponential backoff
      if (reconnectAttemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
        const delay = Math.min(
          BASE_RECONNECT_DELAY * Math.pow(2, reconnectAttemptsRef.current),
          MAX_RECONNECT_DELAY
        )
        reconnectAttemptsRef.current++

        console.log(`SSE connection lost, reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current}/${MAX_RECONNECT_ATTEMPTS})`)

        reconnectTimeoutRef.current = setTimeout(() => {
          if (!isCleaningUpRef.current) {
            connect()
          }
        }, delay)
      } else {
        setError('Connection failed after multiple attempts')
        console.error('SSE connection failed after maximum reconnect attempts')
      }
    }
  }, [envName, enabled, onStateChange, onDeleted])

  useEffect(() => {
    if (!enabled || !envName) {
      setStatus(null)
      setIsConnected(false)
      return
    }

    isCleaningUpRef.current = false
    reconnectAttemptsRef.current = 0
    connect()

    return () => {
      cleanup()
    }
  }, [envName, enabled, connect, cleanup])

  const reconnect = useCallback(() => {
    isCleaningUpRef.current = false
    reconnectAttemptsRef.current = 0
    connect()
  }, [connect])

  return {
    status,
    state: status?.state,
    message: status?.message,
    activated: status?.activated,
    cloningStatus: status?.cloningStatus,
    isConnected,
    error,
    reconnect,
    isActive: status?.state === 'active',
    isActivating: status?.state === 'activating',
    isInactive: status?.state === 'inactive',
    isDeactivating: status?.state === 'deactivating',
    isCloning: status?.cloningStatus?.phase !== undefined && status?.cloningStatus?.phase !== 'Completed',
    isError: status?.state === 'error',
  }
}
