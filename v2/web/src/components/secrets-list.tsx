'use client'

import { useEffect, useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Lock, Upload, Trash2, AlertCircle, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
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
import { getSecret, deleteSecret } from '@/app/actions/environment-config'
import { AddSecretSheet } from '@/components/add-secret-sheet'
import { EditSecretSheet } from '@/components/edit-secret-sheet'
import { toast } from 'sonner'

interface SecretsListProps {
  environmentId: string
}

export function SecretsList({ environmentId }: SecretsListProps) {
  const router = useRouter()
  const [secretKeys, setSecretKeys] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  // Delete dialog state
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [selectedKey, setSelectedKey] = useState<string>('')

  useEffect(() => {
    loadSecrets()
  }, [environmentId])

  const loadSecrets = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await getSecret(environmentId)
      setSecretKeys(response.keys || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load secrets')
      setSecretKeys([])
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteClick = (key: string) => {
    setSelectedKey(key)
    setShowDeleteDialog(true)
  }

  const handleDelete = () => {
    startTransition(async () => {
      try {
        if (secretKeys.length === 1) {
          // Delete the entire secret resource
          await deleteSecret(environmentId)
          toast.success('Secret deleted successfully')
          setShowDeleteDialog(false)
          router.refresh()
        } else {
          // Individual secret deletion is not supported
          toast.error('Individual secret deletion is not supported. This will delete all secrets.')
          setShowDeleteDialog(false)
        }
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to delete secret')
      }
    })
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 border border-red-200 p-4">
        <div className="flex items-center gap-2 text-red-800">
          <AlertCircle className="h-5 w-5" />
          <span className="font-medium">Error loading secrets</span>
        </div>
        <p className="mt-2 text-sm text-red-700">{error}</p>
        <Button onClick={loadSecrets} variant="outline" size="sm" className="mt-3">
          Retry
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">Secrets</h3>
          <p className="text-sm text-gray-500">Encrypted sensitive information (values are hidden for security)</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm">
            <Upload className="h-4 w-4 mr-2" />
            Import
          </Button>
          <AddSecretSheet environmentId={environmentId} />
        </div>
      </div>

      {secretKeys.length === 0 ? (
        <div className="text-center py-12 bg-gray-50 rounded-lg border border-gray-200">
          <Lock className="h-12 w-12 mx-auto text-gray-400 mb-4" />
          <p className="text-gray-500">No secrets configured</p>
          <div className="mt-4">
            <AddSecretSheet environmentId={environmentId} />
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
          <table className="min-w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Key</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Value</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {secretKeys.map((key) => (
                <tr key={key} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                    <div className="flex items-center gap-2">
                      <Lock className="h-3 w-3 text-amber-500" />
                      {key}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600 font-mono">
                    <span className="text-gray-400">••••••••••••••••</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-1">
                    <EditSecretSheet
                      environmentId={environmentId}
                      secretKey={key}
                    />
                    <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700" onClick={() => handleDeleteClick(key)}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Delete Dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Secret</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the secret key <span className="font-mono font-semibold">{selectedKey}</span>?
              {secretKeys.length > 1 && (
                <span className="block mt-2 text-amber-600">
                  Note: Individual secret deletion is not supported. This will delete all secrets. Please recreate the others if needed.
                </span>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} disabled={isPending} className="bg-red-600 hover:bg-red-700">
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
