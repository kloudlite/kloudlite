'use client'

import { useState, useRef, useCallback, useMemo } from 'react'
import { useWebSocket } from './use-websocket'

interface ForkingStatus {
  phase: string
  message?: string
  totalPVCs?: number
  forkedPVCs?: number
  currentPVC?: string
  bytesTransferred?: number
  startTime?: string
  completionTime?: string
  failedPVCs?: string[]
}

interface EnvironmentStatusEvent {
  type?: string
  state: string
  message: string
  activated: boolean
  forkingStatus?: ForkingStatus
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

  const previousStateRef = useRef<string | null>(null)

  // Store callbacks in refs to keep eventHandlers stable
  const onStateChangeRef = useRef(onStateChange)
  const onDeletedRef = useRef(onDeleted)
  onStateChangeRef.current = onStateChange
  onDeletedRef.current = onDeleted

  const url = useMemo(() => {
    if (!envName) return null
    // Use WebSocket endpoint
    return `/api/v1/environments/${encodeURIComponent(envName)}/status-ws`
  }, [envName])

  const handleStatus = useCallback((data: EnvironmentStatusEvent) => {
    setStatus(data)

    // Call onStateChange if state changed
    if (onStateChangeRef.current && data.state !== previousStateRef.current) {
      previousStateRef.current = data.state
      onStateChangeRef.current(data.state)
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

  const { isConnected, error, reconnect } = useWebSocket<EnvironmentStatusEvent>(url, {
    enabled,
    eventHandlers,
  })

  return {
    status,
    state: status?.state,
    message: status?.message,
    activated: status?.activated,
    forkingStatus: status?.forkingStatus,
    isConnected,
    error,
    reconnect,
    isActive: status?.state === 'active',
    isActivating: status?.state === 'activating',
    isInactive: status?.state === 'inactive',
    isDeactivating: status?.state === 'deactivating',
    isForking:
      status?.forkingStatus?.phase !== undefined &&
      status?.forkingStatus?.phase !== 'Completed',
    isError: status?.state === 'error',
  }
}
