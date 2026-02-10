'use client'

import { useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { Button, AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@kloudlite/ui'
import { Trash2, Loader2, AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'

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
  const [phase, setPhase] = useState<'idle' | 'uninstalling' | 'deleting'>('idle')
  const [stepInfo, setStepInfo] = useState<{ current: number; total: number; description: string } | null>(null)

  const isManaged = cloudProvider === 'oci' && hasSecretKey

  const pollJobStatus = useCallback(async () => {
    try {
      const response = await fetch(`/api/installations/${installationId}/job-status`)
      if (!response.ok) return null
      const data = await response.json()

      // Capture step progress
      if (data.currentStep != null && data.totalSteps != null) {
        setStepInfo({
          current: data.currentStep,
          total: data.totalSteps,
          description: data.stepDescription || '',
        })
      }

      return data.status as string
    } catch {
      return null
    }
  }, [installationId])

  // Poll uninstall job while uninstalling
  useEffect(() => {
    if (phase !== 'uninstalling') return

    const poll = async () => {
      const status = await pollJobStatus()
      if (!status) return

      if (status === 'succeeded') {
        setPhase('deleting')
        await doDelete()
      } else if (status === 'failed') {
        toast.error('Uninstall job failed. Please try again.')
        setDeleting(false)
        setPhase('idle')
        setStepInfo(null)
      }
    }

    poll()
    const interval = setInterval(poll, 5000)
    return () => clearInterval(interval)
  }, [phase, installationId, pollJobStatus])

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
      setPhase('idle')
      router.push('/installations')
      router.refresh()
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to delete installation')
      toast.error(error.message)
      setDeleting(false)
      setPhase('idle')
    }
  }

  const handleDelete = async () => {
    setDeleting(true)

    if (isManaged) {
      // Trigger uninstall job first, then delete on success
      setPhase('uninstalling')
      try {
        const response = await fetch(
          `/api/installations/${installationId}/trigger-managed-uninstall`,
          { method: 'POST' },
        )
        if (!response.ok) {
          const data = await response.json()
          throw new Error(data.error || 'Failed to trigger uninstall')
        }
        toast.success('Uninstall job started — infrastructure will be torn down before deletion')
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Failed to trigger uninstall'
        toast.error(message)
        setDeleting(false)
        setPhase('idle')
      }
    } else {
      // Direct delete for BYOC installations
      await doDelete()
    }
  }

  const buttonLabel = 'Delete Installation'
  const actionLabel = 'Delete'

  const progressLabel =
    phase === 'uninstalling'
      ? stepInfo && stepInfo.current > 0
        ? `Uninstalling... (Step ${stepInfo.current}/${stepInfo.total})`
        : 'Uninstalling infrastructure...'
      : phase === 'deleting'
        ? 'Deleting...'
        : 'Deleting...'

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        {variant === 'button' ? (
          <Button variant="destructive" size="default" disabled={deleting}>
            {deleting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {progressLabel}
              </>
            ) : (
              <>
                <Trash2 className="mr-2 h-4 w-4" />
                {buttonLabel}
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
                {progressLabel}
              </>
            ) : (
              actionLabel
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
