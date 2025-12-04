'use client'

import { useState, useEffect } from 'react'
import { Cpu, MemoryStick } from 'lucide-react'
import { getWorkspaceMetrics } from '@/app/actions/workspace.actions'
import type { WorkspaceMetrics } from '@kloudlite/types'

interface WorkspaceMetricsProps {
  workspaceName: string
  namespace: string
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

export function WorkspaceMetrics({ workspaceName, namespace }: WorkspaceMetricsProps) {
  const [metrics, setMetrics] = useState<WorkspaceMetrics | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        const result = await getWorkspaceMetrics(workspaceName, namespace)
        if (result.success && result.data) {
          setMetrics(result.data)
          setError(null)
        } else {
          setError(result.error || 'Failed to load metrics')
        }
      } catch (err) {
        console.error('Failed to fetch metrics:', err)
        setError('Failed to load metrics')
      }
    }

    // Initial fetch
    fetchMetrics()

    // Poll every 3 seconds
    const intervalId = setInterval(fetchMetrics, 3000)

    return () => {
      clearInterval(intervalId)
    }
  }, [workspaceName, namespace])

  if (error) {
    return (
      <div className="bg-card rounded-lg border p-6">
        <h3 className="mb-4 text-sm font-medium">Resource Usage</h3>
        <p className="text-muted-foreground text-sm">Metrics unavailable</p>
        <p className="text-muted-foreground mt-1 text-xs">
          Real-time metrics will be available once the backend endpoint is configured
        </p>
      </div>
    )
  }

  if (!metrics) {
    return (
      <div className="bg-card rounded-lg border p-6">
        <h3 className="mb-4 text-sm font-medium">Resource Usage</h3>
        <div className="space-y-3">
          <div className="flex items-center justify-between animate-pulse">
            <div className="flex items-center gap-2">
              <div className="bg-muted h-4 w-4 rounded"></div>
              <div className="bg-muted h-5 w-8 rounded"></div>
            </div>
            <div className="bg-muted h-5 w-20 rounded"></div>
          </div>
          <div className="flex items-center justify-between animate-pulse">
            <div className="flex items-center gap-2">
              <div className="bg-muted h-4 w-4 rounded"></div>
              <div className="bg-muted h-5 w-14 rounded"></div>
            </div>
            <div className="bg-muted h-5 w-16 rounded"></div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-card rounded-lg border p-6">
      <h3 className="mb-4 text-sm font-medium">Resource Usage</h3>
      <div className="space-y-3">
        {/* CPU */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Cpu className="text-muted-foreground h-4 w-4" />
            <span className="text-sm font-medium">CPU</span>
          </div>
          <span className="font-mono text-sm">{(metrics.cpu.usage / 1000).toFixed(3)} vCPU</span>
        </div>

        {/* Memory */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <MemoryStick className="text-muted-foreground h-4 w-4" />
            <span className="text-sm font-medium">Memory</span>
          </div>
          <span className="font-mono text-sm">{formatBytes(metrics.memory.usage)}</span>
        </div>
      </div>
    </div>
  )
}
