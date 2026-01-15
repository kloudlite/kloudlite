'use client'

import { useState, useTransition, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2, GitFork } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import {
  Sheet,
  SheetContent,
} from '@kloudlite/ui'
import { forkEnvironment, getForkStatus } from '@/app/actions/snapshot.actions'
import { toast } from 'sonner'
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
  const [isCheckingStatus, setIsCheckingStatus] = useState(false)
  const [canFork, setCanFork] = useState(false)
  const [forkMessage, setForkMessage] = useState<string | null>(null)
  const [latestSnapshot, setLatestSnapshot] = useState<string | null>(null)

  const [name, setName] = useState('')
  const [nameError, setNameError] = useState<string | null>(null)

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
      setIsCheckingStatus(true)
      setName('')
      setNameError(null)
      setCanFork(false)
      setForkMessage(null)
      setLatestSnapshot(null)

      // Check fork status
      getForkStatus(sourceEnvironment).then((result) => {
        if (result.success && result.data) {
          setCanFork(result.data.canFork)
          setForkMessage(result.data.message || null)
          setLatestSnapshot(result.data.latestSnapshot || null)

          if (result.data.canFork) {
            // Set default name
            const baseName = sourceEnvironment.replace(/--/g, '-')
            setName(`${baseName}-fork`)
          }
        }
        setIsCheckingStatus(false)
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const error = validateName(name)
    if (error) {
      setNameError(error)
      return
    }

    startTransition(async () => {
      const result = await forkEnvironment(sourceEnvironment, name.trim())

      if (result.success) {
        toast.success('Environment forked', {
          description: `"${name}" is being provisioned from latest snapshot`,
        })
        onOpenChange(false)
        setName('')

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
        toast.error(result.error || 'Failed to fork environment')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full sm:max-w-md p-0 flex flex-col gap-0 border-l">
        {/* Header */}
        <div className="flex items-center h-14 px-4 border-b shrink-0">
          <div className="flex items-center gap-2">
            <GitFork className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium">Fork Environment</span>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto">
          <div className="p-4">
            {/* Source Info */}
            <div className="mb-6">
              <p className="text-xs text-muted-foreground mb-1">Source</p>
              <p className="text-sm font-mono">{sourceEnvironment}</p>
            </div>

            {isCheckingStatus ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
              </div>
            ) : !canFork ? (
              <div className="text-center py-12">
                <p className="text-sm text-muted-foreground mb-1">Cannot fork</p>
                <p className="text-xs text-muted-foreground">{forkMessage || 'No snapshots available. Create a snapshot first to fork this environment.'}</p>
              </div>
            ) : (
              <form onSubmit={handleSubmit}>
                {/* Snapshot Info */}
                {latestSnapshot && (
                  <div className="mb-6 p-3 bg-muted rounded-md">
                    <p className="text-xs text-muted-foreground mb-1">Using latest snapshot</p>
                    <p className="text-sm font-mono truncate">{latestSnapshot}</p>
                    <p className="text-xs text-muted-foreground mt-1">
                      You can restore to a different snapshot later if needed.
                    </p>
                  </div>
                )}

                {/* Name Input */}
                <div className="space-y-2">
                  <Label htmlFor="name" className="text-xs text-muted-foreground">
                    New environment name
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
        </div>

        {/* Footer with action */}
        {canFork && !isCheckingStatus && (
          <div className="p-4 border-t shrink-0">
            <Button
              onClick={handleSubmit}
              className="w-full"
              disabled={isPending || !name || !!nameError}
            >
              {isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Forking...
                </>
              ) : (
                'Fork'
              )}
            </Button>
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
