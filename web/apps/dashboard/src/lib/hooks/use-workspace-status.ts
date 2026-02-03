'use client'

import { useEffect, useState, useCallback, useRef } from 'react'
import { getWorkspace } from '@/app/actions/workspace.actions'
import type { Workspace } from '@kloudlite/lib/k8s'

interface UseWorkspaceStatusOptions {
  pollInterval?: number
  enabled?: boolean
  stopOnPhase?: string[]
  onReady?: (workspace: Workspace) => void
}

interface UseWorkspaceStatusResult {
  workspace: Workspace | null
  phase: string | null
  isPolling: boolean
  error: string | null
  startPolling: () => void
  stopPolling: () => void
}

const DEFAULT_STOP_PHASES = ['Running', 'Failed', 'Stopped']

export function useWorkspaceStatus(
  workspaceName: string,
  namespace: string,
  options: UseWorkspaceStatusOptions = {}
): UseWorkspaceStatusResult {
  const { pollInterval = 2000, enabled = false, stopOnPhase = DEFAULT_STOP_PHASES, onReady } = options

  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [phase, setPhase] = useState<string | null>(null)
  const [isPolling, setIsPolling] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<NodeJS.Timeout | null>(null)
  const onReadyRef = useRef(onReady)
  const stopOnPhaseRef = useRef(stopOnPhase)
  const onReadyCalledRef = useRef(false)

  // Keep refs updated
  useEffect(() => {
    onReadyRef.current = onReady
  }, [onReady])

  useEffect(() => {
    stopOnPhaseRef.current = stopOnPhase
  }, [stopOnPhase])

  const fetchStatus = useCallback(async () => {
    try {
      const result = await getWorkspace(workspaceName, namespace)
      if (result.success && result.data) {
        setWorkspace(result.data)
        const currentPhase = result.data.status?.phase || 'Pending'
        setPhase(currentPhase)
        setError(null)

        // Stop polling if we've reached a terminal phase
        if (stopOnPhaseRef.current.includes(currentPhase)) {
          // Call onReady callback when workspace reaches ready state (only once)
          if (currentPhase === 'Running' && onReadyRef.current && !onReadyCalledRef.current) {
            onReadyCalledRef.current = true
            onReadyRef.current(result.data)
          }
          return true // Signal to stop polling
        }
      } else {
        setError(result.error || 'Failed to fetch workspace status')
      }
    } catch (_err) {
      setError('Failed to fetch workspace status')
    }
    return false
  }, [workspaceName, namespace])

  const startPolling = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
    }

    // Reset the onReady called flag for new polling session
    onReadyCalledRef.current = false
    setIsPolling(true)

    // Immediate first fetch
    fetchStatus().then((shouldStop) => {
      if (shouldStop) {
        setIsPolling(false)
        return
      }

      // Start interval
      intervalRef.current = setInterval(async () => {
        const shouldStop = await fetchStatus()
        if (shouldStop) {
          if (intervalRef.current) {
            clearInterval(intervalRef.current)
            intervalRef.current = null
          }
          setIsPolling(false)
        }
      }, pollInterval)
    })
  }, [fetchStatus, pollInterval])

  const stopPolling = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
    setIsPolling(false)
  }, [])

  // Auto-start polling if enabled
  useEffect(() => {
    if (enabled) {
      startPolling()
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [enabled, startPolling])

  return {
    workspace,
    phase,
    isPolling,
    error,
    startPolling,
    stopPolling,
  }
}
