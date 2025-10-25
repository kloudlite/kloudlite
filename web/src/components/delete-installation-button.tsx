'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Trash2, Loader2, AlertTriangle } from 'lucide-react'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { toast } from 'sonner'

interface DeleteInstallationButtonProps {
  installationId: string
  installationName?: string
  hasSecretKey?: boolean
  variant?: 'icon' | 'button'
}

export function DeleteInstallationButton({
  installationId,
  installationName,
  hasSecretKey = false,
  variant = 'icon',
}: DeleteInstallationButtonProps) {
  const router = useRouter()
  const [deleting, setDeleting] = useState(false)
  const [open, setOpen] = useState(false)

  const handleDelete = async () => {
    setDeleting(true)

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
      router.refresh()
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to delete installation')
      toast.error(error.message)
    } finally {
      setDeleting(false)
    }
  }

  const handleCancel = () => {
    setOpen(false)
  }

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        {variant === 'button' ? (
          <Button variant="destructive" size="default">
            <Trash2 className="mr-2 h-4 w-4" />
            Force Delete Installation
          </Button>
        ) : (
          <Button variant="ghost" size="sm">
            <Trash2 className="h-4 w-4" />
          </Button>
        )}
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-destructive" />
            Force Delete Installation
          </AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-3">
              <p>
                Are you sure you want to force delete <strong>{installationName}</strong>? This action
                cannot be undone.
              </p>
              {hasSecretKey && (
                <div className="rounded-md border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950 p-3">
                  <p className="text-sm text-red-900 dark:text-red-200">
                    <strong>Warning:</strong> This will immediately uninstall Kloudlite from your cluster. All data and configurations will be permanently removed.
                    It&apos;s recommended to uninstall from your installation&apos;s dashboard settings for a cleaner removal.
                  </p>
                </div>
              )}
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={deleting} onClick={handleCancel}>
            Cancel
          </AlertDialogCancel>
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
                Force Deleting...
              </>
            ) : (
              'Force Delete'
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
