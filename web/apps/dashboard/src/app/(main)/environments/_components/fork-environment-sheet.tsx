'use client'

import { useState, useTransition, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2, GitFork, Check, ArrowLeft, ArrowRight, HardDrive, Clock, ChevronRight } from 'lucide-react'
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

  // Form fields
  const [name, setName] = useState('')
  const [nameError, setNameError] = useState<string | null>(null)
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

  // Reset state when sheet opens/closes
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
    // Generate suggested name
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

    startTransition(async () => {
      const result = await createEnvironmentFromSnapshot({
        name: name.trim(),
        snapshotName: selectedSnapshot.name,
        activated: true,
      })

      if (result.success) {
        toast.success('Environment creation initiated', {
          description: `Creating "${name}" from snapshot`,
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
    if (diffMins < 60) return `${diffMins} min ago`
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
    if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full sm:max-w-lg p-0 flex flex-col gap-0">
        {/* Header */}
        <div className="border-b">
          <div className="px-6 py-5">
            <div className="flex items-center gap-4">
              {step === 2 && (
                <button
                  onClick={handleBack}
                  disabled={isPending}
                  className="p-1.5 -ml-1.5 rounded-md hover:bg-muted transition-colors"
                >
                  <ArrowLeft className="h-4 w-4" />
                </button>
              )}
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <GitFork className="h-5 w-5 text-primary" />
                  <h2 className="text-lg font-semibold">Fork Environment</h2>
                </div>
                <p className="text-sm text-muted-foreground">
                  {step === 1
                    ? 'Select a snapshot to create a new environment'
                    : 'Name your new environment'
                  }
                </p>
              </div>
              {/* Step indicator */}
              <div className="flex items-center gap-2">
                <div className={cn(
                  "flex items-center justify-center h-6 w-6 rounded-full text-xs font-medium transition-colors",
                  step === 1 ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
                )}>1</div>
                <div className="w-4 h-px bg-border" />
                <div className={cn(
                  "flex items-center justify-center h-6 w-6 rounded-full text-xs font-medium transition-colors",
                  step === 2 ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
                )}>2</div>
              </div>
            </div>
          </div>
          {/* Source environment badge */}
          <div className="px-6 pb-4">
            <div className="inline-flex items-center gap-2 px-3 py-1.5 bg-muted rounded-full text-xs">
              <span className="text-muted-foreground">Source:</span>
              <span className="font-medium">{sourceEnvironment}</span>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto">
          {step === 1 ? (
            /* Step 1: Snapshot Selection */
            <div className="p-6">
              {isLoadingSnapshots ? (
                <div className="flex flex-col items-center justify-center py-16">
                  <Loader2 className="h-8 w-8 animate-spin text-muted-foreground mb-4" />
                  <p className="text-sm text-muted-foreground">Loading snapshots...</p>
                </div>
              ) : snapshots.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-16 text-center">
                  <div className="rounded-full bg-muted p-4 mb-4">
                    <GitFork className="h-8 w-8 text-muted-foreground" />
                  </div>
                  <h3 className="font-semibold mb-2">No Snapshots Available</h3>
                  <p className="text-sm text-muted-foreground max-w-[280px]">
                    Create a snapshot of this environment first to enable forking.
                  </p>
                </div>
              ) : (
                <div className="space-y-3">
                  <p className="text-xs uppercase tracking-wider text-muted-foreground font-medium mb-4">
                    Available Snapshots ({snapshots.length})
                  </p>
                  {snapshots.map((snapshot) => (
                    <button
                      key={snapshot.name}
                      onClick={() => handleSelectSnapshot(snapshot)}
                      className={cn(
                        "w-full text-left rounded-lg border bg-card p-4 transition-all duration-200",
                        "hover:shadow-md hover:border-primary/40",
                        "focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary",
                        "group"
                      )}
                    >
                      <div className="flex items-center gap-4">
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 mb-2">
                            <span className="font-medium text-sm truncate">
                              {snapshot.name}
                            </span>
                          </div>
                          {snapshot.description && (
                            <p className="text-xs text-muted-foreground mb-3 line-clamp-1">
                              {snapshot.description}
                            </p>
                          )}
                          <div className="flex items-center gap-4 text-xs text-muted-foreground">
                            {snapshot.sizeHuman && (
                              <span className="flex items-center gap-1.5">
                                <HardDrive className="h-3.5 w-3.5" />
                                {snapshot.sizeHuman}
                              </span>
                            )}
                            {snapshot.createdAt && (
                              <span className="flex items-center gap-1.5">
                                <Clock className="h-3.5 w-3.5" />
                                {getRelativeTime(snapshot.createdAt)}
                              </span>
                            )}
                          </div>
                        </div>
                        <ChevronRight className="h-5 w-5 text-muted-foreground/50 group-hover:text-primary transition-colors" />
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
          ) : (
            /* Step 2: Environment Configuration */
            <form onSubmit={handleSubmit} className="p-6">
              {/* Selected Snapshot Summary */}
              {selectedSnapshot && (
                <div className="rounded-lg border bg-muted/50 p-4 mb-6">
                  <p className="text-xs uppercase tracking-wider text-muted-foreground font-medium mb-2">
                    Creating from snapshot
                  </p>
                  <p className="font-medium text-sm mb-1">{selectedSnapshot.name}</p>
                  <div className="flex items-center gap-3 text-xs text-muted-foreground">
                    {selectedSnapshot.sizeHuman && (
                      <span className="flex items-center gap-1">
                        <HardDrive className="h-3 w-3" />
                        {selectedSnapshot.sizeHuman}
                      </span>
                    )}
                    {selectedSnapshot.createdAt && (
                      <span className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {getRelativeTime(selectedSnapshot.createdAt)}
                      </span>
                    )}
                  </div>
                </div>
              )}

              {/* Environment Name Input */}
              <div className="space-y-2 mb-6">
                <Label htmlFor="name" className="text-sm font-medium">
                  Environment Name
                </Label>
                <Input
                  id="name"
                  placeholder="my-new-environment"
                  value={name}
                  onChange={(e) => handleNameChange(e.target.value)}
                  disabled={isPending}
                  className={cn(
                    "font-mono",
                    nameError && "border-destructive focus-visible:ring-destructive"
                  )}
                  autoFocus
                />
                {nameError ? (
                  <p className="text-xs text-destructive">{nameError}</p>
                ) : (
                  <p className="text-xs text-muted-foreground">
                    Lowercase letters, numbers, and hyphens only
                  </p>
                )}
              </div>

              {/* Submit Button */}
              <Button
                type="submit"
                className="w-full"
                disabled={isPending || !name || !!nameError}
              >
                {isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Creating...
                  </>
                ) : (
                  <>
                    Create Environment
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </>
                )}
              </Button>
            </form>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
