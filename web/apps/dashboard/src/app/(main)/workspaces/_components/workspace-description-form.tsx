'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { FileText } from 'lucide-react'
import { Button, Input, Label } from '@kloudlite/ui'
import { updateWorkspace } from '@/app/actions/workspace.actions'

interface WorkspaceDescriptionFormProps {
  workspaceName: string
  namespace: string
  currentDisplayName: string
}

export function WorkspaceDescriptionForm({
  workspaceName,
  namespace,
  currentDisplayName,
}: WorkspaceDescriptionFormProps) {
  const router = useRouter()
  const [displayName, setDisplayName] = useState(currentDisplayName)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const hasChanges = displayName !== currentDisplayName

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!hasChanges) return

    setIsLoading(true)
    setError(null)

    try {
      const result = await updateWorkspace(workspaceName, namespace, { displayName })
      if (result.success) {
        router.refresh()
      } else {
        setError(result.error || 'Failed to update workspace')
      }
    } catch (err) {
      setError('An unexpected error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="bg-card rounded-lg border">
      <div className="border-b p-4">
        <div className="flex items-center gap-2">
          <FileText className="h-4 w-4 text-muted-foreground" />
          <h3 className="text-sm font-semibold">Description</h3>
        </div>
      </div>
      <form onSubmit={handleSubmit} className="p-4 space-y-4">
        <div className="space-y-2">
          <Label htmlFor="displayName">Display Name</Label>
          <Input
            id="displayName"
            value={displayName}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setDisplayName(e.target.value)}
            placeholder="Enter a display name for your workspace"
          />
          <p className="text-xs text-muted-foreground">
            A friendly name to help identify this workspace.
          </p>
        </div>

        {error && <p className="text-sm text-destructive">{error}</p>}

        <div className="flex justify-end">
          <Button type="submit" disabled={!hasChanges || isLoading} size="sm">
            {isLoading ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </form>
    </div>
  )
}
