'use client'

import { useEffect, useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Upload, Trash2, AlertCircle, Loader2 } from 'lucide-react'
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
import { getConfig, setConfig, deleteConfig } from '@/app/actions/environment-config'
import { AddConfigSheet } from '@/components/add-config-sheet'
import { EditConfigSheet } from '@/components/edit-config-sheet'
import { toast } from 'sonner'

interface ConfigMapsListProps {
  environmentId: string
}

type ConfigEntry = {
  key: string
  value: string
}

export function ConfigMapsList({ environmentId }: ConfigMapsListProps) {
  const router = useRouter()
  const [configs, setConfigs] = useState<ConfigEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  // Delete dialog state
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [selectedKey, setSelectedKey] = useState<string>('')

  useEffect(() => {
    loadConfigs()
  }, [environmentId])

  const loadConfigs = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await getConfig(environmentId)
      const entries = Object.entries(response.data || {}).map(([key, value]) => ({
        key,
        value,
      }))
      setConfigs(entries)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load configs')
      setConfigs([])
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
        const configsMap = Object.fromEntries(configs.map(c => [c.key, c.value]))
        delete configsMap[selectedKey]

        if (Object.keys(configsMap).length === 0) {
          await deleteConfig(environmentId)
        } else {
          await setConfig(environmentId, configsMap)
        }

        toast.success('Config deleted successfully')
        setShowDeleteDialog(false)
        router.refresh()
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to delete config')
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
          <span className="font-medium">Error loading configs</span>
        </div>
        <p className="mt-2 text-sm text-red-700">{error}</p>
        <Button onClick={loadConfigs} variant="outline" size="sm" className="mt-3">
          Retry
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">Config Maps</h3>
          <p className="text-sm text-gray-500">Environment configuration variables</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm">
            <Upload className="h-4 w-4 mr-2" />
            Import
          </Button>
          <AddConfigSheet environmentId={environmentId} />
        </div>
      </div>

      {configs.length === 0 ? (
        <div className="text-center py-12 bg-gray-50 rounded-lg border border-gray-200">
          <p className="text-gray-500">No configuration variables</p>
          <AddConfigSheet environmentId={environmentId} />
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
              {configs.map((config) => (
                <tr key={config.key} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">{config.key}</td>
                  <td className="px-6 py-4 text-sm text-gray-600 font-mono max-w-md truncate">{config.value}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-1">
                    <EditConfigSheet
                      environmentId={environmentId}
                      configKey={config.key}
                      configValue={config.value}
                    />
                    <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700" onClick={() => handleDeleteClick(config.key)}>
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
            <AlertDialogTitle>Delete Config</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the config key <span className="font-mono font-semibold">{selectedKey}</span>?
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
