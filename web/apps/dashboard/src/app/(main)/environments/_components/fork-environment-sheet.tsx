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
import { listReadySnapshots, createEnvironmentFromSnapshot } from '@/app/actions/snapshot.actions'
import { toast } from 'sonner'
import type { Snapshot } from '@/lib/services/snapshot.service'
import { cn } from '@/lib/utils'

interface ForkEnvironmentSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sourceEnvironment: string
  onSuccess?: () => void
}

export function ForkEnvironmentSheet({
  open,
  onOpenChange,
  sourceEnvironment,
  onSuccess,
}: ForkEnvironmentSheetProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [isLoadingSnapshots, setIsLoadingSnapshots] = useState(false)
  const [snapshots, setSnapshots] = useState<Snapshot[]>([])
  const [popoverOpen, setPopoverOpen] = useState(false)

  // Form fields
  const [name, setName] = useState('')
  const [selectedSnapshot, setSelectedSnapshot] = useState<Snapshot | null>(null)

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

  // Load snapshots for this environment when sheet opens
  useEffect(() => {
    if (open && sourceEnvironment) {
      setIsLoadingSnapshots(true)
      setSelectedSnapshot(null)
      setName('')

      listReadySnapshots('environment', sourceEnvironment).then((result) => {
        if (result.success && result.data) {
          setSnapshots(result.data.snapshots || [])
        }
        setIsLoadingSnapshots(false)
      })
    }
  }, [open, sourceEnvironment])

  const validateName = (name: string): string | null => {
    if (!name) return 'Environment name is required'
    if (name.length > 63) return 'Name must be no more than 63 characters'
    if (name.includes('--')) return 'Name cannot contain "--"'
    const dnsLabelRegex = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
    if (!dnsLabelRegex.test(name)) {
      return 'Name must be lowercase alphanumeric or "-", start and end with alphanumeric'
    }
    return null
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const nameError = validateName(name)
    if (nameError) {
      toast.error(nameError)
      return
    }

    if (!selectedSnapshot) {
      toast.error('Please select a snapshot')
      return
    }

    startTransition(async () => {
      const result = await createEnvironmentFromSnapshot({
        name: name.trim(),
        snapshotName: selectedSnapshot.name,
        activated: true,
      })

      if (result.success) {
        toast.success('Environment creation initiated from snapshot')
        onOpenChange(false)
        setName('')
        setSelectedSnapshot(null)

        if (onSuccess) {
          onSuccess()
        }

        // Refresh and poll for updates
        router.refresh()

        if (pollIntervalRef.current) {
          clearInterval(pollIntervalRef.current)
        }

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
        toast.error(result.error || 'Failed to create environment from snapshot')
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
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full sm:max-w-lg">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Fork Environment</SheetTitle>
            <SheetDescription>
              Create a new environment from a snapshot of <strong>{sourceEnvironment}</strong>
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
                        {selectedSnapshot.name}
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
                          ? 'No snapshots available for this environment. Create a snapshot first.'
                          : 'No matching snapshots found.'}
                      </CommandEmpty>
                      <CommandGroup>
                        {snapshots.map((snapshot) => (
                          <CommandItem
                            key={snapshot.name}
                            value={snapshot.name}
                            onSelect={() => {
                              setSelectedSnapshot(snapshot)
                              setPopoverOpen(false)
                              if (!name) {
                                setName(`${sourceEnvironment}-fork`)
                              }
                            }}
                          >
                            <Check
                              className={cn(
                                'mr-2 h-4 w-4',
                                selectedSnapshot?.name === snapshot.name
                                  ? 'opacity-100'
                                  : 'opacity-0',
                              )}
                            />
                            <div className="flex flex-col">
                              <span className="font-medium">{snapshot.name}</span>
                              <span className="text-muted-foreground text-xs">
                                Size: {snapshot.sizeHuman || 'N/A'}
                                {snapshot.createdAt && (
                                  <> | {formatDate(snapshot.createdAt)}</>
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
                  <div>Name: {selectedSnapshot.name}</div>
                  <div>Size: {selectedSnapshot.sizeHuman || 'N/A'}</div>
                  {selectedSnapshot.description && (
                    <div>Description: {selectedSnapshot.description}</div>
                  )}
                  {selectedSnapshot.createdAt && (
                    <div>Created: {formatDate(selectedSnapshot.createdAt)}</div>
                  )}
                </div>
              </div>
            )}

            {/* New Environment Name */}
            <div className="space-y-2">
              <Label htmlFor="name">New Environment Name *</Label>
              <Input
                id="name"
                placeholder="my-new-environment"
                value={name}
                onChange={(e) => setName(e.target.value.toLowerCase())}
                disabled={isPending}
                className="font-mono text-sm"
              />
              <p className="text-muted-foreground text-xs">
                Lowercase alphanumeric or &quot;-&quot;, max 63 characters
              </p>
            </div>
          </div>

          <SheetFooter className="p-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isPending || !selectedSnapshot}>
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Environment
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
