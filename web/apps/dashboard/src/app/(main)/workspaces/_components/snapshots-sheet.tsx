'use client'

import { useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import {
  Camera,
  Plus,
  Loader2,
  RotateCcw,
  Trash2,
  Clock,
  HardDrive,
  AlertCircle,
} from 'lucide-react'
import { Button } from '@kloudlite/ui'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import { ScrollArea } from '@kloudlite/ui'
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
import type { Snapshot } from '@/lib/services/snapshot.service'
import {
  listSnapshots,
  createSnapshot,
  restoreSnapshot,
  deleteSnapshot,
} from '@/app/actions/snapshot.actions'
import { toast } from 'sonner'

interface SnapshotsSheetProps {
  workspace: Workspace
  trigger?: React.ReactNode
  workMachineRunning?: boolean
}

function formatTimeAgo(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000)

  if (diffInSeconds < 60) return 'just now'
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} min ago`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} hours ago`
  if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)} days ago`
  return date.toLocaleDateString()
}

function getStateBadge(state: Snapshot['status']['state']) {
  const baseClasses = 'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium'

  switch (state) {
    case 'Ready':
      return (
        <span className={`${baseClasses} bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400`}>
          Ready
        </span>
      )
    case 'Creating':
      return (
        <span className={`${baseClasses} bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400`}>
          <Loader2 className="h-3 w-3 animate-spin" />
          Creating
        </span>
      )
    case 'Restoring':
      return (
        <span className={`${baseClasses} bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400`}>
          <Loader2 className="h-3 w-3 animate-spin" />
          Restoring
        </span>
      )
    case 'Deleting':
      return (
        <span className={`${baseClasses} bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400`}>
          <Loader2 className="h-3 w-3 animate-spin" />
          Deleting
        </span>
      )
    case 'Failed':
      return (
        <span className={`${baseClasses} bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400`}>
          <AlertCircle className="h-3 w-3" />
          Failed
        </span>
      )
    case 'Pending':
    default:
      return (
        <span className={`${baseClasses} bg-secondary text-secondary-foreground`}>
          Pending
        </span>
      )
  }
}

export function SnapshotsSheet({ workspace, trigger, workMachineRunning = false }: SnapshotsSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [snapshots, setSnapshots] = useState<Snapshot[]>([])
  const [isCreating, setIsCreating] = useState(false)
  const [description, setDescription] = useState('')

  // Confirmation dialogs
  const [restoreDialogOpen, setRestoreDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [selectedSnapshot, setSelectedSnapshot] = useState<Snapshot | null>(null)
  const [isRestoring, setIsRestoring] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const loadSnapshots = useCallback(async () => {
    const result = await listSnapshots(workspace.metadata.name, workspace.metadata.namespace)
    if (result.success && result.data) {
      setSnapshots(result.data.snapshots || [])
    }
  }, [workspace.metadata.name, workspace.metadata.namespace])

  // Load snapshots when sheet opens
  useEffect(() => {
    if (open) {
      loadSnapshots()
    }
  }, [open, loadSnapshots])

  // Auto-refresh when there are in-progress snapshots
  useEffect(() => {
    if (!open) return undefined

    const hasInProgress = snapshots.some(
      (s) => s.status.state === 'Creating' || s.status.state === 'Restoring' || s.status.state === 'Deleting'
    )

    if (hasInProgress) {
      const interval = setInterval(loadSnapshots, 3000)
      return () => clearInterval(interval)
    }
    return undefined
  }, [open, snapshots, loadSnapshots])

  const handleCreate = async () => {
    setIsCreating(true)

    const result = await createSnapshot(
      workspace.metadata.name,
      workspace.metadata.namespace,
      description ? { description, includeMetadata: true } : { includeMetadata: true }
    )

    if (result.success) {
      toast.success('Snapshot creation started')
      setDescription('')
      loadSnapshots()
    } else {
      toast.error(result.error || 'Failed to create snapshot')
    }

    setIsCreating(false)
  }

  const handleRestoreClick = (snapshot: Snapshot) => {
    setSelectedSnapshot(snapshot)
    setRestoreDialogOpen(true)
  }

  const handleRestoreConfirm = async () => {
    if (!selectedSnapshot) return

    setIsRestoring(true)

    const result = await restoreSnapshot(selectedSnapshot.metadata.name)

    if (result.success) {
      toast.success('Snapshot restore started')
      setRestoreDialogOpen(false)
      loadSnapshots()
      router.refresh()
    } else {
      toast.error(result.error || 'Failed to restore snapshot')
    }

    setIsRestoring(false)
  }

  const handleDeleteClick = (snapshot: Snapshot) => {
    setSelectedSnapshot(snapshot)
    setDeleteDialogOpen(true)
  }

  const handleDeleteConfirm = async () => {
    if (!selectedSnapshot) return

    setIsDeleting(true)

    const result = await deleteSnapshot(selectedSnapshot.metadata.name)

    if (result.success) {
      toast.success('Snapshot deleted')
      setDeleteDialogOpen(false)
      loadSnapshots()
    } else {
      toast.error(result.error || 'Failed to delete snapshot')
    }

    setIsDeleting(false)
  }

  return (
    <>
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger asChild>
          {trigger || (
            <Button variant="outline" size="sm">
              <Camera className="mr-1 h-4 w-4" />
              Snapshots
            </Button>
          )}
        </SheetTrigger>
        <SheetContent className="flex w-full flex-col p-6 sm:max-w-2xl">
          <SheetHeader className="mb-6">
            <SheetTitle>Workspace Snapshots</SheetTitle>
            <SheetDescription>
              Save and restore your workspace state
            </SheetDescription>
          </SheetHeader>

          <ScrollArea className="flex-1">
            <div className="space-y-6 pr-4">
              {/* Create Snapshot Form */}
              <div className="bg-muted/50 space-y-3 rounded-lg border p-4">
                <h4 className="text-sm font-medium">Create New Snapshot</h4>
                {!workMachineRunning && (
                  <div className="flex items-center gap-2 rounded-md bg-yellow-50 p-3 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-400">
                    <AlertCircle className="h-4 w-4 flex-shrink-0" />
                    <p className="text-xs">WorkMachine must be running to create snapshots</p>
                  </div>
                )}
                <div className="space-y-2">
                  <Label htmlFor="description">Description (optional)</Label>
                  <Input
                    id="description"
                    placeholder="e.g., Before refactoring auth module"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    disabled={isCreating || !workMachineRunning}
                  />
                </div>
                <Button
                  onClick={handleCreate}
                  disabled={isCreating || !workMachineRunning}
                  className="w-full"
                >
                  {isCreating ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    <>
                      <Plus className="mr-2 h-4 w-4" />
                      Create Snapshot
                    </>
                  )}
                </Button>
              </div>

              {/* Snapshots List */}
              {snapshots.length > 0 && (
                <div className="space-y-2">
                  <h4 className="text-sm font-medium">Snapshots ({snapshots.length})</h4>
                  <div className="space-y-2">
                    {snapshots.map((snapshot) => (
                      <div
                        key={snapshot.metadata.name}
                        className="bg-card rounded-lg border p-4"
                      >
                        <div className="flex items-start justify-between gap-4">
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-2 flex-wrap">
                              <span className="truncate font-mono text-sm">
                                {snapshot.metadata.name}
                              </span>
                              {getStateBadge(snapshot.status.state)}
                            </div>
                            <div className="text-muted-foreground mt-2 flex items-center gap-3 text-xs">
                              {snapshot.status.sizeHuman && (
                                <span className="flex items-center gap-1">
                                  <HardDrive className="h-3 w-3" />
                                  {snapshot.status.sizeHuman}
                                </span>
                              )}
                              <span className="flex items-center gap-1">
                                <Clock className="h-3 w-3" />
                                {formatTimeAgo(snapshot.status.createdAt || snapshot.metadata.creationTimestamp)}
                              </span>
                            </div>
                            {snapshot.spec.description && (
                              <p className="text-muted-foreground mt-2 text-sm italic">
                                &quot;{snapshot.spec.description}&quot;
                              </p>
                            )}
                            {snapshot.status.state === 'Failed' && snapshot.status.message && (
                              <p className="mt-2 text-xs text-red-600 dark:text-red-400">
                                {snapshot.status.message}
                              </p>
                            )}
                          </div>
                          <div className="flex items-center gap-2 flex-shrink-0">
                            {snapshot.status.state === 'Ready' && (
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleRestoreClick(snapshot)}
                                disabled={!workMachineRunning}
                                title={!workMachineRunning ? 'WorkMachine must be running to restore' : undefined}
                              >
                                <RotateCcw className="mr-1 h-3 w-3" />
                                Restore
                              </Button>
                            )}
                            {(snapshot.status.state === 'Ready' || snapshot.status.state === 'Failed') && (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDeleteClick(snapshot)}
                                className="text-destructive hover:text-destructive"
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {snapshots.length === 0 && (
                <div className="text-muted-foreground py-8 text-center">
                  <Camera className="mx-auto mb-3 h-12 w-12 opacity-50" />
                  <p className="text-sm">No snapshots yet</p>
                  <p className="text-xs">Create your first snapshot to save workspace state</p>
                </div>
              )}
            </div>
          </ScrollArea>
        </SheetContent>
      </Sheet>

      {/* Restore Confirmation Dialog */}
      <AlertDialog open={restoreDialogOpen} onOpenChange={setRestoreDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Restore Snapshot</AlertDialogTitle>
            <AlertDialogDescription>
              This will restore your workspace to the state when this snapshot was taken.
              Your workspace will be suspended during the restore process.
              Any unsaved changes since the snapshot will be lost.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRestoring}>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleRestoreConfirm} disabled={isRestoring}>
              {isRestoring ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Restoring...
                </>
              ) : (
                'Restore'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Snapshot</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this snapshot? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConfirm}
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
