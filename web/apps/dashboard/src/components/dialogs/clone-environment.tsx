'use client'

import { useState, useEffect } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@kloudlite/ui'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import { Alert, AlertDescription } from '@kloudlite/ui'
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
import { Loader2, AlertCircle, Camera, Check, ChevronsUpDown } from 'lucide-react'
import { listPushedSnapshots, createEnvironmentFromSnapshot } from '@/app/actions/snapshot.actions'
import type { Snapshot } from '@/lib/services/snapshot.service'
import { cn } from '@/lib/utils'

interface CloneEnvironmentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
  preselectedSnapshot?: Snapshot
}

export function CloneEnvironmentDialog({
  open,
  onOpenChange,
  onSuccess,
  preselectedSnapshot,
}: CloneEnvironmentDialogProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isLoadingSnapshots, setIsLoadingSnapshots] = useState(false)
  const [snapshots, setSnapshots] = useState<Snapshot[]>([])
  const [popoverOpen, setPopoverOpen] = useState(false)
  const [selectedSnapshot, setSelectedSnapshot] = useState<Snapshot | null>(preselectedSnapshot || null)
  const [name, setName] = useState('')

  // Load pushed snapshots when dialog opens
  useEffect(() => {
    if (open) {
      setIsLoadingSnapshots(true)
      listPushedSnapshots('environment').then((result) => {
        if (result.success && result.data) {
          setSnapshots(result.data.snapshots || [])
        }
        setIsLoadingSnapshots(false)
      })

      // Set preselected snapshot if provided
      if (preselectedSnapshot) {
        setSelectedSnapshot(preselectedSnapshot)
        const sourceName = preselectedSnapshot.status.targetName || preselectedSnapshot.metadata.name
        setName(`${sourceName}-clone`)
      } else {
        setSelectedSnapshot(null)
        setName('')
      }
      setError(null)
    }
  }, [open, preselectedSnapshot])

  const validateNamespace = (name: string): string | null => {
    if (!name) {
      return 'Namespace name is required'
    }
    if (name.length > 63) {
      return 'Namespace name must be no more than 63 characters'
    }
    if (name.includes('--')) {
      return 'Environment name cannot contain "--" (double hyphens)'
    }
    const dnsLabelRegex = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
    if (!dnsLabelRegex.test(name)) {
      return 'Namespace name must consist of lower case alphanumeric characters or "-", and must start and end with an alphanumeric character'
    }

    const reservedNamespaces = [
      'kube-system',
      'kube-public',
      'kube-node-lease',
      'default',
      'kloudlite-system',
    ]

    if (reservedNamespaces.includes(name)) {
      return `Cannot use reserved namespace name: ${name}`
    }

    const reservedPrefixes = ['kube-', 'openshift-', 'kubernetes-']
    for (const prefix of reservedPrefixes) {
      if (name.startsWith(prefix)) {
        return `Cannot use namespace name with reserved prefix: ${prefix}`
      }
    }

    return null
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    if (!selectedSnapshot) {
      setError('Please select a snapshot')
      return
    }

    // Validate environment name
    const nameError = validateNamespace(name)
    if (nameError) {
      setError(`Environment name: ${nameError}`)
      return
    }

    setLoading(true)

    try {
      const result = await createEnvironmentFromSnapshot({
        name,
        snapshotName: selectedSnapshot.metadata.name,
        activated: true,
      })

      if (result.success) {
        // Reset form
        setName('')
        setSelectedSnapshot(null)
        onOpenChange(false)

        // Call success callback
        if (onSuccess) {
          onSuccess()
        }
      } else {
        setError(result.error || 'Failed to create environment from snapshot. Please try again.')
      }
    } catch (err) {
      console.error('Failed to create environment from snapshot:', err)
      const error = err instanceof Error ? err : new Error('Failed to create environment from snapshot')
      setError(error.message)
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    if (!loading) {
      setName('')
      setSelectedSnapshot(null)
      setError(null)
      onOpenChange(false)
    }
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
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Camera className="h-5 w-5" />
              Create Environment from Snapshot
            </DialogTitle>
            <DialogDescription>
              Create a new environment from a pushed snapshot. The snapshot will be pulled from the
              registry and all resources will be restored.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
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
                    disabled={loading || isLoadingSnapshots}
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
                                Source: {snapshot.spec.ownedBy}/{snapshot.status.targetName} | Size: {snapshot.status.sizeHuman || 'N/A'}
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
              <div className="bg-muted space-y-2 rounded-lg p-3">
                <div className="text-sm font-medium">Snapshot Details</div>
                <div className="text-muted-foreground text-xs space-y-1">
                  <div>Tag: {selectedSnapshot.status.registryStatus?.tag || 'N/A'}</div>
                  <div>Source: {selectedSnapshot.spec.ownedBy}/{selectedSnapshot.status.targetName}</div>
                  <div>Size: {selectedSnapshot.status.sizeHuman || 'N/A'}</div>
                </div>
              </div>
            )}

            {/* New Environment Name */}
            <div className="space-y-2">
              <Label htmlFor="name">New Environment Name *</Label>
              <Input
                id="name"
                placeholder="my-cloned-environment"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={loading}
                required
              />
              <p className="text-muted-foreground text-xs">
                Must be lowercase alphanumeric or &quot;-&quot;, max 63 characters
              </p>
            </div>

            {error && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading || !selectedSnapshot}>
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {loading ? 'Creating...' : 'Create from Snapshot'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
