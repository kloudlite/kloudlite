'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { AlertTriangle, Trash2 } from 'lucide-react'
import {
  Button,
  Input,
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@kloudlite/ui'
import { deleteWorkspace } from '@/app/actions/workspace.actions'

interface WorkspaceDangerZoneProps {
  workspaceName: string
  namespace: string
  hash: string
}

export function WorkspaceDangerZone({
  workspaceName,
  namespace,
}: WorkspaceDangerZoneProps) {
  const router = useRouter()
  const [confirmName, setConfirmName] = useState('')
  const [isDeleting, setIsDeleting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const canDelete = confirmName === workspaceName

  const handleDelete = async () => {
    if (!canDelete) return

    setIsDeleting(true)
    setError(null)

    try {
      const result = await deleteWorkspace(workspaceName, namespace)
      if (result.success) {
        router.push('/workspaces')
      } else {
        setError(result.error || 'Failed to delete workspace')
      }
    } catch {
      setError('An unexpected error occurred')
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <div className="bg-card rounded-lg border border-destructive/50">
      <div className="border-b border-destructive/50 p-4">
        <div className="flex items-center gap-2">
          <AlertTriangle className="h-4 w-4 text-destructive" />
          <h3 className="text-sm font-semibold text-destructive">Danger Zone</h3>
        </div>
      </div>
      <div className="p-4">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium">Delete this workspace</p>
            <p className="text-xs text-muted-foreground">
              Once deleted, this workspace and all its data will be permanently removed.
            </p>
          </div>

          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive" size="sm">
                <Trash2 className="h-4 w-4 mr-2" />
                Delete Workspace
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete workspace?</AlertDialogTitle>
                <AlertDialogDescription className="space-y-3">
                  <p>
                    This action cannot be undone. This will permanently delete the workspace{' '}
                    <span className="font-mono font-semibold">{workspaceName}</span> and all of its
                    data.
                  </p>
                  <div className="space-y-2">
                    <p className="text-sm">
                      Please type <span className="font-mono font-semibold">{workspaceName}</span>{' '}
                      to confirm.
                    </p>
                    <Input
                      value={confirmName}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setConfirmName(e.target.value)}
                      placeholder="Type workspace name to confirm"
                      className="font-mono"
                    />
                  </div>
                  {error && <p className="text-sm text-destructive">{error}</p>}
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel onClick={() => setConfirmName('')}>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  onClick={handleDelete}
                  disabled={!canDelete || isDeleting}
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                >
                  {isDeleting ? 'Deleting...' : 'Delete Workspace'}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </div>
    </div>
  )
}
