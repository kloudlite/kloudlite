'use client'

import { useState, useEffect } from 'react'
import { Cpu, MemoryStick } from 'lucide-react'
import { getWorkspaceMetrics } from '@/app/actions/workspace-query.actions'

interface WorkspaceMetricsProps {
  workspaceName: string
  namespace: string
  isRunning?: boolean
}

interface MetricsData {
  cpu: { usage: number }
  memory: { usage: number; usagePercent: number }
  timestamp: string
}

function MetricsSkeletonCard() {
  return (
    <div className="bg-card animate-pulse rounded-lg border p-6">
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="bg-muted h-9 w-9 rounded-lg"></div>
          <div className="space-y-1">
            <div className="bg-muted h-4 w-[70px] rounded"></div>
            <div className="bg-muted h-3 w-[100px] rounded"></div>
          </div>
        </div>
        <div className="bg-muted h-8 w-20 rounded"></div>
      </div>
    </div>
  )
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

export function WorkspaceMetrics({ workspaceName, namespace, isRunning = false }: WorkspaceMetricsProps) {
  const [metrics, setMetrics] = useState<MetricsData | null>(null)
  const [fetchError, setFetchError] = useState<string | null>(null)

  useEffect(() => {
    // Don't fetch metrics if workspace is not running
    if (!isRunning) {
      return
    }

    const fetchMetrics = async () => {
      try {
        const result = await getWorkspaceMetrics(workspaceName, namespace)
        if (result.success && result.data) {
          setMetrics(result.data as MetricsData)
          setFetchError(null)
        } else {
          setFetchError(result.error || 'Failed to load metrics')
        }
      } catch (err) {
        console.error('Failed to fetch metrics:', err)
        setFetchError('Failed to load metrics')
      }
    }

    // Initial fetch
    fetchMetrics()

    // Poll every 5 seconds (reduced frequency)
    const intervalId = setInterval(fetchMetrics, 5000)

    return () => {
      clearInterval(intervalId)
    }
  }, [workspaceName, namespace, isRunning])

  const error = !isRunning ? 'Workspace not running' : fetchError

  if (error) {
    return (
      <div className="grid gap-4 md:grid-cols-2">
        <div className="bg-card rounded-lg border p-6">
          <p className="text-muted-foreground text-sm">Metrics unavailable</p>
          <p className="text-muted-foreground mt-1 text-xs">
            Real-time metrics will be available once the workspace is running
          </p>
        </div>
      </div>
    )
  }

  if (!metrics) {
    return (
      <div className="grid gap-4 md:grid-cols-2">
        <MetricsSkeletonCard />
        <MetricsSkeletonCard />
      </div>
    )
  }

  const cpuValue = (metrics.cpu.usage / 1000).toFixed(3)
  const memoryValue = formatBytes(metrics.memory.usage)

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* CPU Usage */}
      <div className="bg-card rounded-lg border p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="bg-info/10 rounded-lg p-2">
              <Cpu className="text-info h-5 w-5" />
            </div>
            <div>
              <h3 className="text-sm font-semibold">CPU Usage</h3>
              <p className="text-muted-foreground text-xs">Processing power</p>
            </div>
          </div>
          <span className="text-2xl font-medium tabular-nums">
            {cpuValue}
            <span className="text-muted-foreground ml-1 text-sm font-normal">vCPU</span>
          </span>
        </div>
      </div>

      {/* Memory Usage */}
      <div className="bg-card rounded-lg border p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="bg-accent/10 rounded-lg p-2">
              <MemoryStick className="text-accent-foreground h-5 w-5" />
            </div>
            <div>
              <h3 className="text-sm font-semibold">Memory Usage</h3>
              <p className="text-muted-foreground text-xs">RAM utilization</p>
            </div>
          </div>
          <span className="text-2xl font-medium tabular-nums">
            {memoryValue}
          </span>
        </div>
      </div>
    </div>
  )
}
