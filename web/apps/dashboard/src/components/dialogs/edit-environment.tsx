'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@kloudlite/ui'
import { Button } from '@kloudlite/ui'
import { updateEnvironment } from '@/app/actions/environment.actions'
import { toast } from 'sonner'
import type { EnvironmentUIModel, Visibility } from '@kloudlite/types'
import { VisibilitySelector } from '@/components/visibility-selector'

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
  currentUser,
}: EditEnvironmentDialogProps) {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Prevent editing if environment is in a transitional state
  const isTransitional = ['deleting', 'activating', 'deactivating'].includes(environment.status)

  const [visibility, setVisibility] = useState<Visibility>(environment.spec?.visibility || 'private')
  const [sharedWith, setSharedWith] = useState<string[]>(environment.spec?.sharedWith || [])

  // Reset form when environment changes
  useEffect(() => {
    setVisibility(environment.spec?.visibility || 'private')
    setSharedWith(environment.spec?.sharedWith || [])
  }, [environment])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)

    try {
      const updateData = {
        spec: {
          targetNamespace: environment.targetNamespace,
          activated: environment.status === 'active',
          ownedBy: currentUser,
          visibility,
          sharedWith: visibility === 'shared' ? sharedWith : undefined,
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
      <DialogContent className="sm:max-w-[500px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Edit Environment</DialogTitle>
            <DialogDescription>
              Update sharing settings for {environment.name}
            </DialogDescription>
          </DialogHeader>

          <div className="py-4">
            <VisibilitySelector
              visibility={visibility}
              sharedWith={sharedWith}
              onVisibilityChange={setVisibility}
              onSharedWithChange={setSharedWith}
              disabled={isSubmitting}
            />
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
