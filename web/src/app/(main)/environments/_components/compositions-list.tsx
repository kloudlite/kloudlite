'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { FileCode, Clock, Loader2 } from 'lucide-react'
import type { Composition } from '@/types/composition'
import { CreateCompositionSheet } from './create-composition-sheet'
import { CompositionRowActions } from './composition-row-actions'

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

export function CompositionsList({ compositions, namespace, user }: CompositionsListProps) {
  const router = useRouter()

  // Poll for updates when resources are in transitional states
  useEffect(() => {
    const transitionalStates = ['deploying', 'pending', 'stopping', 'starting']
    const hasTransitionalState = compositions.some(
      comp =>
        comp.metadata.deletionTimestamp ||
        transitionalStates.includes(comp.status?.state || '') ||
        !comp.status?.state // Also poll if no status yet (newly created)
    )

    if (!hasTransitionalState) {
      return
    }

    const interval = setInterval(() => {
      router.refresh()
    }, 3000) // Poll every 3 seconds

    return () => clearInterval(interval)
  }, [compositions, router])

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">Compositions</h3>
          <p className="text-sm text-muted-foreground mt-1">Container stacks managed with Docker Compose</p>
        </div>
        <CreateCompositionSheet namespace={namespace} user={user} />
      </div>

      <div className="bg-card rounded-lg border">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Composition
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Last Deployed
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {compositions.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-6 py-8 text-center text-sm text-muted-foreground">
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
                  const isDeleting = state === 'deleting'

                  // Determine status color based on state
                  let statusColor = 'bg-secondary text-secondary-foreground'
                  if (state === 'running') statusColor = 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                  else if (state === 'deploying') statusColor = 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
                  else if (state === 'degraded') statusColor = 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
                  else if (state === 'failed') statusColor = 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                  else if (state === 'stopped') statusColor = 'bg-secondary text-secondary-foreground'
                  else if (state === 'deleting') statusColor = 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'

                  return (
                    <tr key={composition.metadata.uid || composition.metadata.name} className="hover:bg-muted/50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <FileCode className="h-5 w-5 text-muted-foreground mr-3" />
                          <div>
                            <div className="text-sm font-medium">{composition.spec.displayName}</div>
                            {composition.spec.description && (
                              <div className="text-xs text-muted-foreground mt-0.5">{composition.spec.description}</div>
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
                        <div className="flex items-center gap-1 text-sm text-muted-foreground">
                          <Clock className="h-3 w-3" />
                          {lastDeployed}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <CompositionRowActions
                          composition={composition}
                          namespace={namespace}
                          user={user}
                          isDeleting={isDeleting}
                        />
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
