'use client'

import { useEffect, useState, useRef, useCallback } from 'react'

interface WorkspaceStatusEvent {
  phase: string
  message: string
  status: string
  activeConnections: number
  idleState: string
  accessUrls?: Record<string, string>
  timestamp: string
}

interface UseWorkspaceStatusStreamOptions {
  enabled?: boolean
  onPhaseChange?: (phase: string) => void
  onReady?: () => void
  onDeleted?: () => void
}

// Maximum reconnection attempts before giving up
const MAX_RECONNECT_ATTEMPTS = 10
// Base delay for exponential backoff (ms)
const BASE_RECONNECT_DELAY = 1000
// Maximum delay between reconnections (ms)
const MAX_RECONNECT_DELAY = 30000

export function useWorkspaceStatusStream(
  namespace: string | undefined,
  workspaceName: string | undefined,
  options: UseWorkspaceStatusStreamOptions = {}
) {
  const { enabled = true, onPhaseChange, onReady, onDeleted } = options
  const [status, setStatus] = useState<WorkspaceStatusEvent | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const eventSourceRef = useRef<EventSource | null>(null)
  const previousPhaseRef = useRef<string | null>(null)
  const onReadyCalledRef = useRef(false)
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
    if (!namespace || !workspaceName || !enabled) return
    if (isCleaningUpRef.current) return

    // Close existing connection if any
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    // Create SSE connection
    const eventSource = new EventSource(
      `/api/v1/namespaces/${encodeURIComponent(namespace)}/workspaces/${encodeURIComponent(workspaceName)}/status-stream`
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
        const data: WorkspaceStatusEvent = JSON.parse(event.data)
        setStatus(data)

        // Call onPhaseChange if phase changed
        if (onPhaseChange && data.phase !== previousPhaseRef.current) {
          previousPhaseRef.current = data.phase
          onPhaseChange(data.phase)
        }

        // Call onReady when workspace becomes Running (only once per connection cycle)
        if (onReady && data.phase === 'Running' && !onReadyCalledRef.current) {
          onReadyCalledRef.current = true
          onReady()
        }
      } catch (err) {
        console.error('Failed to parse workspace status event:', err)
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
  }, [namespace, workspaceName, enabled, onPhaseChange, onReady, onDeleted])

  useEffect(() => {
    if (!enabled || !namespace || !workspaceName) {
      setStatus(null)
      setIsConnected(false)
      return
    }

    isCleaningUpRef.current = false
    onReadyCalledRef.current = false
    reconnectAttemptsRef.current = 0
    connect()

    return () => {
      cleanup()
    }
  }, [namespace, workspaceName, enabled, connect, cleanup])

  const reconnect = useCallback(() => {
    isCleaningUpRef.current = false
    reconnectAttemptsRef.current = 0
    onReadyCalledRef.current = false
    connect()
  }, [connect])

  return {
    status,
    phase: status?.phase,
    message: status?.message,
    workspaceStatus: status?.status,
    activeConnections: status?.activeConnections,
    idleState: status?.idleState,
    accessUrls: status?.accessUrls,
    isConnected,
    error,
    reconnect,
    isRunning: status?.phase === 'Running',
    isPending: status?.phase === 'Pending',
    isCreating: status?.phase === 'Creating',
    isStopping: status?.phase === 'Stopping',
    isStopped: status?.phase === 'Stopped',
    isFailed: status?.phase === 'Failed',
    isTerminating: status?.phase === 'Terminating',
    isIdle: status?.idleState === 'idle',
    isActive: status?.idleState === 'active',
  }
}
