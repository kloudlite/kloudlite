'use client'

import { useState, useTransition, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, Loader2, GitBranch } from 'lucide-react'
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
import { createWorkspace } from '@/app/actions/workspace.actions'
import { toast } from 'sonner'
import type { Visibility } from '@kloudlite/types'
import { VisibilitySelector } from '@/components/visibility-selector'

interface CreateWorkspaceSheetProps {
  namespace: string
  user: string
}

export function CreateWorkspaceSheet({ namespace, user }: CreateWorkspaceSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()

  // Basic fields
  const [name, setName] = useState('')

  // Visibility
  const [visibility, setVisibility] = useState<Visibility>('private')
  const [sharedWith, setSharedWith] = useState<string[]>([])

  // Git repository
  const [gitRepoUrl, setGitRepoUrl] = useState('')
  const [gitBranch, setGitBranch] = useState('')

  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      toast.error('Please enter a workspace name')
      return
    }

    startTransition(async () => {
      // Build git repository config if URL is provided
      const gitRepository = gitRepoUrl.trim()
        ? {
            url: gitRepoUrl.trim(),
            ...(gitBranch.trim() && { branch: gitBranch.trim() }),
          }
        : undefined

      const workspaceName = name
        .trim()
        .toLowerCase()
        .replace(/[^a-z0-9-]/g, '-')

      const result = await createWorkspace(namespace, {
        name: workspaceName,
        spec: {
          displayName: name.trim(),
          ownedBy: user,
          visibility,
          sharedWith: visibility === 'shared' ? sharedWith : undefined,
          gitRepository,
          status: 'active',
        },
      })

      if (result.success) {
        toast.success('Workspace created successfully')
        setOpen(false)
        setName('')
        setVisibility('private')
        setSharedWith([])
        setGitRepoUrl('')
        setGitBranch('')

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
        toast.error(result.error || 'Failed to create workspace')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button size="sm" className="gap-2">
          <Plus className="h-4 w-4" />
          New Workspace
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full sm:max-w-2xl">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Create Workspace</SheetTitle>
            <SheetDescription>
              Create a new development workspace
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-6 overflow-y-auto p-4">
            {/* Basic Information */}
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Workspace Name *</Label>
                <Input
                  id="name"
                  placeholder="my-workspace"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  disabled={isPending}
                  className="font-mono text-sm"
                />
                <p className="text-muted-foreground text-xs">
                  Lowercase letters, numbers, and hyphens only
                </p>
              </div>

              {/* Visibility */}
              <VisibilitySelector
                visibility={visibility}
                sharedWith={sharedWith}
                onVisibilityChange={setVisibility}
                onSharedWithChange={setSharedWith}
                disabled={isPending}
              />
            </div>

            {/* Git Repository Section */}
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <GitBranch className="h-4 w-4" />
                <Label>Git Repository (Optional)</Label>
              </div>
              <p className="text-muted-foreground text-sm">
                Clone a git repository when workspace starts. SSH keys from your WorkMachine will be used for authentication.
              </p>

              <div className="space-y-3">
                <div className="space-y-2">
                  <Label htmlFor="gitRepoUrl">Repository URL</Label>
                  <Input
                    id="gitRepoUrl"
                    placeholder="https://github.com/user/repo.git or git@github.com:user/repo.git"
                    value={gitRepoUrl}
                    onChange={(e) => setGitRepoUrl(e.target.value)}
                    disabled={isPending}
                    className="font-mono text-sm"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="gitBranch">Branch (Optional)</Label>
                  <Input
                    id="gitBranch"
                    placeholder="main"
                    value={gitBranch}
                    onChange={(e) => setGitBranch(e.target.value)}
                    disabled={isPending}
                    className="font-mono text-sm"
                  />
                  <p className="text-muted-foreground text-xs">
                    Leave empty to use repository&apos;s default branch
                  </p>
                </div>
              </div>
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
              Create Workspace
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
