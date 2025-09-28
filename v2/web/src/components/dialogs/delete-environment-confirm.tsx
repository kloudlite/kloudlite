'use client'

import { useState } from 'react'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle, Loader2 } from 'lucide-react'
import { deleteEnvironment } from '@/app/actions/environment.actions'

interface DeleteEnvironmentConfirmProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  environmentName: string
  onSuccess?: () => void
  currentUser?: string
}

export function DeleteEnvironmentConfirm({
  open,
  onOpenChange,
  environmentName,
  onSuccess,
  currentUser = 'test-user',
}: DeleteEnvironmentConfirmProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleDelete = async () => {
    setError(null)
    setLoading(true)

    try {
      const result = await deleteEnvironment(environmentName, currentUser)
      if (result.success) {
        onOpenChange(false)
        if (onSuccess) {
          onSuccess()
        }
      } else {
        setError(result.error || 'Failed to delete environment. Please try again.')
      }
    } catch (err: any) {
      console.error('Failed to delete environment:', err)
      setError(err.message || 'Failed to delete environment. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Environment</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete the environment "{environmentName}"?
            This action cannot be undone and will remove all associated resources.
          </AlertDialogDescription>
        </AlertDialogHeader>

        {error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <AlertDialogFooter>
          <AlertDialogCancel disabled={loading}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={loading}
            className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
          >
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {loading ? 'Deleting...' : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}