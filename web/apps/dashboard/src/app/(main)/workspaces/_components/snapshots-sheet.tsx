'use client'

import { useState, useEffect, useCallback, useMemo } from 'react'
import { useRouter } from 'next/navigation'
import {
  Camera,
  Plus,
  Loader2,
  AlertCircle,
  Search,
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
import type { Workspace } from '@kloudlite/types'
import type { Snapshot } from '@/lib/services/snapshot.service'
import {
  listSnapshots,
  createSnapshot,
  restoreSnapshot,
  deleteSnapshot,
  pushSnapshot,
} from '@/app/actions/snapshot.actions'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import { toast } from 'sonner'
import { SnapshotTimeGroup } from '../../_components/snapshot-time-group'
import { groupSnapshotsByTime } from '@/lib/utils/time-grouping'

interface SnapshotsSheetProps {
  workspace: Workspace
  trigger?: React.ReactNode
  workMachineRunning?: boolean
}

export function SnapshotsSheet({ workspace, trigger, workMachineRunning = false }: SnapshotsSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [snapshots, setSnapshots] = useState<Snapshot[]>([])
  const [isCreating, setIsCreating] = useState(false)
  const [description, setDescription] = useState('')
  const [searchQuery, setSearchQuery] = useState('')

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
      (s) =>
        s.state === 'Creating' ||
        s.state === 'Uploading' ||
        s.state === 'Restoring' ||
        s.state === 'Deleting' ||
        s.state === 'Pushing' ||
        s.state === 'Pulling'
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

    const result = await restoreSnapshot(selectedSnapshot.name)

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
        <SheetTrigger asChild>
          {trigger || (
            <Button variant="outline" size="sm">
              <Camera className="mr-1 h-4 w-4" />
              Snapshots
            </Button>
          )}
        </SheetTrigger>
        <SheetContent className="flex w-full flex-col p-0 sm:max-w-2xl">
          <SheetHeader className="border-b px-6 py-5">
            <SheetTitle className="text-lg">Workspace Snapshots</SheetTitle>
            <SheetDescription>
              Save and restore your workspace state at any point in time
            </SheetDescription>
          </SheetHeader>

          {/* Search + Create Section */}
          <div className="border-b bg-muted/30 px-6 py-4 space-y-3">
            {!workMachineRunning && (
              <div className="flex items-center gap-2 rounded-md bg-yellow-50 p-3 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-400">
                <AlertCircle className="h-4 w-4 flex-shrink-0" />
                <p className="text-xs">WorkMachine must be running to create snapshots</p>
              </div>
            )}

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
                disabled={isCreating || !workMachineRunning}
                className="flex-1"
              />
              <Button
                onClick={handleCreate}
                disabled={isCreating || !workMachineRunning}
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
                      currentSnapshotName={workspace.status?.lastRestoredSnapshot?.name}
                      onRestore={handleRestoreClick}
                      onDelete={handleDeleteClick}
                      onPush={handlePushClick}
                      disabled={!workMachineRunning}
                    />
                  )
              )}

              {snapshots.length === 0 && (
                <div className="text-muted-foreground py-12 text-center">
                  <Camera className="mx-auto mb-3 h-10 w-10 opacity-40" />
                  <p className="text-sm font-medium">No snapshots yet</p>
                  <p className="text-xs mt-1">
                    Create your first snapshot to save workspace state
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
