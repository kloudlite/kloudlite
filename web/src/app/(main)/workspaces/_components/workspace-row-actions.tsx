'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { MoreHorizontal, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Workspace } from '@/types/workspace'
import {
  deleteWorkspace,
  suspendWorkspace,
  activateWorkspace,
  archiveWorkspace,
} from '@/app/actions/workspace.actions'

interface WorkspaceRowActionsProps {
  workspace: Workspace
}

export function WorkspaceRowActions({ workspace }: WorkspaceRowActionsProps) {
  const router = useRouter()
  const [isDeleting, setIsDeleting] = useState(false)

  const handleDelete = async () => {
    if (!confirm(`Are you sure you want to delete workspace "${workspace.metadata.name}"?`)) {
      return
    }

    setIsDeleting(true)
    try {
      const result = await deleteWorkspace(workspace.metadata.name, workspace.metadata.namespace)
      if (!result.success) {
        throw new Error(result.error)
      }
      router.refresh()
    } catch (error) {
      console.error('Failed to delete workspace:', error)
      alert(error instanceof Error ? error.message : 'Failed to delete workspace')
    } finally {
      setIsDeleting(false)
    }
  }

  const handleWorkspaceAction = async (action: 'suspend' | 'activate' | 'archive') => {
    try {
      let result
      if (action === 'suspend') {
        result = await suspendWorkspace(workspace.metadata.name, workspace.metadata.namespace)
      } else if (action === 'activate') {
        result = await activateWorkspace(workspace.metadata.name, workspace.metadata.namespace)
      } else if (action === 'archive') {
        result = await archiveWorkspace(workspace.metadata.name, workspace.metadata.namespace)
      }

      if (result && !result.success) {
        throw new Error(result.error)
      }
      router.refresh()
    } catch (error) {
      console.error(`Failed to ${action} workspace:`, error)
      alert(error instanceof Error ? error.message : `Failed to ${action} workspace`)
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="h-8 w-8 p-0"
          disabled={isDeleting}
        >
          {isDeleting ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <MoreHorizontal className="h-4 w-4" />
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem asChild>
          <Link href={`/workspaces/${workspace.metadata.namespace}/${workspace.metadata.name}`}>
            Open Workspace
          </Link>
        </DropdownMenuItem>
        {workspace.spec.status !== 'suspended' && (
          <DropdownMenuItem onClick={() => handleWorkspaceAction('suspend')}>
            Suspend
          </DropdownMenuItem>
        )}
        {workspace.spec.status === 'suspended' && (
          <DropdownMenuItem onClick={() => handleWorkspaceAction('activate')}>
            Activate
          </DropdownMenuItem>
        )}
        {workspace.spec.status !== 'archived' && (
          <DropdownMenuItem onClick={() => handleWorkspaceAction('archive')}>
            Archive
          </DropdownMenuItem>
        )}
        <DropdownMenuItem>Settings</DropdownMenuItem>
        <DropdownMenuItem
          className="text-red-600"
          onClick={handleDelete}
        >
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
