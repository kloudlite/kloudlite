'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { FileCode, Clock, Trash2, Loader2 } from 'lucide-react'
import type { Composition } from '@/types/composition'
import { CreateCompositionSheet } from './create-composition-sheet'
import { EditCompositionSheet } from './edit-composition-sheet'
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
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { deleteComposition } from '@/app/actions/composition.actions'
import { toast } from 'sonner'

interface CompositionsListProps {
  compositions: Composition[]
  namespace: string
  user: string
}

function formatTimeAgo(timestamp?: string): string {
  if (!timestamp) return 'Never'

  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / (1000 * 60))
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins} min${diffMins > 1 ? 's' : ''} ago`
  if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
  return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
}

function DeleteButton({ composition, namespace, user }: { composition: Composition; namespace: string; user: string }) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()

  const handleDelete = async () => {
    startTransition(async () => {
      const result = await deleteComposition(namespace, composition.metadata.name, user)
      if (result.success) {
        toast.success('Composition deleted successfully')
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to delete composition')
      }
    })
  }

  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button variant="ghost" size="sm" className="h-8 w-8 p-0 text-red-600 hover:text-red-700 hover:bg-red-50" disabled={isPending}>
          <Trash2 className="h-4 w-4" />
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Composition</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete <strong>{composition.spec.displayName}</strong>? This action cannot be undone and will remove all associated resources.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={handleDelete} className="bg-red-600 hover:bg-red-700">
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

export function CompositionsList({ compositions: initialCompositions, namespace, user }: CompositionsListProps) {
  const router = useRouter()
  const [compositions, setCompositions] = useState<Composition[]>(initialCompositions)

  // Poll for composition updates every 3 seconds
  useEffect(() => {
    const pollInterval = setInterval(() => {
      // Check if any composition is in a transitional state (deletionTimestamp set or deploying)
      const hasTransitionalComp = compositions.some(
        comp => comp.metadata.deletionTimestamp || comp.status?.state === 'deploying'
      )

      if (hasTransitionalComp) {
        router.refresh()
      }
    }, 3000)

    return () => clearInterval(pollInterval)
  }, [compositions, router])

  // Update local state when server data changes
  useEffect(() => {
    setCompositions(initialCompositions)
  }, [initialCompositions])

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">Compositions</h3>
          <p className="text-sm text-gray-500 mt-1">Container stacks managed with Docker Compose</p>
        </div>
        <CreateCompositionSheet namespace={namespace} user={user} />
      </div>

      <div className="bg-white rounded-lg border border-gray-200">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 bg-gray-50">
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Composition
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Last Deployed
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {compositions.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-6 py-8 text-center text-sm text-gray-500">
                    No compositions found. Create your first composition to get started.
                  </td>
                </tr>
              ) : (
                compositions.map((composition) => {
                  // Determine state: prioritize deletionTimestamp, then status.state
                  const state = composition.metadata.deletionTimestamp
                    ? 'deleting'
                    : (composition.status?.state || 'pending')
                  const lastDeployed = formatTimeAgo(composition.status?.lastDeployedTime)

                  // Determine status color based on state
                  let statusColor = 'bg-gray-100 text-gray-800'
                  if (state === 'running') statusColor = 'bg-green-100 text-green-800'
                  else if (state === 'deploying') statusColor = 'bg-blue-100 text-blue-800'
                  else if (state === 'degraded') statusColor = 'bg-yellow-100 text-yellow-800'
                  else if (state === 'failed') statusColor = 'bg-red-100 text-red-800'
                  else if (state === 'stopped') statusColor = 'bg-gray-100 text-gray-800'
                  else if (state === 'deleting') statusColor = 'bg-red-100 text-red-800'

                  return (
                    <tr key={composition.metadata.uid || composition.metadata.name} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <FileCode className="h-5 w-5 text-gray-400 mr-3" />
                          <div>
                            <div className="text-sm font-medium text-gray-900">{composition.spec.displayName}</div>
                            {composition.spec.description && (
                              <div className="text-xs text-gray-500 mt-0.5">{composition.spec.description}</div>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColor}`}>
                          {(state === 'deleting' || state === 'deploying') && (
                            <Loader2 className="h-3 w-3 animate-spin" />
                          )}
                          {state}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-1 text-sm text-gray-500">
                          <Clock className="h-3 w-3" />
                          {lastDeployed}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {state === 'deleting' ? (
                          <span className="text-xs text-gray-500">Deleting...</span>
                        ) : (
                          <div className="flex items-center gap-2">
                            <EditCompositionSheet composition={composition} namespace={namespace} user={user} />
                            <DeleteButton composition={composition} namespace={namespace} user={user} />
                          </div>
                        )}
                      </td>
                    </tr>
                  )
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
