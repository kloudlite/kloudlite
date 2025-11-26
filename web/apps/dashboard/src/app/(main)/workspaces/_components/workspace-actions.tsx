'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { Pause, Play, Settings } from 'lucide-react'
import { suspendWorkspace, activateWorkspace } from '@/app/actions/workspace.actions'
import type { Workspace } from '@kloudlite/types'

interface WorkspaceActionsProps {
  workspace: Workspace
}

export function WorkspaceActions({ workspace }: WorkspaceActionsProps) {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

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
      router.refresh()
    } else {
      setError(result.error || 'Failed to activate workspace')
    }

    setIsLoading(false)
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex gap-2">
        {workspace.spec.status === 'active' && (
          <Button variant="outline" size="sm" onClick={handleSuspend} disabled={isLoading}>
            <Pause className="mr-1 h-4 w-4" />
            {isLoading ? 'Suspending...' : 'Suspend'}
          </Button>
        )}
        {workspace.spec.status === 'suspended' && (
          <Button variant="outline" size="sm" onClick={handleActivate} disabled={isLoading}>
            <Play className="mr-1 h-4 w-4" />
            {isLoading ? 'Activating...' : 'Activate'}
          </Button>
        )}
        <Button variant="outline" size="sm">
          <Settings className="mr-1 h-4 w-4" />
          Settings
        </Button>
      </div>
      {error && <div className="text-destructive text-xs">{error}</div>}
    </div>
  )
}
