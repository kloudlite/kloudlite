'use client'

import { useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import {
  Button,
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@kloudlite/ui'
import { Trash2, Loader2, AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'

interface ManagedUninstallButtonProps {
  installationId: string
  installationName?: string
}

export function ManagedUninstallButton({
  installationId,
  installationName,
}: ManagedUninstallButtonProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [uninstalling, setUninstalling] = useState(false)
  const [jobStatus, setJobStatus] = useState<string | null>(null)

  const pollJobStatus = useCallback(async () => {
    try {
      const response = await fetch(`/api/installations/${installationId}/job-status`)
      if (!response.ok) return null

      const data = await response.json()
      return data.status as string
    } catch {
      return null
    }
  }, [installationId])

  // Poll while uninstalling
  useEffect(() => {
    if (!uninstalling) return

    const poll = async () => {
      const status = await pollJobStatus()
      if (!status) return

      setJobStatus(status)

      if (status === 'succeeded') {
        // Auto-delete the installation record
        try {
          await fetch(`/api/installations/${installationId}/delete`, {
            method: 'DELETE',
          })
          toast.success('Installation uninstalled and deleted successfully')
          setUninstalling(false)
          setOpen(false)
          router.push('/installations')
          router.refresh()
        } catch {
          toast.error('Uninstall succeeded but failed to delete installation record')
          setUninstalling(false)
        }
      } else if (status === 'failed') {
        toast.error('Uninstall job failed. Please try again.')
        setUninstalling(false)
        setJobStatus(null)
      }
    }

    poll()
    const interval = setInterval(poll, 5000)
    return () => clearInterval(interval)
  }, [uninstalling, installationId, pollJobStatus, router])

  const handleUninstall = async () => {
    setUninstalling(true)
    setJobStatus('pending')

    try {
      const response = await fetch(
        `/api/installations/${installationId}/trigger-managed-uninstall`,
        { method: 'POST' },
      )

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Failed to trigger uninstall')
      }

      toast.success('Uninstall job started')
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to trigger uninstall'
      toast.error(message)
      setUninstalling(false)
      setJobStatus(null)
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        <Button variant="destructive" size="default" disabled={uninstalling}>
          {uninstalling ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              {jobStatus === 'running' ? 'Uninstalling...' : 'Starting uninstall...'}
            </>
          ) : (
            <>
              <Trash2 className="mr-2 h-4 w-4" />
              Uninstall
            </>
          )}
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <AlertTriangle className="text-destructive h-5 w-5" />
            Uninstall Kloudlite Cloud
          </AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-3">
              <p>
                Are you sure you want to uninstall <strong>{installationName}</strong>? This will
                tear down all infrastructure and permanently delete the installation.
              </p>
              <div className="rounded-md border border-red-200 bg-red-50 p-3 dark:border-red-900 dark:bg-red-950">
                <p className="text-sm text-red-900 dark:text-red-200">
                  <strong>Warning:</strong> This action cannot be undone. All data, configurations,
                  and cloud resources associated with this installation will be permanently removed.
                </p>
              </div>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={uninstalling}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={(e) => {
              e.preventDefault()
              handleUninstall()
            }}
            disabled={uninstalling}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {uninstalling ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Uninstalling...
              </>
            ) : (
              'Uninstall'
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
