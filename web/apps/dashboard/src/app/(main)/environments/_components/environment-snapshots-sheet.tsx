'use client'

import { useState, useEffect, useCallback, useMemo, useId } from 'react'
import { useRouter } from 'next/navigation'
import { useResourceWatchContext } from '@/components/resource-watch-provider'
import {
  Camera,
  Plus,
  Loader2,
  Search,
  RefreshCw,
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
  restoreEnvironmentFromSnapshot,
  deleteSnapshot,
  pushSnapshot,
  getEnvironmentSnapshotStatus,
} from '@/app/actions/snapshot.actions'
import type { SnapshotOperationStatus } from '@/lib/services/snapshot.service'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import { getEnvironment } from '@/app/actions/environment.actions'
import { toast } from 'sonner'
import { SnapshotTimeGroup } from '../../_components/snapshot-time-group'
import { groupSnapshotsByTime } from '@/lib/utils/time-grouping'

interface EnvironmentSnapshotsSheetProps {
  environmentName: string
  trigger?: React.ReactNode
}

export function EnvironmentSnapshotsSheet({ environmentName, trigger }: EnvironmentSnapshotsSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [snapshots, setSnapshots] = useState<Snapshot[]>([])
  const [currentSnapshotName, setCurrentSnapshotName] = useState<string | undefined>()
  const [isCreating, setIsCreating] = useState(false)
  const [description, setDescription] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [operationStatus, setOperationStatus] = useState<SnapshotOperationStatus | null>(null)

  // Confirmation dialogs
  const [restoreDialogOpen, setRestoreDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [pushDialogOpen, setPushDialogOpen] = useState(false)
  const [selectedSnapshot, setSelectedSnapshot] = useState<Snapshot | null>(null)
  const [isRestoring, setIsRestoring] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [isPushing, setIsPushing] = useState(false)
  const [pushTag, setPushTag] = useState('')

  const loadSnapshots = useCallback(async () => {
    // First get environment to extract namespace
    const envResult = await getEnvironment(environmentName)

    if (!envResult.success || !envResult.data?.metadata?.namespace) {
      console.error('Failed to get environment or namespace')
      return
    }

    const namespace = envResult.data.metadata.namespace

    // Fetch snapshots and operation status in parallel
    const [snapshotResult, statusResult] = await Promise.all([
      listEnvironmentSnapshots(environmentName, namespace),
      getEnvironmentSnapshotStatus(environmentName),
    ])

    if (snapshotResult.success && snapshotResult.data) {
      setSnapshots(snapshotResult.data as unknown as Snapshot[])
    }

    if (envResult.success && envResult.data?.status?.lastRestoredSnapshot) {
      setCurrentSnapshotName(envResult.data.status.lastRestoredSnapshot.name)
    } else {
      setCurrentSnapshotName(undefined)
    }

    if (statusResult.success && statusResult.data) {
      setOperationStatus(statusResult.data)
    }
  }, [environmentName])

  // Load snapshots when sheet opens
  useEffect(() => {
    if (!open) return

    const frame = requestAnimationFrame(() => {
      loadSnapshots()
    })
    return () => cancelAnimationFrame(frame)
  }, [open, loadSnapshots])

  // Re-fetch snapshots when SSE reports a snapshot change (while sheet is open)
  const watchCtx = useResourceWatchContext()
  const watchId = useId()

  useEffect(() => {
    if (!open || !watchCtx) return

    watchCtx.subscribe(watchId, { plural: 'snapshots' }, () => {
      loadSnapshots()
    })

    return () => {
      watchCtx.unsubscribe(watchId)
    }
  }, [open, watchCtx, watchId, loadSnapshots])

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

    const result = await restoreEnvironmentFromSnapshot(
      environmentName,
      selectedSnapshot.name
      // sourceNamespace is optional - defaults to environment's target namespace
    )

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

  const handlePushClick = (snapshot: Snapshot) => {
    setSelectedSnapshot(snapshot)
    // Suggest a default tag based on snapshot name or date
    const shortHash = snapshot.name.split('-').slice(-1)[0]
    setPushTag(`v${shortHash}`)
    setPushDialogOpen(true)
  }

  const handlePushConfirm = async () => {
    if (!selectedSnapshot || !pushTag.trim()) return

    setIsPushing(true)

    const result = await pushSnapshot(selectedSnapshot.name, pushTag.trim())

    if (result.success) {
      toast.success('Pushing snapshot to registry')
      setPushDialogOpen(false)
      setPushTag('')
      loadSnapshots()
    } else {
      toast.error(result.error || 'Failed to push snapshot')
    }

    setIsPushing(false)
  }

  const handleDeleteConfirm = async () => {
    if (!selectedSnapshot || !selectedSnapshot.namespace) return

    setIsDeleting(true)

    const result = await deleteSnapshot(selectedSnapshot.name, selectedSnapshot.namespace)

    if (result.success) {
      toast.success('Snapshot deleted')
      setDeleteDialogOpen(false)
      loadSnapshots()
    } else {
      toast.error(result.error || 'Failed to delete snapshot')
    }

    setIsDeleting(false)
  }

  // Group snapshots by time with search filtering
  const groupedSnapshots = useMemo(() => {
    const filtered = snapshots.filter(
      (s) =>
        s.description?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        s.name.toLowerCase().includes(searchQuery.toLowerCase())
    )
    return groupSnapshotsByTime(filtered)
  }, [snapshots, searchQuery])

  return (
    <>
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger asChild suppressHydrationWarning>
          {trigger || (
            <Button variant="outline" size="sm">
              <Camera className="mr-1 h-4 w-4" />
              Snapshots
            </Button>
          )}
        </SheetTrigger>
        <SheetContent className="flex w-full flex-col p-0 sm:max-w-2xl">
          <SheetHeader className="border-b px-6 py-5">
            <SheetTitle className="text-lg">Environment Snapshots</SheetTitle>
            <SheetDescription>
              Save and restore your environment state at any point in time
            </SheetDescription>
          </SheetHeader>

          {/* Operation Status Banner */}
          {operationStatus?.inProgress && (
            <div className="border-b bg-blue-50 dark:bg-blue-950/30 px-6 py-4">
              <div className="flex items-center gap-3">
                <Loader2 className="h-4 w-4 animate-spin text-blue-600 dark:text-blue-400" />
                <div className="flex-1">
                  <p className="text-sm font-medium text-foreground">
                    {operationStatus.operation === 'creating' ? 'Creating snapshot...' : 'Restoring snapshot...'}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {operationStatus.phase && `${operationStatus.phase}`}
                    {operationStatus.message && ` - ${operationStatus.message}`}
                  </p>
                </div>
                <Button variant="ghost" size="sm" onClick={loadSnapshots}>
                  <RefreshCw className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}

          {/* Search + Create Section */}
          <div className="border-b bg-muted/30 px-6 py-4 space-y-3">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search snapshots..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* Create */}
            <div className="flex gap-2">
              <Input
                placeholder="Describe this snapshot (optional)"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={isCreating || operationStatus?.inProgress}
                className="flex-1"
              />
              <Button
                onClick={handleCreate}
                disabled={isCreating || operationStatus?.inProgress}
                size="sm"
              >
                {isCreating ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <>
                    <Plus className="mr-2 h-4 w-4" />
                    Create
                  </>
                )}
              </Button>
            </div>
          </div>

          {/* Snapshots by Time Group */}
          <ScrollArea className="flex-1">
            <div className="px-6 py-4 space-y-6">
              {Object.entries(groupedSnapshots).map(
                ([label, snaps]) =>
                  snaps.length > 0 && (
                    <SnapshotTimeGroup
                      key={label}
                      label={label}
                      snapshots={snaps}
                      currentSnapshotName={currentSnapshotName}
                      onRestore={handleRestoreClick}
                      onDelete={handleDeleteClick}
                      onPush={handlePushClick}
                      disabled={operationStatus?.inProgress}
                    />
                  )
              )}

              {snapshots.length === 0 && (
                <div className="text-muted-foreground py-12 text-center">
                  <Camera className="mx-auto mb-3 h-10 w-10 opacity-40" />
                  <p className="text-sm font-medium">No snapshots yet</p>
                  <p className="text-xs mt-1">
                    Create your first snapshot to save environment state
                  </p>
                </div>
              )}

              {snapshots.length > 0 &&
                Object.values(groupedSnapshots).every((g) => g.length === 0) && (
                  <div className="text-muted-foreground py-12 text-center">
                    <Search className="mx-auto mb-3 h-10 w-10 opacity-40" />
                    <p className="text-sm font-medium">No snapshots found</p>
                    <p className="text-xs mt-1">
                      No snapshots match &quot;{searchQuery}&quot;
                    </p>
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

      {/* Push Dialog with Tag Input */}
      <Dialog open={pushDialogOpen} onOpenChange={setPushDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Push Snapshot</DialogTitle>
            <DialogDescription>
              Push this snapshot to the registry. Provide a tag to identify this version.
              Once pushed, the snapshot cannot be pushed again (snapshots are immutable).
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="push-tag">Tag</Label>
              <Input
                id="push-tag"
                placeholder="e.g., v1.0, latest, stable"
                value={pushTag}
                onChange={(e) => setPushTag(e.target.value)}
                disabled={isPushing}
              />
              <p className="text-xs text-muted-foreground">
                Tags help you identify and pull specific versions of your snapshot.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPushDialogOpen(false)} disabled={isPushing}>
              Cancel
            </Button>
            <Button onClick={handlePushConfirm} disabled={isPushing || !pushTag.trim()}>
              {isPushing ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Pushing...
                </>
              ) : (
                'Push'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
