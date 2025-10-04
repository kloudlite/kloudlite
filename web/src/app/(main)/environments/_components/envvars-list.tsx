'use client'

import { useEffect, useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Key, Trash2, AlertCircle, Loader2, Plus, Pencil } from 'lucide-react'
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
import { getEnvVars, deleteEnvVar } from '@/app/actions/environment-config'
import type { EnvVar } from '@/types/environment'
import { AddEnvVarSheet } from './add-envvar-sheet'
import { EditEnvVarSheet } from './edit-envvar-sheet'
import { toast } from 'sonner'

interface EnvVarsListProps {
  environmentId: string
}

export function EnvVarsList({ environmentId }: EnvVarsListProps) {
  const router = useRouter()
  const [envVars, setEnvVars] = useState<EnvVar[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  // Delete dialog state
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [selectedEnvVar, setSelectedEnvVar] = useState<EnvVar | null>(null)

  // Edit sheet state
  const [showEditSheet, setShowEditSheet] = useState(false)
  const [editingEnvVar, setEditingEnvVar] = useState<EnvVar | null>(null)

  useEffect(() => {
    loadEnvVars()
  }, [environmentId])

  const loadEnvVars = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await getEnvVars(environmentId)
      setEnvVars(response.envVars || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load environment variables')
      setEnvVars([])
    } finally {
      setLoading(false)
    }
  }

  const handleEditClick = (envVar: EnvVar) => {
    setEditingEnvVar(envVar)
    setShowEditSheet(true)
  }

  const handleDeleteClick = (envVar: EnvVar) => {
    setSelectedEnvVar(envVar)
    setShowDeleteDialog(true)
  }

  const handleDelete = () => {
    if (!selectedEnvVar) return

    startTransition(async () => {
      try {
        await deleteEnvVar(environmentId, selectedEnvVar.key)
        toast.success('Environment variable deleted successfully')
        setShowDeleteDialog(false)
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to delete environment variable')
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
          <span className="font-medium">Error loading envvars</span>
        </div>
        <p className="mt-2 text-sm text-red-700">{error}</p>
        <Button onClick={loadEnvVars} variant="outline" size="sm" className="mt-3">
          Retry
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">Envvars</h3>
          <p className="text-sm text-gray-500">Configuration and secret envvars</p>
        </div>
        <AddEnvVarSheet environmentId={environmentId} onSuccess={loadEnvVars} />
      </div>

      {envVars.length === 0 ? (
        <div className="text-center py-12 bg-gray-50 rounded-lg border border-gray-200">
          <Key className="h-12 w-12 mx-auto text-gray-400 mb-4" />
          <p className="text-gray-500">No envvars</p>
          <div className="mt-4">
            <AddEnvVarSheet environmentId={environmentId} onSuccess={loadEnvVars} />
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
          <table className="min-w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Key</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Value</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {envVars.map((envVar) => (
                <tr key={envVar.key} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    <div className="flex items-center gap-2">
                      <Key className="h-4 w-4 text-gray-400" />
                      {envVar.key}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 font-mono">
                    {envVar.type === 'secret' ? (
                      <span className="text-gray-400">••••••••</span>
                    ) : (
                      <span className="max-w-xs truncate block">{envVar.value}</span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {envVar.type === 'config' ? (
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                        Config
                      </span>
                    ) : (
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                        Secret
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-1">
                    <Button variant="ghost" size="sm" className="text-gray-600 hover:text-gray-700" onClick={() => handleEditClick(envVar)}>
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700" onClick={() => handleDeleteClick(envVar)}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Edit Sheet */}
      {editingEnvVar && (
        <EditEnvVarSheet
          environmentId={environmentId}
          envVar={editingEnvVar}
          open={showEditSheet}
          onOpenChange={setShowEditSheet}
          onSuccess={loadEnvVars}
        />
      )}

      {/* Delete Dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Envvar</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the envvar <span className="font-mono font-semibold">{selectedEnvVar?.key}</span>?
              This action cannot be undone.
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
