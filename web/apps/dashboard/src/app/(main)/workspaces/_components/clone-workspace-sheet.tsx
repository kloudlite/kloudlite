'use client'

import { useState, useTransition, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2, Camera, Check, ChevronsUpDown } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@kloudlite/ui'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@kloudlite/ui'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@kloudlite/ui'
import { listPushedSnapshots, createWorkspaceFromSnapshot } from '@/app/actions/snapshot.actions'
import { toast } from 'sonner'
import type { Snapshot } from '@/lib/services/snapshot.service'
import { cn } from '@/lib/utils'

interface CloneWorkspaceSheetProps {
  trigger?: React.ReactNode
  open?: boolean
  onOpenChange?: (open: boolean) => void
  preselectedSnapshot?: Snapshot
}

export function CloneWorkspaceSheet({
  trigger,
  open: controlledOpen,
  onOpenChange,
  preselectedSnapshot,
}: CloneWorkspaceSheetProps) {
  const router = useRouter()
  const [internalOpen, setInternalOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [isLoadingSnapshots, setIsLoadingSnapshots] = useState(false)
  const [snapshots, setSnapshots] = useState<Snapshot[]>([])
  const [popoverOpen, setPopoverOpen] = useState(false)

  // Use controlled state if provided, otherwise use internal state
  const open = controlledOpen !== undefined ? controlledOpen : internalOpen
  const setOpen = onOpenChange || setInternalOpen

  // Form fields
  const [name, setName] = useState('')
  const [selectedSnapshot, setSelectedSnapshot] = useState<Snapshot | null>(preselectedSnapshot || null)

  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null)

  // Clean up polling interval on unmount
  useEffect(() => {
    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current)
        pollIntervalRef.current = null
      }
    }
  }, [])

  // Load pushed snapshots when sheet opens
  useEffect(() => {
    if (open) {
      setIsLoadingSnapshots(true)
      listPushedSnapshots('workspace').then((result) => {
        if (result.success && result.data) {
          setSnapshots(result.data.snapshots || [])
        }
        setIsLoadingSnapshots(false)
      })

      // Set preselected snapshot if provided
      if (preselectedSnapshot) {
        setSelectedSnapshot(preselectedSnapshot)
        // Suggest a default name based on snapshot source
        const sourceName = preselectedSnapshot.status.targetName || preselectedSnapshot.metadata.name
        setName(`${sourceName}-clone`)
      } else {
        setSelectedSnapshot(null)
        setName('')
      }
    }
  }, [open, preselectedSnapshot])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      toast.error('Please enter a workspace name')
      return
    }

    if (!selectedSnapshot) {
      toast.error('Please select a snapshot')
      return
    }

    startTransition(async () => {
      // Generate resource name from display name
      const normalizedName = name
        .trim()
        .toLowerCase()
        .replace(/[^a-z0-9-]/g, '-')
        .replace(/-+/g, '-')
        .replace(/^-|-$/g, '')

      const result = await createWorkspaceFromSnapshot({
        name: normalizedName,
        displayName: name.trim(),
        snapshotName: selectedSnapshot.metadata.name,
      })

      if (result.success) {
        toast.success('Workspace creation initiated from snapshot')
        setOpen(false)
        setName('')
        setSelectedSnapshot(null)

        // Immediately refresh and then poll for a few seconds to catch state changes
        router.refresh()

        // Clear any existing interval before starting a new one
        if (pollIntervalRef.current) {
          clearInterval(pollIntervalRef.current)
        }

        // Poll every second for 10 seconds to catch the workspace state updates
        let pollCount = 0
        pollIntervalRef.current = setInterval(() => {
          router.refresh()
          pollCount++
          if (pollCount >= 10 && pollIntervalRef.current) {
            clearInterval(pollIntervalRef.current)
            pollIntervalRef.current = null
          }
        }, 1000)
      } else {
        toast.error(result.error || 'Failed to create workspace from snapshot')
      }
    })
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      {trigger && <SheetTrigger asChild>{trigger}</SheetTrigger>}
      <SheetContent side="right" className="w-full sm:max-w-lg">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Create Workspace from Snapshot</SheetTitle>
            <SheetDescription>
              Create a new workspace from a pushed snapshot. The snapshot will be pulled from the
              registry and restored.
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-6 overflow-y-auto p-4">
            {/* Snapshot Selection */}
            <div className="space-y-2">
              <Label>Select Snapshot *</Label>
              <Popover open={popoverOpen} onOpenChange={setPopoverOpen}>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    role="combobox"
                    aria-expanded={popoverOpen}
                    className="w-full justify-between font-normal"
                    disabled={isPending || isLoadingSnapshots}
                  >
                    {isLoadingSnapshots ? (
                      <span className="text-muted-foreground flex items-center gap-2">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Loading snapshots...
                      </span>
                    ) : selectedSnapshot ? (
                      <span className="flex items-center gap-2">
                        <Camera className="h-4 w-4" />
                        {selectedSnapshot.status.registryStatus?.tag || selectedSnapshot.metadata.name}
                      </span>
                    ) : (
                      <span className="text-muted-foreground">Select a snapshot...</span>
                    )}
                    <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-[400px] p-0" align="start">
                  <Command>
                    <CommandInput placeholder="Search snapshots..." />
                    <CommandList>
                      <CommandEmpty>
                        {snapshots.length === 0
                          ? 'No pushed snapshots available. Push a snapshot first to clone from it.'
                          : 'No matching snapshots found.'}
                      </CommandEmpty>
                      <CommandGroup>
                        {snapshots.map((snapshot) => (
                          <CommandItem
                            key={snapshot.metadata.name}
                            value={snapshot.metadata.name}
                            onSelect={() => {
                              setSelectedSnapshot(snapshot)
                              setPopoverOpen(false)
                              // Suggest name based on source
                              if (!name) {
                                const sourceName = snapshot.status.targetName || snapshot.metadata.name
                                setName(`${sourceName}-clone`)
                              }
                            }}
                          >
                            <Check
                              className={cn(
                                'mr-2 h-4 w-4',
                                selectedSnapshot?.metadata.name === snapshot.metadata.name
                                  ? 'opacity-100'
                                  : 'opacity-0',
                              )}
                            />
                            <div className="flex flex-col">
                              <span className="font-medium">
                                {snapshot.status.registryStatus?.tag || snapshot.metadata.name}
                              </span>
                              <span className="text-muted-foreground text-xs">
                                Source: {snapshot.status.targetName} | Size: {snapshot.status.sizeHuman || 'N/A'}
                                {snapshot.status.registryStatus?.pushedAt && (
                                  <> | Pushed: {formatDate(snapshot.status.registryStatus.pushedAt)}</>
                                )}
                              </span>
                            </div>
                          </CommandItem>
                        ))}
                      </CommandGroup>
                    </CommandList>
                  </Command>
                </PopoverContent>
              </Popover>
            </div>

            {/* Selected Snapshot Info */}
            {selectedSnapshot && (
              <div className="bg-muted space-y-2 rounded-lg p-4">
                <div className="text-sm font-medium">Snapshot Details</div>
                <div className="text-muted-foreground text-xs space-y-1">
                  <div>Tag: {selectedSnapshot.status.registryStatus?.tag || 'N/A'}</div>
                  <div>Source: {selectedSnapshot.status.targetName}</div>
                  <div>Owner: {selectedSnapshot.spec.ownedBy}</div>
                  <div>Size: {selectedSnapshot.status.sizeHuman || 'N/A'}</div>
                  {selectedSnapshot.status.registryStatus?.imageRef && (
                    <div className="font-mono text-xs break-all">
                      Image: {selectedSnapshot.status.registryStatus.imageRef}
                    </div>
                  )}
                </div>
                <div className="bg-background mt-2 rounded border p-2 text-xs">
                  <div className="font-medium">What will be restored:</div>
                  <ul className="text-muted-foreground mt-1 list-inside list-disc space-y-0.5">
                    <li>All files and directories from the snapshot</li>
                    <li>Package specifications (packages will be reinstalled)</li>
                    <li>Workspace configuration</li>
                  </ul>
                </div>
              </div>
            )}

            {/* New Workspace Name */}
            <div className="space-y-2">
              <Label htmlFor="name">New Workspace Name *</Label>
              <Input
                id="name"
                placeholder="my-new-workspace"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isPending}
                className="font-mono text-sm"
              />
              <p className="text-muted-foreground text-xs">
                The workspace name will be used to create the workspace resources
              </p>
            </div>

            {/* Info Notice */}
            <div className="bg-blue-50 dark:bg-blue-950/20 space-y-2 rounded-lg border border-blue-200 p-4 dark:border-blue-900">
              <div className="text-sm font-medium text-blue-900 dark:text-blue-100">
                Snapshot Cloning
              </div>
              <p className="text-xs text-blue-800 dark:text-blue-200">
                The snapshot will be pulled from the registry and a new workspace will be created.
                This process may take a few minutes depending on the snapshot size. The workspace
                will start automatically after restoration is complete.
              </p>
            </div>
          </div>

          <SheetFooter className="p-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isPending || !selectedSnapshot}>
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create from Snapshot
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
