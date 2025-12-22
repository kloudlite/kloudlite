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

  const connect = useCallback(() => {
    if (!envName || !enabled) return

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

    eventSource.onerror = (err) => {
      console.error('SSE connection error:', err)
      setError('Connection error')
      setIsConnected(false)
      eventSource.close()
      eventSourceRef.current = null
    }
  }, [envName, enabled, onStateChange, onDeleted])

  useEffect(() => {
    if (!enabled || !envName) {
      setStatus(null)
      setIsConnected(false)
      return
    }

    connect()

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
        eventSourceRef.current = null
      }
    }
  }, [envName, enabled, connect])

  const reconnect = useCallback(() => {
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
