'use client'

import { useState, useTransition, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2, Camera, Check, ArrowLeft, ArrowRight, Calendar, HardDrive, Clock } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
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

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
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
    return formatDate(dateString)
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full sm:max-w-xl p-0 flex flex-col">
        {/* Header */}
        <SheetHeader className="px-6 py-4 border-b bg-muted/30">
          <div className="flex items-center gap-3">
            {step === 2 && (
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 shrink-0"
                onClick={handleBack}
                disabled={isPending}
              >
                <ArrowLeft className="h-4 w-4" />
              </Button>
            )}
            <div className="flex-1 min-w-0">
              <SheetTitle className="text-lg">
                {step === 1 ? 'Select Snapshot' : 'Create Environment'}
              </SheetTitle>
              <SheetDescription className="text-sm mt-0.5">
                {step === 1
                  ? `Choose a snapshot from ${sourceEnvironment} to fork`
                  : 'Configure your new environment'
                }
              </SheetDescription>
            </div>
            {/* Step indicator */}
            <div className="flex items-center gap-1.5 shrink-0">
              <div className={cn(
                "h-2 w-2 rounded-full transition-colors",
                step === 1 ? "bg-primary" : "bg-muted-foreground/30"
              )} />
              <div className={cn(
                "h-2 w-2 rounded-full transition-colors",
                step === 2 ? "bg-primary" : "bg-muted-foreground/30"
              )} />
            </div>
          </div>
        </SheetHeader>

        {/* Content */}
        <div className="flex-1 overflow-y-auto">
          {step === 1 ? (
            /* Step 1: Snapshot Selection */
            <div className="p-4">
              {isLoadingSnapshots ? (
                <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                  <Loader2 className="h-8 w-8 animate-spin mb-3" />
                  <p className="text-sm">Loading snapshots...</p>
                </div>
              ) : snapshots.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <div className="rounded-full bg-muted p-4 mb-4">
                    <Camera className="h-8 w-8 text-muted-foreground" />
                  </div>
                  <h3 className="font-medium mb-1">No Snapshots Available</h3>
                  <p className="text-sm text-muted-foreground max-w-[280px]">
                    Create a snapshot of this environment first before you can fork it.
                  </p>
                </div>
              ) : (
                <div className="space-y-2">
                  {snapshots.map((snapshot) => (
                    <button
                      key={snapshot.name}
                      onClick={() => handleSelectSnapshot(snapshot)}
                      className={cn(
                        "w-full text-left rounded-lg border p-4 transition-all",
                        "hover:border-primary/50 hover:bg-accent/50",
                        "focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2",
                        selectedSnapshot?.name === snapshot.name && "border-primary bg-accent"
                      )}
                    >
                      <div className="flex items-start gap-3">
                        <div className={cn(
                          "rounded-lg p-2 shrink-0",
                          "bg-primary/10 text-primary"
                        )}>
                          <Camera className="h-5 w-5" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="font-medium truncate">
                              {snapshot.name}
                            </span>
                            {selectedSnapshot?.name === snapshot.name && (
                              <Check className="h-4 w-4 text-primary shrink-0" />
                            )}
                          </div>
                          {snapshot.description && (
                            <p className="text-sm text-muted-foreground mb-2 line-clamp-2">
                              {snapshot.description}
                            </p>
                          )}
                          <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
                            {snapshot.sizeHuman && (
                              <span className="flex items-center gap-1">
                                <HardDrive className="h-3 w-3" />
                                {snapshot.sizeHuman}
                              </span>
                            )}
                            {snapshot.createdAt && (
                              <>
                                <span className="flex items-center gap-1">
                                  <Calendar className="h-3 w-3" />
                                  {formatDate(snapshot.createdAt)}
                                </span>
                                <span className="flex items-center gap-1">
                                  <Clock className="h-3 w-3" />
                                  {getRelativeTime(snapshot.createdAt)}
                                </span>
                              </>
                            )}
                          </div>
                        </div>
                        <ArrowRight className="h-4 w-4 text-muted-foreground shrink-0 mt-1" />
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
          ) : (
            /* Step 2: Environment Configuration */
            <form onSubmit={handleSubmit} className="p-6 space-y-6">
              {/* Selected Snapshot Summary */}
              {selectedSnapshot && (
                <div className="rounded-lg border bg-muted/30 p-4">
                  <div className="flex items-center gap-3">
                    <div className="rounded-lg bg-primary/10 p-2">
                      <Camera className="h-5 w-5 text-primary" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-xs text-muted-foreground uppercase tracking-wide mb-0.5">
                        Forking from
                      </p>
                      <p className="font-medium truncate">{selectedSnapshot.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {selectedSnapshot.sizeHuman}
                        {selectedSnapshot.createdAt && ` • ${getRelativeTime(selectedSnapshot.createdAt)}`}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* Environment Name Input */}
              <div className="space-y-3">
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
                    "font-mono h-11",
                    nameError && "border-destructive focus-visible:ring-destructive"
                  )}
                  autoFocus
                />
                {nameError ? (
                  <p className="text-xs text-destructive">{nameError}</p>
                ) : (
                  <p className="text-xs text-muted-foreground">
                    Lowercase letters, numbers, and hyphens only. Max 63 characters.
                  </p>
                )}
              </div>

              {/* Info Box */}
              <div className="rounded-lg border border-blue-200 bg-blue-50/50 dark:border-blue-900 dark:bg-blue-950/20 p-4">
                <h4 className="text-sm font-medium text-blue-900 dark:text-blue-100 mb-2">
                  What happens next?
                </h4>
                <ul className="text-xs text-blue-800 dark:text-blue-200 space-y-1.5">
                  <li className="flex items-start gap-2">
                    <span className="rounded-full bg-blue-200 dark:bg-blue-800 w-1.5 h-1.5 mt-1.5 shrink-0" />
                    A new environment will be created with all data from the snapshot
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="rounded-full bg-blue-200 dark:bg-blue-800 w-1.5 h-1.5 mt-1.5 shrink-0" />
                    The environment will be activated automatically
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="rounded-full bg-blue-200 dark:bg-blue-800 w-1.5 h-1.5 mt-1.5 shrink-0" />
                    This may take a few minutes depending on snapshot size
                  </li>
                </ul>
              </div>

              {/* Submit Button */}
              <div className="pt-2">
                <Button
                  type="submit"
                  className="w-full h-11"
                  disabled={isPending || !name || !!nameError}
                >
                  {isPending ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Creating Environment...
                    </>
                  ) : (
                    <>
                      Create Environment
                      <ArrowRight className="ml-2 h-4 w-4" />
                    </>
                  )}
                </Button>
              </div>
            </form>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
