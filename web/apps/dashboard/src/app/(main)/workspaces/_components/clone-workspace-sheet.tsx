'use client'

import { useState, useTransition, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Loader2 } from 'lucide-react'
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
import { cloneWorkspace } from '@/app/actions/workspace.actions'
import { toast } from 'sonner'
import type { Workspace } from '@kloudlite/types'

interface CloneWorkspaceSheetProps {
  workspace: Workspace
  trigger?: React.ReactNode
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

export function CloneWorkspaceSheet({
  workspace,
  trigger,
  open: controlledOpen,
  onOpenChange
}: CloneWorkspaceSheetProps) {
  const router = useRouter()
  const [internalOpen, setInternalOpen] = useState(false)
  const [isPending, startTransition] = useTransition()

  // Use controlled state if provided, otherwise use internal state
  const open = controlledOpen !== undefined ? controlledOpen : internalOpen
  const setOpen = onOpenChange || setInternalOpen

  // Form fields
  const [name, setName] = useState('')

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

  // Reset form when sheet opens
  useEffect(() => {
    if (open) {
      // Suggest a default name based on source workspace
      const suggestedName = `${workspace.spec.displayName}-clone`
      setName(suggestedName)
    }
  }, [open, workspace])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      toast.error('Please enter a workspace name')
      return
    }

    startTransition(async () => {
      // Generate folder name from workspace name
      const normalizedName = name
        .trim()
        .toLowerCase()
        .replace(/[^a-z0-9-]/g, '-')

      // Clone the entire workspace spec with only name changes
      const result = await cloneWorkspace(
        workspace.metadata.name,
        {
          name: normalizedName,
          spec: {
            ...workspace.spec,
            displayName: name.trim(),
            folderName: normalizedName,
            status: 'active',
          },
        },
        workspace.metadata.namespace,
      )

      if (result.success) {
        toast.success('Workspace cloning initiated')
        setOpen(false)
        setName('')

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
        toast.error(result.error || 'Failed to clone workspace')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      {trigger && <SheetTrigger asChild>{trigger}</SheetTrigger>}
      <SheetContent side="right" className="w-full sm:max-w-lg">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Clone Workspace</SheetTitle>
            <SheetDescription>
              Create a copy of &quot;{workspace.spec.displayName}&quot; with all files from the
              workspace directory
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-6 overflow-y-auto p-4">
            {/* Source Workspace Info */}
            <div className="bg-muted space-y-2 rounded-lg p-4">
              <div className="text-sm font-medium">Source Workspace</div>
              <div className="text-muted-foreground text-xs space-y-1">
                <div>Name: {workspace.spec.displayName}</div>
                <div>Folder: {workspace.spec.folderName || workspace.metadata.name}</div>
                {workspace.spec.workmachineName && (
                  <div>WorkMachine: {workspace.spec.workmachineName}</div>
                )}
              </div>
              <div className="bg-background mt-2 rounded border p-2 text-xs">
                <div className="font-medium">What will be cloned:</div>
                <ul className="text-muted-foreground mt-1 list-inside list-disc space-y-0.5">
                  <li>All files and directories from the workspace folder</li>
                  <li>All workspace settings and configuration</li>
                  <li>Package specifications (packages will be reinstalled)</li>
                  <li>Git repository configuration</li>
                  <li>Resource quotas and limits</li>
                </ul>
              </div>
            </div>

            {/* New Workspace Details */}
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">New Workspace Name *</Label>
                <Input
                  id="name"
                  placeholder="my-workspace-clone"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  disabled={isPending}
                  className="font-mono text-sm"
                  autoFocus
                />
                <p className="text-muted-foreground text-xs">
                  The workspace folder name will be automatically generated from this name
                </p>
              </div>
            </div>

            {/* Cloning Notice */}
            <div className="bg-blue-50 dark:bg-blue-950/20 space-y-2 rounded-lg border border-blue-200 p-4 dark:border-blue-900">
              <div className="text-sm font-medium text-blue-900 dark:text-blue-100">
                Cloning Process
              </div>
              <p className="text-xs text-blue-800 dark:text-blue-200">
                The cloning process will temporarily suspend the source workspace while copying
                files. The source workspace will automatically resume after cloning completes. This
                ensures data consistency during the copy operation.
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
            <Button type="submit" disabled={isPending}>
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Clone Workspace
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
