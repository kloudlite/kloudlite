'use client'

import { useState, useTransition, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Copy, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { cloneWorkspace } from '@/app/actions/workspace.actions'
import { toast } from 'sonner'
import type { Workspace } from '@/types/workspace'

interface CloneWorkspaceSheetProps {
  workspace: Workspace
  trigger?: React.ReactNode
}

export function CloneWorkspaceSheet({ workspace, trigger }: CloneWorkspaceSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()

  // Form fields
  const [name, setName] = useState('')
  const [folderName, setFolderName] = useState('')

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
      // Suggest a default folder name
      const suggestedFolder = `${workspace.spec.folderName || workspace.metadata.name}-clone`
      setFolderName(suggestedFolder)
    }
  }, [open, workspace])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      toast.error('Please enter a workspace name')
      return
    }

    if (!folderName.trim()) {
      toast.error('Please enter a folder name')
      return
    }

    startTransition(async () => {
      const result = await cloneWorkspace(
        workspace.metadata.name,
        {
          name: name
            .trim()
            .toLowerCase()
            .replace(/[^a-z0-9-]/g, '-'),
          spec: {
            displayName: name.trim(),
            ownedBy: workspace.spec.ownedBy,
            folderName: folderName.trim(),
            workmachineName: workspace.spec.workmachineName,
            status: 'active',
          },
        },
        workspace.metadata.namespace,
      )

      if (result.success) {
        toast.success('Workspace cloning initiated')
        setOpen(false)
        setName('')
        setFolderName('')

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
      {trigger ? (
        <SheetTrigger asChild>{trigger}</SheetTrigger>
      ) : (
        <SheetTrigger asChild>
          <Button size="sm" variant="outline" className="gap-2">
            <Copy className="h-4 w-4" />
            Clone
          </Button>
        </SheetTrigger>
      )}
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
                </ul>
                <div className="text-muted-foreground mt-2 font-medium">What will NOT be cloned:</div>
                <ul className="text-muted-foreground mt-1 list-inside list-disc space-y-0.5">
                  <li>Installed packages (will need to be reinstalled)</li>
                  <li>Workspace settings</li>
                  <li>Environment connections</li>
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
                />
                <p className="text-muted-foreground text-xs">
                  Lowercase letters, numbers, and hyphens only
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="folderName">Folder Name *</Label>
                <Input
                  id="folderName"
                  placeholder="my-workspace-clone"
                  value={folderName}
                  onChange={(e) => setFolderName(e.target.value)}
                  disabled={isPending}
                  className="font-mono text-sm"
                />
                <p className="text-muted-foreground text-xs">
                  Name of the directory on the WorkMachine where files will be stored
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
