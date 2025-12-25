'use client'

import { useState, useRef, useCallback, useMemo } from 'react'
import { useSSE } from './use-sse'

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

  const previousPhaseRef = useRef<string | null>(null)
  const onReadyCalledRef = useRef(false)

  // Store callbacks in refs to keep eventHandlers stable
  const onPhaseChangeRef = useRef(onPhaseChange)
  const onReadyRef = useRef(onReady)
  const onDeletedRef = useRef(onDeleted)
  onPhaseChangeRef.current = onPhaseChange
  onReadyRef.current = onReady
  onDeletedRef.current = onDeleted

  const url = useMemo(() => {
    if (!namespace || !workspaceName) return null
    return `/api/v1/namespaces/${encodeURIComponent(namespace)}/workspaces/${encodeURIComponent(workspaceName)}/status-stream`
  }, [namespace, workspaceName])

  const handleStatus = useCallback((data: WorkspaceStatusEvent) => {
    setStatus(data)

    // Call onPhaseChange if phase changed
    if (onPhaseChangeRef.current && data.phase !== previousPhaseRef.current) {
      previousPhaseRef.current = data.phase
      onPhaseChangeRef.current(data.phase)
    }

    // Call onReady when workspace becomes Running (only once per connection cycle)
    if (onReadyRef.current && data.phase === 'Running' && !onReadyCalledRef.current) {
      onReadyCalledRef.current = true
      onReadyRef.current()
    }
  }, [])

  const handleDeleted = useCallback(() => {
    onDeletedRef.current?.()
  }, [])

  const eventHandlers = useMemo(
    () => ({
      status: handleStatus,
      deleted: handleDeleted,
    }),
    [handleStatus, handleDeleted]
  )

  const handleOpen = useCallback(() => {
    onReadyCalledRef.current = false
  }, [])

  const { isConnected, error, reconnect } = useSSE<WorkspaceStatusEvent>(url, {
    enabled,
    eventHandlers,
    onOpen: handleOpen,
  })

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
