'use client'

import { useState, useEffect } from 'react'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
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
  currentUser: _currentUser = 'test-user',
}: DeleteEnvironmentConfirmProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Clear error when dialog opens/closes
  useEffect(() => {
    if (open) {
      setError(null)
    }
  }, [open])

  const parseErrorMessage = (errorString: string): string => {
    // Try to extract meaningful error message from API response
    if (errorString.includes('API Error:')) {
      // Try to parse JSON from the error string
      try {
        const jsonMatch = errorString.match(/\{[^}]+\}/)
        if (jsonMatch) {
          const errorData = JSON.parse(jsonMatch[0])
          // Check for various error fields
          if (errorData.error) {
            return errorData.error
          }
          if (errorData.details) {
            return errorData.details
          }
        }
      } catch (_e) {
        // If JSON parsing fails, try to extract key messages
      }
    }

    // Check for common error patterns
    if (errorString.includes('Cannot delete an activated environment')) {
      return 'Cannot delete an activated environment. Please deactivate it first.'
    }
    if (errorString.includes('Deactivate the environment first')) {
      return 'Please deactivate the environment before deleting it.'
    }
    if (errorString.includes('not found')) {
      return 'Environment not found.'
    }
    if (errorString.includes('permission denied')) {
      return 'You do not have permission to delete this environment.'
    }

    // If no specific pattern matched, return the original or a generic message
    return errorString || 'Failed to delete environment. Please try again.'
  }

  const handleDelete = async () => {
    setError(null)
    setLoading(true)

    try {
      const result = await deleteEnvironment(environmentName)
      if (result.success) {
        onOpenChange(false)
        if (onSuccess) {
          onSuccess()
        }
      } else {
        setError(parseErrorMessage(result.error || ''))
      }
    } catch (err) {
      console.error('Failed to delete environment:', err)
      const error = err instanceof Error ? err : new Error('Unknown error')
      setError(parseErrorMessage(error.message))
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
            Are you sure you want to delete the environment &quot;{environmentName}&quot;?
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
          <Button
            onClick={(e) => {
              e.preventDefault()
              handleDelete()
            }}
            disabled={loading}
            variant="destructive"
            className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
          >
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {loading ? 'Deleting...' : 'Delete'}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}