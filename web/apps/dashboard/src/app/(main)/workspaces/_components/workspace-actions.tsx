'use client'

import { useState, useCallback, startTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { Pause, Play, Loader2 } from 'lucide-react'
import { suspendWorkspace, activateWorkspace } from '@/app/actions/workspace.actions'
import { useWorkspaceStatus } from '@/lib/hooks/use-workspace-status'
import type { Workspace } from '@kloudlite/types'

interface WorkspaceActionsProps {
  workspace: Workspace
  workMachineRunning?: boolean
}

export function WorkspaceActions({ workspace, workMachineRunning = false }: WorkspaceActionsProps) {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Callback when workspace becomes ready
  const handleWorkspaceReady = useCallback(() => {
    startTransition(() => {
      router.refresh()
    })
  }, [router])

  const { phase, isPolling, startPolling } = useWorkspaceStatus(
    workspace.metadata.name,
    workspace.metadata.namespace,
    {
      stopOnPhase: ['Running', 'Failed', 'Stopped'],
      onReady: handleWorkspaceReady
    }
  )

  const handleSuspend = async () => {
    setIsLoading(true)
    setError(null)

    const result = await suspendWorkspace(workspace.metadata.name, workspace.metadata.namespace)

    if (result.success) {
      router.refresh()
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
      // Start polling for status updates
      startPolling()
    } else {
      setError(result.error || 'Failed to activate workspace')
    }

    setIsLoading(false)
  }

  // Show polling state
  const isActivating = isPolling || (workspace.status?.phase === 'Creating' || workspace.status?.phase === 'Pending')
  const showActivatingState = isActivating && workspace.spec.status === 'active'

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2">
        {workspace.spec.status === 'active' && !showActivatingState && (
          <Button variant="outline" size="sm" onClick={handleSuspend} disabled={isLoading}>
            <Pause className="mr-1 h-4 w-4" />
            {isLoading ? 'Suspending...' : 'Suspend'}
          </Button>
        )}
        {showActivatingState && (
          <Button variant="outline" size="sm" disabled>
            <Loader2 className="mr-1 h-4 w-4 animate-spin" />
            Activating...
          </Button>
        )}
        {workspace.spec.status === 'suspended' && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleActivate}
            disabled={isLoading || !workMachineRunning}
            title={!workMachineRunning ? 'WorkMachine is stopped' : undefined}
          >
            <Play className="mr-1 h-4 w-4" />
            {isLoading ? 'Activating...' : !workMachineRunning ? 'Activate (VM stopped)' : 'Activate'}
          </Button>
        )}
      </div>
      {phase && isPolling && (
        <div className="text-muted-foreground flex items-center gap-1.5 text-xs">
          <Loader2 className="h-3 w-3 animate-spin" />
          <span>Status: {phase}</span>
        </div>
      )}
      {error && <div className="text-destructive text-xs">{error}</div>}
    </div>
  )
}
