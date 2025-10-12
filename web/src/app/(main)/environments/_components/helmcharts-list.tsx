'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Package2, ExternalLink, Clock, Loader2 } from 'lucide-react'
import type { HelmChart, HelmChartState } from '@/types/helmchart'
import { CreateHelmChartSheet } from './create-helmchart-sheet'

interface HelmChartsListProps {
  helmCharts: HelmChart[]
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

function getHelmChartState(chart: HelmChart): HelmChartState {
  // Check if deleting
  if (chart.metadata.deletionTimestamp) {
    return 'deleting'
  }

  // Use the state from status if available
  if (chart.status?.state) {
    return chart.status.state as HelmChartState
  }

  // Default to pending if no clear status
  return 'pending'
}

export function HelmChartsList({ helmCharts, namespace, user }: HelmChartsListProps) {
  const router = useRouter()

  // Poll for updates when resources are in transitional states
  useEffect(() => {
    const transitionalStates: HelmChartState[] = ['installing', 'pending', 'upgrading', 'uninstalling']
    const hasTransitionalState = helmCharts.some(
      chart =>
        chart.metadata.deletionTimestamp ||
        transitionalStates.includes(getHelmChartState(chart))
    )

    if (!hasTransitionalState) {
      return
    }

    const interval = setInterval(() => {
      router.refresh()
    }, 3000) // Poll every 3 seconds

    return () => clearInterval(interval)
  }, [helmCharts, router])

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-medium">Helm Charts</h3>
          <p className="text-sm text-muted-foreground mt-1">Kubernetes applications deployed via Helm</p>
        </div>
        <CreateHelmChartSheet namespace={namespace} user={user} />
      </div>

      <div className="bg-card rounded-lg border">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Chart
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Version
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Repository
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Updated
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Target
                </th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {helmCharts.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-8 text-center text-sm text-muted-foreground">
                    No helm charts found. Add your first helm chart to get started.
                  </td>
                </tr>
              ) : (
                helmCharts.map((chart) => {
                  const state = getHelmChartState(chart)
                  const lastUpdated = formatTimeAgo(chart.metadata.creationTimestamp)
                  const isDeleting = !!chart.metadata.deletionTimestamp

                  // Determine status color based on state
                  let statusColor = 'bg-secondary text-secondary-foreground'
                  if (state === 'installed') statusColor = 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                  else if (state === 'installing' || state === 'upgrading') statusColor = 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
                  else if (state === 'failed') statusColor = 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                  else if (state === 'uninstalling' || state === 'deleting') statusColor = 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'

                  return (
                    <tr key={chart.metadata.uid || chart.metadata.name} className="hover:bg-muted/50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <Package2 className="h-5 w-5 text-muted-foreground mr-3" />
                          <div>
                            <div className="text-sm font-medium">{chart.spec.displayName}</div>
                            {chart.spec.description && (
                              <div className="text-xs text-muted-foreground mt-0.5">{chart.spec.description}</div>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm">{chart.spec.chart?.version || 'latest'}</span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColor}`}>
                          {(state === 'deploying' || isDeleting) && (
                            <Loader2 className="h-3 w-3 animate-spin" />
                          )}
                          {isDeleting ? 'deleting' : state}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {chart.spec.chart?.url ? (
                          <a
                            href={chart.spec.chart.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-500 dark:hover:text-blue-300 flex items-center gap-1"
                          >
                            <span className="truncate max-w-[200px]">{chart.spec.chart.url}</span>
                            <ExternalLink className="h-3 w-3" />
                          </a>
                        ) : (
                          <span className="text-sm text-muted-foreground">-</span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-1 text-sm text-muted-foreground">
                          <Clock className="h-3 w-3" />
                          {lastUpdated}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm">{chart.spec.targetNamespace || namespace}</span>
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
