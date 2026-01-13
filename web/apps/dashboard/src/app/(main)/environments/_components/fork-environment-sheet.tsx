'use client'

import { useState, useTransition, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2, GitFork, ArrowLeft, HardDrive, Clock, ChevronRight } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import {
  Sheet,
  SheetContent,
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
  const [step, setStep] = useState<1 | 2>(1)

  const [name, setName] = useState('')
  const [nameError, setNameError] = useState<string | null>(null)
  const [selectedSnapshot, setSelectedSnapshot] = useState<Snapshot | null>(null)

  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current)
        pollIntervalRef.current = null
      }
    }
  }, [])

  useEffect(() => {
    if (open && sourceEnvironment) {
      setIsLoadingSnapshots(true)
      setSelectedSnapshot(null)
      setName('')
      setNameError(null)
      setStep(1)

      listReadySnapshots('environment', sourceEnvironment).then((result) => {
        if (result.success && result.data) {
          setSnapshots(result.data.snapshots || [])
        }
        setIsLoadingSnapshots(false)
      })
    }
  }, [open, sourceEnvironment])

  const validateName = (value: string): string | null => {
    if (!value) return 'Environment name is required'
    if (value.length > 63) return 'Name must be no more than 63 characters'
    if (value.includes('--')) return 'Name cannot contain "--"'
    const dnsLabelRegex = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
    if (!dnsLabelRegex.test(value)) {
      return 'Name must be lowercase alphanumeric or "-", start and end with alphanumeric'
    }
    return null
  }

  const handleNameChange = (value: string) => {
    const lowercaseValue = value.toLowerCase().replace(/[^a-z0-9-]/g, '-')
    setName(lowercaseValue)
    setNameError(validateName(lowercaseValue))
  }

  const handleSelectSnapshot = (snapshot: Snapshot) => {
    setSelectedSnapshot(snapshot)
    const baseName = sourceEnvironment.replace(/--/g, '-')
    setName(`${baseName}-fork`)
    setNameError(null)
    setStep(2)
  }

  const handleBack = () => {
    setStep(1)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const error = validateName(name)
    if (error) {
      setNameError(error)
      return
    }

    if (!selectedSnapshot) {
      toast.error('Please select a snapshot')
      return
    }

    if (!selectedSnapshot.namespace) {
      toast.error('Selected snapshot does not have a namespace')
      return
    }

    startTransition(async () => {
      const result = await createEnvironmentFromSnapshot({
        name: name.trim(),
        snapshotName: selectedSnapshot.name,
        sourceNamespace: selectedSnapshot.namespace,
        activated: true,
      })

      if (result.success) {
        toast.success('Environment created', {
          description: `"${name}" is being provisioned`,
        })
        onOpenChange(false)
        setName('')
        setSelectedSnapshot(null)
        setStep(1)

        if (onSuccess) {
          onSuccess()
        }

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
        toast.error(result.error || 'Failed to create environment')
      }
    })
  }

  const getRelativeTime = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMins / 60)
    const diffDays = Math.floor(diffHours / 24)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full sm:max-w-md p-0 flex flex-col gap-0 border-l">
        {/* Minimal Header */}
        <div className="flex items-center h-14 px-4 border-b shrink-0">
          {step === 2 ? (
            <button
              onClick={handleBack}
              disabled={isPending}
              className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              <span>Back</span>
            </button>
          ) : (
            <div className="flex items-center gap-2">
              <GitFork className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm font-medium">Fork Environment</span>
            </div>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto">
          {step === 1 ? (
            <div className="p-4">
              {/* Source Info */}
              <div className="mb-6">
                <p className="text-xs text-muted-foreground mb-1">Source</p>
                <p className="text-sm font-mono">{sourceEnvironment}</p>
              </div>

              {/* Snapshots */}
              <div>
                <p className="text-xs text-muted-foreground mb-3">Select snapshot</p>

                {isLoadingSnapshots ? (
                  <div className="flex items-center justify-center py-12">
                    <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
                  </div>
                ) : snapshots.length === 0 ? (
                  <div className="text-center py-12">
                    <p className="text-sm text-muted-foreground mb-1">No snapshots available</p>
                    <p className="text-xs text-muted-foreground">Create a snapshot first to fork this environment</p>
                  </div>
                ) : (
                  <div className="space-y-1">
                    {snapshots.map((snapshot) => (
                      <button
                        key={snapshot.name}
                        onClick={() => handleSelectSnapshot(snapshot)}
                        className={cn(
                          "w-full text-left px-3 py-3 rounded-md transition-colors",
                          "hover:bg-muted",
                          "focus:outline-none focus:bg-muted",
                          "group"
                        )}
                      >
                        <div className="flex items-center justify-between">
                          <div className="min-w-0 flex-1">
                            <p className="text-sm font-medium truncate pr-2">
                              {snapshot.name}
                            </p>
                            <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
                              {snapshot.sizeHuman && (
                                <span className="flex items-center gap-1">
                                  <HardDrive className="h-3 w-3" />
                                  {snapshot.sizeHuman}
                                </span>
                              )}
                              {snapshot.createdAt && (
                                <span className="flex items-center gap-1">
                                  <Clock className="h-3 w-3" />
                                  {getRelativeTime(snapshot.createdAt)}
                                </span>
                              )}
                            </div>
                          </div>
                          <ChevronRight className="h-4 w-4 text-muted-foreground/40 group-hover:text-muted-foreground transition-colors shrink-0" />
                        </div>
                      </button>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="p-4">
              {/* Selected Snapshot */}
              {selectedSnapshot && (
                <div className="mb-6">
                  <p className="text-xs text-muted-foreground mb-1">From snapshot</p>
                  <p className="text-sm font-mono truncate">{selectedSnapshot.name}</p>
                  <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
                    {selectedSnapshot.sizeHuman && <span>{selectedSnapshot.sizeHuman}</span>}
                    {selectedSnapshot.createdAt && (
                      <span>{getRelativeTime(selectedSnapshot.createdAt)}</span>
                    )}
                  </div>
                </div>
              )}

              {/* Name Input */}
              <div className="space-y-2">
                <Label htmlFor="name" className="text-xs text-muted-foreground">
                  Environment name
                </Label>
                <Input
                  id="name"
                  placeholder="my-environment"
                  value={name}
                  onChange={(e) => handleNameChange(e.target.value)}
                  disabled={isPending}
                  className={cn(
                    "font-mono text-sm",
                    nameError && "border-destructive focus-visible:ring-destructive"
                  )}
                  autoFocus
                />
                {nameError && (
                  <p className="text-xs text-destructive">{nameError}</p>
                )}
              </div>
            </form>
          )}
        </div>

        {/* Footer with action */}
        {step === 2 && (
          <div className="p-4 border-t shrink-0">
            <Button
              onClick={handleSubmit}
              className="w-full"
              disabled={isPending || !name || !!nameError}
            >
              {isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating...
                </>
              ) : (
                'Create Environment'
              )}
            </Button>
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
