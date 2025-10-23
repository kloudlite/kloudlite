'use client'

import { useState, useEffect } from 'react'
import { Cpu, MemoryStick } from 'lucide-react'
import { getNodeMetrics } from '@/app/actions/workspace.actions'

interface NodeMetrics {
  cpu: {
    usage: number
    capacity: number
    allocatable: number
  }
  memory: {
    usage: number
    capacity: number
    allocatable: number
  }
  timestamp: string
}

interface WorkMachineMetricsProps {
  nodeName?: string
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

export function WorkMachineMetrics({ nodeName = 'master' }: WorkMachineMetricsProps) {
  const [metrics, setMetrics] = useState<NodeMetrics | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        const result = await getNodeMetrics(nodeName)
        if (result.success && result.data) {
          setMetrics(result.data)
          setError(null)
        } else {
          setError(result.error || 'Failed to load metrics')
        }
      } catch (err) {
        console.error('Failed to fetch node metrics:', err)
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
  }, [nodeName])

  const getUsageColor = (value: number) => {
    if (value < 50) return 'bg-green-500'
    if (value < 80) return 'bg-yellow-500'
    return 'bg-red-500'
  }

  const getUsageTextColor = (value: number) => {
    if (value < 50) return 'text-green-600'
    if (value < 80) return 'text-yellow-600'
    return 'text-red-600'
  }

  // Calculate percentages
  const cpuPercent = metrics?.cpu.capacity
    ? Math.round((metrics.cpu.usage / metrics.cpu.capacity) * 100)
    : 0

  const memoryPercent = metrics?.memory.capacity
    ? Math.round((metrics.memory.usage / metrics.memory.capacity) * 100)
    : 0

  if (error) {
    return (
      <div className="grid gap-4 md:grid-cols-3">
        <div className="bg-card rounded-lg border p-6">
          <p className="text-sm text-muted-foreground">Metrics unavailable</p>
          <p className="text-xs text-muted-foreground mt-1">Real-time metrics will be available once the node is ready</p>
        </div>
      </div>
    )
  }

  if (!metrics) {
    return (
      <div className="grid gap-4 md:grid-cols-3">
        <div className="bg-card rounded-lg border p-6">
          <div className="animate-pulse">
            <div className="h-4 bg-muted rounded w-1/3 mb-2"></div>
            <div className="h-2 bg-muted rounded"></div>
          </div>
        </div>
        <div className="bg-card rounded-lg border p-6">
          <div className="animate-pulse">
            <div className="h-4 bg-muted rounded w-1/3 mb-2"></div>
            <div className="h-2 bg-muted rounded"></div>
          </div>
        </div>
        <div className="bg-card rounded-lg border p-6">
          <div className="animate-pulse">
            <div className="h-4 bg-muted rounded w-1/3 mb-2"></div>
            <div className="h-2 bg-muted rounded"></div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* CPU Usage */}
      <div className="bg-card rounded-lg border p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <div className="p-2 bg-blue-50 dark:bg-blue-900/30 rounded-lg">
              <Cpu className="h-5 w-5 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <h3 className="text-sm font-semibold">CPU Usage</h3>
              <p className="text-xs text-muted-foreground">Processing power</p>
            </div>
          </div>
          <span className={`text-2xl font-medium ${getUsageTextColor(cpuPercent)}`}>
            {cpuPercent}%
          </span>
        </div>
        <div className="space-y-2">
          <div className="h-2 bg-muted rounded-full overflow-hidden">
            <div
              className={`h-full ${getUsageColor(cpuPercent)} transition-all duration-300`}
              style={{ width: `${cpuPercent}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>0%</span>
            <span>100%</span>
          </div>
        </div>
        <div className="mt-3 text-xs">
          {(metrics.cpu.usage / 1000).toFixed(2)} / {(metrics.cpu.capacity / 1000).toFixed(2)} vCPU
        </div>
      </div>

      {/* Memory Usage */}
      <div className="bg-card rounded-lg border p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <div className="p-2 bg-purple-50 dark:bg-purple-900/30 rounded-lg">
              <MemoryStick className="h-5 w-5 text-purple-600 dark:text-purple-400" />
            </div>
            <div>
              <h3 className="text-sm font-semibold">Memory Usage</h3>
              <p className="text-xs text-muted-foreground">RAM utilization</p>
            </div>
          </div>
          <span className={`text-2xl font-medium ${getUsageTextColor(memoryPercent)}`}>
            {memoryPercent}%
          </span>
        </div>
        <div className="space-y-2">
          <div className="h-2 bg-muted rounded-full overflow-hidden">
            <div
              className={`h-full ${getUsageColor(memoryPercent)} transition-all duration-300`}
              style={{ width: `${memoryPercent}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>0%</span>
            <span>100%</span>
          </div>
        </div>
        <div className="mt-3 text-xs">
          {formatBytes(metrics.memory.usage)} / {formatBytes(metrics.memory.capacity)}
        </div>
      </div>
    </div>
  )
}
