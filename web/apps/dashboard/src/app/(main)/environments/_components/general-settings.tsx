'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Textarea } from '@kloudlite/ui'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'

interface GeneralSettingsProps {
  environmentId: string
  environmentName: string
  description: string
  ownedBy: string
}

export function GeneralSettings({ environmentId: _environmentId, environmentName, description, ownedBy }: GeneralSettingsProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [currentDescription, setCurrentDescription] = useState(description)

  const handleSave = () => {
    startTransition(async () => {
      // TODO: Implement update logic
      toast.success('Settings updated')
      router.refresh()
    })
  }

  return (
    <div className="space-y-6">
      <div className="bg-card rounded-lg border p-6">
        <div className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium">Environment Name</label>
            <Input
              value={environmentName}
              disabled
              className="bg-muted"
            />
            <p className="text-muted-foreground text-xs mt-1.5">
              Environment name cannot be changed after creation
            </p>
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium">Owner</label>
            <Input
              value={ownedBy}
              disabled
              className="bg-muted"
            />
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium">Description</label>
            <Textarea
              placeholder="Add a description for this environment"
              value={currentDescription}
              onChange={(e) => setCurrentDescription(e.target.value)}
              disabled={isPending}
              rows={3}
            />
          </div>

          <div className="flex justify-end pt-2">
            <Button onClick={handleSave} disabled={isPending}>
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Save Changes
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
