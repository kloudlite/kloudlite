'use client'

import { useState, useRef, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { Pause, Play, Loader2 } from 'lucide-react'
import { suspendWorkspace, activateWorkspace } from '@/app/actions/workspace.actions'
import type { Workspace } from '@kloudlite/types'

interface WorkspaceActionsProps {
  workspace: Workspace
  workMachineRunning?: boolean
}

export function WorkspaceActions({ workspace, workMachineRunning = false }: WorkspaceActionsProps) {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const pollTimerRef = useRef<NodeJS.Timeout | null>(null)

  // Poll for phase transition after an action (activate/suspend)
  // K8s may not process the change immediately, so we refresh periodically
  const startPostActionPolling = useCallback(() => {
    // Clear any existing poll timer
    if (pollTimerRef.current) {
      clearInterval(pollTimerRef.current)
    }

    // Refresh immediately
    router.refresh()

    // Then refresh every 2 seconds for 15 seconds
    let elapsed = 0
    pollTimerRef.current = setInterval(() => {
      elapsed += 2000
      router.refresh()
      if (elapsed >= 15000) {
        if (pollTimerRef.current) {
          clearInterval(pollTimerRef.current)
          pollTimerRef.current = null
        }
      }
    }, 2000)
  }, [router])

  const handleSuspend = async () => {
    setIsLoading(true)
    setError(null)

    const result = await suspendWorkspace(workspace.metadata.name, workspace.metadata.namespace)

    if (result.success) {
      startPostActionPolling()
    } else {
      setError(result.error || 'Failed to suspend workspace')
    }

    setIsLoading(false)
  }

  const handleActivate = async () => {
    setIsLoading(true)
    setError(null)

    const result = await activateWorkspace(workspace.metadata.name, workspace.metadata.namespace)

    if (result.success) {
      startPostActionPolling()
    } else {
      setError(result.error || 'Failed to activate workspace')
    }

    setIsLoading(false)
  }

  // Workspace is in a transitional state (activating)
  const isTransitioning = workspace.status?.phase === 'Creating' || workspace.status?.phase === 'Pending'
  const isActive = workspace.spec.status === 'active'
  const isSuspended = workspace.spec.status === 'suspended'

  return (
    <div className="flex items-center gap-2">
      {/* Show Suspend button when workspace is running */}
      {isActive && !isTransitioning && (
        <Button variant="outline" size="sm" onClick={handleSuspend} disabled={isLoading}>
          {isLoading ? (
            <Loader2 className="mr-1.5 h-4 w-4 animate-spin" />
          ) : (
            <Pause className="mr-1.5 h-4 w-4" />
          )}
          {isLoading ? 'Suspending...' : 'Suspend'}
        </Button>
      )}

      {/* Show Activate button when workspace is suspended */}
      {isSuspended && (
        <Button
          variant="outline"
          size="sm"
          onClick={handleActivate}
          disabled={isLoading || !workMachineRunning}
          title={!workMachineRunning ? 'Start your WorkMachine first' : undefined}
        >
          {isLoading ? (
            <Loader2 className="mr-1.5 h-4 w-4 animate-spin" />
          ) : (
            <Play className="mr-1.5 h-4 w-4" />
          )}
          {isLoading ? 'Activating...' : !workMachineRunning ? 'Activate (VM stopped)' : 'Activate'}
        </Button>
      )}

      {/* No button when transitioning - status badge shows the state */}

      {error && <span className="text-destructive text-xs">{error}</span>}
    </div>
  )
}
