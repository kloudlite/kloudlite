'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button, AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@kloudlite/ui'
import { Trash2, Loader2, AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'
import { getErrorMessage } from '@/lib/errors'

interface DeleteInstallationButtonProps {
  installationId: string
  installationName?: string
  hasSecretKey?: boolean
  cloudProvider?: string
  variant?: 'icon' | 'button'
}

export function DeleteInstallationButton({
  installationId,
  installationName,
  hasSecretKey = false,
  cloudProvider,
  variant = 'icon',
}: DeleteInstallationButtonProps) {
  const router = useRouter()
  const [deleting, setDeleting] = useState(false)
  const [open, setOpen] = useState(false)

  const isManaged = cloudProvider === 'oci' && hasSecretKey

  const doDelete = async () => {
    try {
      const response = await fetch(`/api/installations/${installationId}/delete`, {
        method: 'DELETE',
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to delete installation')
      }

      toast.success('Installation deleted successfully')
      setOpen(false)
      setDeleting(false)
      router.push('/installations')
      router.refresh()
    } catch (err) {
      toast.error(getErrorMessage(err, 'Failed to delete installation'))
      setDeleting(false)
    }
  }

  const handleDelete = async () => {
    setDeleting(true)

    if (isManaged) {
      // Trigger uninstall job, then navigate to list where progress is shown.
      // The job-lock endpoint auto-deletes the record when uninstall succeeds.
      try {
        const response = await fetch(
          `/api/installations/${installationId}/trigger-managed-uninstall`,
          { method: 'POST' },
        )
        if (!response.ok) {
          const data = await response.json()
          throw new Error(data.error || 'Failed to trigger uninstall')
        }
        toast.success('Uninstall started — infrastructure is being torn down')
        setOpen(false)
        setDeleting(false)
        router.push('/installations')
        router.refresh()
      } catch (err) {
        toast.error(getErrorMessage(err, 'Failed to trigger uninstall'))
        setDeleting(false)
      }
    } else {
      // Direct delete for BYOC installations
      await doDelete()
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        {variant === 'button' ? (
          <Button variant="destructive" size="default" disabled={deleting}>
            {deleting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Deleting...
              </>
            ) : (
              <>
                <Trash2 className="mr-2 h-4 w-4" />
                Delete Installation
              </>
            )}
          </Button>
        ) : (
          <Button variant="ghost" size="sm" disabled={deleting}>
            {deleting ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Trash2 className="h-4 w-4" />
            )}
          </Button>
        )}
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <AlertTriangle className="text-destructive h-5 w-5" />
            Delete Installation
          </AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-3">
              <p>
                Are you sure you want to delete <strong>{installationName}</strong>? This
                action cannot be undone.
              </p>
              <div className="rounded-md border border-red-200 bg-red-50 p-3 dark:border-red-900 dark:bg-red-950">
                <p className="text-sm text-red-900 dark:text-red-200">
                  <strong>Warning:</strong> {isManaged
                    ? 'This will tear down all managed infrastructure (instance, load balancer, DNS, storage) and permanently delete the installation.'
                    : 'This will permanently delete the installation. This action cannot be undone.'}
                </p>
              </div>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={deleting}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={(e) => {
              e.preventDefault()
              handleDelete()
            }}
            disabled={deleting}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {deleting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {isManaged ? 'Starting uninstall...' : 'Deleting...'}
              </>
            ) : (
              'Delete'
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
