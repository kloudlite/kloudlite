'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { updateEnvironment } from '@/app/actions/environment.actions'
import { toast } from 'sonner'
import type { EnvironmentUIModel } from '@/types/environment'

interface EditEnvironmentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  environment: EnvironmentUIModel
  onSuccess?: () => void
  currentUser: string
}

export function EditEnvironmentDialog({
  open,
  onOpenChange,
  environment,
  onSuccess,
  currentUser: _currentUser,
}: EditEnvironmentDialogProps) {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Prevent editing if environment is in a transitional state
  const isTransitional = ['deleting', 'activating', 'deactivating'].includes(environment.status)

  const [formData, setFormData] = useState({
    cpuRequests: '',
    memoryRequests: '',
    cpuLimits: '',
    memoryLimits: '',
    storageRequests: '',
    allowIngress: false,
    allowEgress: false,
    isolateNamespace: false,
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)

    try {
      const updateData = {
        spec: {
          targetNamespace: environment.targetNamespace,
          activated: environment.status === 'active',
          ownedBy: _currentUser,
        },
      }

      const result = await updateEnvironment(environment.name, updateData)

      if (result.success) {
        toast.success('Environment updated', {
          description: `${environment.name} has been updated successfully.`,
        })

        onOpenChange(false)
        onSuccess?.()
        router.refresh()
      } else {
        toast.error('Failed to update environment', {
          description: result.error || 'An error occurred while updating the environment',
        })
      }
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error')
      toast.error('Failed to update environment', {
        description: error.message,
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Edit Environment</DialogTitle>
            <DialogDescription>
              Update resource quotas and network policies for {environment.name}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-6 py-4">
            <div className="space-y-4">
              <h3 className="text-sm font-medium">Resource Quotas</h3>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="cpu-requests">CPU Requests</Label>
                  <Input
                    id="cpu-requests"
                    placeholder="e.g., 500m, 2"
                    value={formData.cpuRequests}
                    onChange={(e) => setFormData({ ...formData, cpuRequests: e.target.value })}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="cpu-limits">CPU Limits</Label>
                  <Input
                    id="cpu-limits"
                    placeholder="e.g., 1000m, 4"
                    value={formData.cpuLimits}
                    onChange={(e) => setFormData({ ...formData, cpuLimits: e.target.value })}
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="memory-requests">Memory Requests</Label>
                  <Input
                    id="memory-requests"
                    placeholder="e.g., 256Mi, 1Gi"
                    value={formData.memoryRequests}
                    onChange={(e) => setFormData({ ...formData, memoryRequests: e.target.value })}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="memory-limits">Memory Limits</Label>
                  <Input
                    id="memory-limits"
                    placeholder="e.g., 512Mi, 2Gi"
                    value={formData.memoryLimits}
                    onChange={(e) => setFormData({ ...formData, memoryLimits: e.target.value })}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="storage-requests">Storage Requests</Label>
                <Input
                  id="storage-requests"
                  placeholder="e.g., 10Gi, 100Gi"
                  value={formData.storageRequests}
                  onChange={(e) => setFormData({ ...formData, storageRequests: e.target.value })}
                />
              </div>
            </div>

            <div className="space-y-4">
              <h3 className="text-sm font-medium">Network Policies</h3>

              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="allow-ingress">Allow Ingress</Label>
                    <p className="text-muted-foreground text-xs">
                      Allow incoming traffic to this namespace
                    </p>
                  </div>
                  <Switch
                    id="allow-ingress"
                    checked={formData.allowIngress}
                    onCheckedChange={(checked) =>
                      setFormData({ ...formData, allowIngress: checked })
                    }
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="allow-egress">Allow Egress</Label>
                    <p className="text-muted-foreground text-xs">
                      Allow outgoing traffic from this namespace
                    </p>
                  </div>
                  <Switch
                    id="allow-egress"
                    checked={formData.allowEgress}
                    onCheckedChange={(checked) =>
                      setFormData({ ...formData, allowEgress: checked })
                    }
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="isolate-namespace">Isolate Namespace</Label>
                    <p className="text-muted-foreground text-xs">
                      Prevent cross-namespace communication
                    </p>
                  </div>
                  <Switch
                    id="isolate-namespace"
                    checked={formData.isolateNamespace}
                    onCheckedChange={(checked) =>
                      setFormData({ ...formData, isolateNamespace: checked })
                    }
                  />
                </div>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting || isTransitional}>
              {isSubmitting ? 'Updating...' : 'Update Environment'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
