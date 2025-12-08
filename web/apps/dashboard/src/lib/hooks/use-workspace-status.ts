'use client'

import { useEffect, useState, useCallback, useRef } from 'react'
import { getWorkspace } from '@/app/actions/workspace.actions'
import type { Workspace } from '@kloudlite/types'

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
        console.log('[useWorkspaceStatus] Fetched workspace, phase:', currentPhase)
        setPhase(currentPhase)
        setError(null)

        // Stop polling if we've reached a terminal phase
        if (stopOnPhaseRef.current.includes(currentPhase)) {
          console.log('[useWorkspaceStatus] Terminal phase reached:', currentPhase)
          // Call onReady callback when workspace reaches ready state
          if (currentPhase === 'Running' && onReadyRef.current) {
            console.log('[useWorkspaceStatus] Calling onReady callback')
            onReadyRef.current(result.data)
          }
          return true // Signal to stop polling
        }
      } else {
        setError(result.error || 'Failed to fetch workspace status')
      }
    } catch (err) {
      console.error('Failed to fetch workspace status:', err)
      setError('Failed to fetch workspace status')
    }
    return false
  }, [workspaceName, namespace])

  const startPolling = useCallback(() => {
    console.log('[useWorkspaceStatus] Starting polling for workspace:', workspaceName)
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
    }

    setIsPolling(true)

    // Immediate first fetch
    fetchStatus().then((shouldStop) => {
      if (shouldStop) {
        console.log('[useWorkspaceStatus] First fetch returned terminal state, stopping')
        setIsPolling(false)
        return
      }

      console.log('[useWorkspaceStatus] Setting up polling interval, pollInterval:', pollInterval)
      // Start interval
      intervalRef.current = setInterval(async () => {
        console.log('[useWorkspaceStatus] Interval tick - fetching status...')
        try {
          const shouldStop = await fetchStatus()
          console.log('[useWorkspaceStatus] Interval fetch complete, shouldStop:', shouldStop)
          if (shouldStop) {
            console.log('[useWorkspaceStatus] Polling complete, stopping interval')
            if (intervalRef.current) {
              clearInterval(intervalRef.current)
              intervalRef.current = null
            }
            setIsPolling(false)
          }
        } catch (err) {
          console.error('[useWorkspaceStatus] Interval fetch error:', err)
        }
      }, pollInterval)
    })
  }, [fetchStatus, pollInterval, workspaceName])

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
        console.log('[useWorkspaceStatus] Cleanup: clearing interval')
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
