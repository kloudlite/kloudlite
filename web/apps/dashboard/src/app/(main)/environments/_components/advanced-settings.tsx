'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'

interface AdvancedSettingsProps {
  environmentId: string
  resourceLimits: {
    cpu: string
    memory: string
  }
}

export function AdvancedSettings({ environmentId: _environmentId, resourceLimits }: AdvancedSettingsProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [cpu, setCpu] = useState(resourceLimits.cpu)
  const [memory, setMemory] = useState(resourceLimits.memory)

  const handleSave = () => {
    startTransition(async () => {
      // TODO: Implement update logic
      toast.success('Advanced settings updated')
      router.refresh()
    })
  }

  return (
    <div className="space-y-6">
      <div className="bg-card rounded-lg border p-6">
        <div className="space-y-4">
          <div>
            <h4 className="text-sm font-medium mb-4">Resource Limits</h4>
            <div className="space-y-3">
              <div>
                <label className="mb-2 block text-sm font-medium">CPU Limit</label>
                <Input
                  placeholder="e.g., 1000m or 1"
                  value={cpu}
                  onChange={(e) => setCpu(e.target.value)}
                  disabled={isPending}
                />
                <p className="text-muted-foreground text-xs mt-1.5">
                  Specify CPU limit in millicores (m) or cores
                </p>
              </div>

              <div>
                <label className="mb-2 block text-sm font-medium">Memory Limit</label>
                <Input
                  placeholder="e.g., 512Mi or 1Gi"
                  value={memory}
                  onChange={(e) => setMemory(e.target.value)}
                  disabled={isPending}
                />
                <p className="text-muted-foreground text-xs mt-1.5">
                  Specify memory limit in Mi or Gi
                </p>
              </div>
            </div>
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
