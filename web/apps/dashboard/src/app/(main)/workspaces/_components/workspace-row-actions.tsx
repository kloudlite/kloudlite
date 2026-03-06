'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { MoreHorizontal, Loader2, AlertCircle, Pin, PinOff } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@kloudlite/ui'
import type { Workspace } from '@kloudlite/types'
import {
  deleteWorkspace,
  suspendWorkspace,
  activateWorkspace,
  archiveWorkspace,
} from '@/app/actions/workspace-mutation.actions'
import { pinWorkspace, unpinWorkspace } from '@/app/actions/user-preferences.actions'
import { ForkWorkspaceSheet } from './fork-workspace-sheet'
import { toast } from 'sonner'

interface WorkspaceRowActionsProps {
  workspace: Workspace
  workMachineRunning?: boolean
  isPinned?: boolean
}

export function WorkspaceRowActions({ workspace, workMachineRunning = false, isPinned = false }: WorkspaceRowActionsProps) {
  const router = useRouter()
  const [mounted, setMounted] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [showForkSheet, setShowForkSheet] = useState(false)

  // Prevent hydration mismatch with Radix UI components
  useEffect(() => {
    setMounted(true)
  }, [])

  const handlePin = async () => {
    try {
      const result = await pinWorkspace(workspace.metadata.name, workspace.metadata.namespace)
      if (result.success) {
        toast.success('Workspace pinned to dashboard')
        router.refresh()
      } else {
        toast.error('Failed to pin workspace', { description: result.error })
      }
    } catch {
      toast.error('Failed to pin workspace')
    }
  }

  const handleUnpin = async () => {
    try {
      const result = await unpinWorkspace(workspace.metadata.name, workspace.metadata.namespace)
      if (result.success) {
        toast.success('Workspace unpinned from dashboard')
        router.refresh()
      } else {
        toast.error('Failed to unpin workspace', { description: result.error })
      }
    } catch {
      toast.error('Failed to unpin workspace')
    }
  }

  const handleDelete = async () => {
    setIsDeleting(true)
    setDeleteError(null)
    try {
      const result = await deleteWorkspace(workspace.metadata.name, workspace.metadata.namespace)
      if (!result.success) {
        setDeleteError(result.error || 'Failed to delete workspace')
        return
      }
      setShowDeleteDialog(false)
      router.refresh()
    } catch (error) {
      console.error('Failed to delete workspace:', error)
      setDeleteError(error instanceof Error ? error.message : 'Failed to delete workspace')
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
        toast.error(`Failed to ${action} workspace`, { description: result.error })
        return
      }
      router.refresh()
    } catch (error) {
      console.error(`Failed to ${action} workspace:`, error)
      toast.error(`Failed to ${action} workspace`)
    }
  }

  // Show placeholder button during SSR to prevent hydration mismatch
  if (!mounted) {
    return (
      <Button variant="ghost" size="sm" className="h-8 w-8 p-0" disabled={isDeleting}>
        {isDeleting ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <MoreHorizontal className="h-4 w-4" />
        )}
      </Button>
    )
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm" className="h-8 w-8 p-0" disabled={isDeleting}>
            {isDeleting ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <MoreHorizontal className="h-4 w-4" />
            )}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem asChild>
            <Link href={`/workspaces/${workspace.status?.hash || workspace.metadata.name}`}>
              Open Workspace
            </Link>
          </DropdownMenuItem>
          {isPinned ? (
            <DropdownMenuItem onClick={handleUnpin}>
              <PinOff className="mr-2 h-4 w-4" />
              Unpin from Dashboard
            </DropdownMenuItem>
          ) : (
            <DropdownMenuItem onClick={handlePin}>
              <Pin className="mr-2 h-4 w-4" />
              Pin to Dashboard
            </DropdownMenuItem>
          )}
          <DropdownMenuItem
            onSelect={(e) => {
              e.preventDefault()
              if (workMachineRunning) {
                setShowForkSheet(true)
              }
            }}
            className={!workMachineRunning ? "text-muted-foreground cursor-not-allowed" : ""}
            disabled={!workMachineRunning}
          >
            {workMachineRunning ? 'Fork' : 'Fork (VM stopped)'}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          {workspace.spec.status !== 'suspended' && (
            <DropdownMenuItem onClick={() => handleWorkspaceAction('suspend')}>
              Suspend
            </DropdownMenuItem>
          )}
          {workspace.spec.status === 'suspended' && (
            <DropdownMenuItem
              onClick={() => workMachineRunning && handleWorkspaceAction('activate')}
              className={!workMachineRunning ? "text-muted-foreground cursor-not-allowed" : ""}
              disabled={!workMachineRunning}
            >
              {workMachineRunning ? 'Activate' : 'Activate (VM stopped)'}
            </DropdownMenuItem>
          )}
          {workspace.spec.status !== 'archived' && (
            <DropdownMenuItem onClick={() => handleWorkspaceAction('archive')}>
              Archive
            </DropdownMenuItem>
          )}
          <DropdownMenuItem>Settings</DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            className="text-destructive focus:text-destructive"
            onClick={() => setShowDeleteDialog(true)}
          >
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <ForkWorkspaceSheet
        open={showForkSheet}
        onOpenChange={setShowForkSheet}
      />

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Workspace</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete workspace <strong>{workspace.metadata.name}</strong>?
              This action cannot be undone. All data associated with this workspace will be
              permanently deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>

          {deleteError && (
            <div className="bg-destructive/10 border border-destructive/20 rounded-md p-3 flex items-start gap-2">
              <AlertCircle className="h-5 w-5 text-destructive mt-0.5 flex-shrink-0" />
              <div className="flex-1">
                <p className="text-sm font-medium text-destructive">Unable to delete workspace</p>
                <p className="text-sm text-destructive/90 mt-1">{deleteError}</p>
              </div>
            </div>
          )}

          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
