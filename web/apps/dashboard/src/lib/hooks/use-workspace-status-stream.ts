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

  const connect = useCallback(() => {
    if (!namespace || !workspaceName || !enabled) return

    // Close existing connection if any
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    // Reset onReady flag for new connection
    onReadyCalledRef.current = false

    // Create SSE connection
    const eventSource = new EventSource(
      `/api/v1/namespaces/${encodeURIComponent(namespace)}/workspaces/${encodeURIComponent(workspaceName)}/status-stream`
    )
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setIsConnected(true)
      setError(null)
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

        // Call onReady when workspace becomes Running (only once)
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

    eventSource.onerror = (err) => {
      console.error('SSE connection error:', err)
      setError('Connection error')
      setIsConnected(false)
      eventSource.close()
      eventSourceRef.current = null
    }
  }, [namespace, workspaceName, enabled, onPhaseChange, onReady, onDeleted])

  useEffect(() => {
    if (!enabled || !namespace || !workspaceName) {
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
  }, [namespace, workspaceName, enabled, connect])

  const reconnect = useCallback(() => {
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
