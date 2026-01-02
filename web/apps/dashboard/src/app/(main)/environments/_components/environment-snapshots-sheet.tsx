'use client'

import { useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import {
  Camera,
  Plus,
  Loader2,
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
import type { Snapshot } from '@/lib/services/snapshot.service'
import {
  listEnvironmentSnapshots,
  createEnvironmentSnapshot,
  restoreSnapshot,
  deleteSnapshot,
} from '@/app/actions/snapshot.actions'
import { getEnvironment } from '@/app/actions/environment.actions'
import { toast } from 'sonner'
import { SnapshotTimeline } from '@/app/(main)/workspaces/_components/snapshot-timeline'

interface EnvironmentSnapshotsSheetProps {
  environmentName: string
  trigger?: React.ReactNode
  isActive?: boolean
}

export function EnvironmentSnapshotsSheet({ environmentName, trigger, isActive = false }: EnvironmentSnapshotsSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [snapshots, setSnapshots] = useState<Snapshot[]>([])
  const [currentSnapshotName, setCurrentSnapshotName] = useState<string | undefined>()
  const [isCreating, setIsCreating] = useState(false)
  const [description, setDescription] = useState('')

  // Confirmation dialogs
  const [restoreDialogOpen, setRestoreDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [selectedSnapshot, setSelectedSnapshot] = useState<Snapshot | null>(null)
  const [isRestoring, setIsRestoring] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const loadSnapshots = useCallback(async () => {
    // Fetch snapshots and environment in parallel
    const [snapshotResult, envResult] = await Promise.all([
      listEnvironmentSnapshots(environmentName),
      getEnvironment(environmentName),
    ])

    if (snapshotResult.success && snapshotResult.data) {
      setSnapshots(snapshotResult.data.snapshots || [])
    }

    if (envResult.success && envResult.data?.status?.lastRestoredSnapshot) {
      setCurrentSnapshotName(envResult.data.status.lastRestoredSnapshot.name)
    } else {
      setCurrentSnapshotName(undefined)
    }
  }, [environmentName])

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

    const result = await createEnvironmentSnapshot(
      environmentName,
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
        <SheetContent className="flex w-full flex-col p-0 sm:max-w-xl">
          <SheetHeader className="border-b px-6 py-4">
            <SheetTitle>Environment Snapshots</SheetTitle>
            <SheetDescription>
              Save and restore your environment state
            </SheetDescription>
          </SheetHeader>

          {/* Sticky Create Section */}
          <div className="border-b bg-muted/30 px-6 py-4">
            <div className="space-y-3">
              {!isActive && (
                <div className="flex items-center gap-2 rounded-md bg-yellow-50 p-3 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-400">
                  <AlertCircle className="h-4 w-4 flex-shrink-0" />
                  <p className="text-xs">Environment must be active to create snapshots</p>
                </div>
              )}
              <div className="flex gap-2">
                <Input
                  placeholder="Snapshot description (optional)"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  disabled={isCreating || !isActive}
                  className="flex-1"
                />
                <Button
                  onClick={handleCreate}
                  disabled={isCreating || !isActive}
                  size="sm"
                >
                  {isCreating ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <>
                      <Plus className="mr-1 h-4 w-4" />
                      Create
                    </>
                  )}
                </Button>
              </div>
            </div>
          </div>

          {/* Scrollable History Section */}
          <ScrollArea className="flex-1">
            <div className="px-6 py-4">
              {snapshots.length > 0 ? (
                <SnapshotTimeline
                  snapshots={snapshots}
                  onRestore={handleRestoreClick}
                  onDelete={handleDeleteClick}
                  disabled={!isActive}
                  currentSnapshotName={currentSnapshotName}
                />
              ) : (
                <div className="text-muted-foreground py-12 text-center">
                  <Camera className="mx-auto mb-3 h-10 w-10 opacity-40" />
                  <p className="text-sm font-medium">No snapshots yet</p>
                  <p className="text-xs mt-1">Create your first snapshot to save environment state</p>
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
              This will restore your environment to the state when this snapshot was taken.
              All PVCs will be restored to their previous state.
              Any changes since the snapshot will be lost.
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
