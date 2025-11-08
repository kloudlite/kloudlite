'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { MoreHorizontal, Loader2, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
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
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)

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
        throw new Error(result.error)
      }
      router.refresh()
    } catch (error) {
      console.error(`Failed to ${action} workspace:`, error)
      alert(error instanceof Error ? error.message : `Failed to ${action} workspace`)
    }
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
            className="text-destructive focus:text-destructive"
            onClick={() => setShowDeleteDialog(true)}
          >
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

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
