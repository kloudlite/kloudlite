'use client'

import { useState, useEffect, useRef } from 'react'
import { Cpu, MemoryStick, Zap } from 'lucide-react'

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

interface GPUMetrics {
  detected: boolean
  model?: string
  driverVersion?: string
  count?: number
  memoryTotal?: number
  memoryUsed?: number
  memoryFree?: number
  utilizationGpu?: number
  utilizationMemory?: number
  temperature?: number
  powerDraw?: number
  powerLimit?: number
}

interface WorkMachineMetricsProps {
  workMachineName: string
  machineState?: string
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

export function WorkMachineMetrics({ workMachineName, machineState }: WorkMachineMetricsProps) {
  const [metrics, setMetrics] = useState<NodeMetrics | null>(null)
  const [gpuMetrics, setGpuMetrics] = useState<GPUMetrics | null>(null)
  const [error, setError] = useState<string | null>(null)
  const eventSourceRef = useRef<EventSource | null>(null)

  useEffect(() => {
    // Don't fetch metrics when machine is stopped or no name provided
    if (machineState === 'stopped' || !workMachineName) {
      return
    }

    // Close existing connection if any
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    // Create SSE connection
    const eventSource = new EventSource(
      `/api/v1/work-machines/${workMachineName}/metrics-stream`
    )
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      console.log('SSE connection established for metrics')
      setError(null)
    }

    eventSource.addEventListener('metrics', (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.nodeMetrics) {
          setMetrics(data.nodeMetrics)
        }
        if (data.gpuMetrics) {
          setGpuMetrics(data.gpuMetrics)
        }
      } catch (err) {
        console.error('Failed to parse metrics event:', err)
      }
    })

    eventSource.onerror = (err) => {
      console.error('SSE connection error:', err)
      setError('Failed to load metrics - connection error')
      eventSource.close()
    }

    return () => {
      eventSource.close()
      eventSourceRef.current = null
    }
  }, [workMachineName, machineState])

  // Don't show metrics when machine is stopped
  if (machineState === 'stopped') {
    return null
  }

  const getUsageColor = (value: number) => {
    if (value < 50) return 'bg-success'
    if (value < 80) return 'bg-warning'
    return 'bg-destructive'
  }

  const getUsageTextColor = (value: number) => {
    if (value < 50) return 'text-success'
    if (value < 80) return 'text-warning'
    return 'text-destructive'
  }

  // Calculate percentages
  const cpuPercent = metrics?.cpu.capacity
    ? Math.round((metrics.cpu.usage / metrics.cpu.capacity) * 100)
    : 0

  const memoryPercent = metrics?.memory.capacity
    ? Math.round((metrics.memory.usage / metrics.memory.capacity) * 100)
    : 0

  const gpuUtilPercent = gpuMetrics?.utilizationGpu ?? 0
  const gpuMemoryPercent = gpuMetrics?.utilizationMemory ?? 0

  if (error) {
    return (
      <div className="grid gap-4 md:grid-cols-3">
        <div className="bg-card rounded-lg border p-6">
          <p className="text-muted-foreground text-sm">Metrics unavailable</p>
          <p className="text-muted-foreground mt-1 text-xs">
            Real-time metrics will be available once the node is ready
          </p>
        </div>
      </div>
    )
  }

  if (!metrics) {
    const SkeletonCard = () => (
      <div className="bg-card rounded-lg border p-6 animate-pulse">
        <div className="mb-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="bg-muted h-9 w-9 rounded-lg"></div>
            <div>
              <div className="bg-muted h-4 w-20 rounded mb-1"></div>
              <div className="bg-muted h-3 w-24 rounded"></div>
            </div>
          </div>
          <div className="bg-muted h-8 w-12 rounded"></div>
        </div>
        <div className="space-y-2">
          <div className="bg-muted h-2 rounded-full"></div>
          <div className="flex justify-between">
            <div className="bg-muted h-3 w-6 rounded"></div>
            <div className="bg-muted h-3 w-8 rounded"></div>
          </div>
        </div>
        <div className="mt-3">
          <div className="bg-muted h-3 w-28 rounded"></div>
        </div>
      </div>
    )

    return (
      <div className="grid gap-4 md:grid-cols-2">
        <SkeletonCard />
        <SkeletonCard />
      </div>
    )
  }

  const hasGPU = gpuMetrics?.detected ?? false
  const gridCols = hasGPU ? 'md:grid-cols-3' : 'md:grid-cols-2'

  return (
    <div className={`grid gap-4 ${gridCols}`}>
      {/* CPU Usage */}
      <div className="bg-card rounded-lg border p-6">
        <div className="mb-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="bg-info/10 rounded-lg p-2">
              <Cpu className="text-info h-5 w-5" />
            </div>
            <div>
              <h3 className="text-sm font-semibold">CPU Usage</h3>
              <p className="text-muted-foreground text-xs">Processing power</p>
            </div>
          </div>
          <span className={`text-2xl font-medium ${getUsageTextColor(cpuPercent)}`}>
            {cpuPercent}%
          </span>
        </div>
        <div className="space-y-2">
          <div className="bg-muted h-2 overflow-hidden rounded-full">
            <div
              className={`h-full ${getUsageColor(cpuPercent)} transition-all duration-300`}
              style={{ width: `${cpuPercent}%` }}
            />
          </div>
          <div className="text-muted-foreground flex justify-between text-xs">
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
        <div className="mb-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="bg-accent/10 rounded-lg p-2">
              <MemoryStick className="text-accent-foreground h-5 w-5" />
            </div>
            <div>
              <h3 className="text-sm font-semibold">Memory Usage</h3>
              <p className="text-muted-foreground text-xs">RAM utilization</p>
            </div>
          </div>
          <span className={`text-2xl font-medium ${getUsageTextColor(memoryPercent)}`}>
            {memoryPercent}%
          </span>
        </div>
        <div className="space-y-2">
          <div className="bg-muted h-2 overflow-hidden rounded-full">
            <div
              className={`h-full ${getUsageColor(memoryPercent)} transition-all duration-300`}
              style={{ width: `${memoryPercent}%` }}
            />
          </div>
          <div className="text-muted-foreground flex justify-between text-xs">
            <span>0%</span>
            <span>100%</span>
          </div>
        </div>
        <div className="mt-3 text-xs">
          {formatBytes(metrics.memory.usage)} / {formatBytes(metrics.memory.capacity)}
        </div>
      </div>

      {/* GPU Usage - only show if GPU is detected */}
      {hasGPU && (
        <div className="bg-card rounded-lg border p-6">
          <div className="mb-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="bg-warning/10 rounded-lg p-2">
                <Zap className="text-warning h-5 w-5" />
              </div>
              <div>
                <h3 className="text-sm font-semibold">GPU Usage</h3>
                <p className="text-muted-foreground text-xs">{gpuMetrics?.model || 'Graphics processor'}</p>
              </div>
            </div>
            <span className={`text-2xl font-medium ${getUsageTextColor(gpuUtilPercent)}`}>
              {gpuUtilPercent}%
            </span>
          </div>
          <div className="space-y-3">
            {/* GPU Utilization */}
            <div className="space-y-1">
              <div className="text-muted-foreground flex justify-between text-xs">
                <span>Compute</span>
                <span>{gpuUtilPercent}%</span>
              </div>
              <div className="bg-muted h-2 overflow-hidden rounded-full">
                <div
                  className={`h-full ${getUsageColor(gpuUtilPercent)} transition-all duration-300`}
                  style={{ width: `${gpuUtilPercent}%` }}
                />
              </div>
            </div>
            {/* GPU Memory */}
            <div className="space-y-1">
              <div className="text-muted-foreground flex justify-between text-xs">
                <span>Memory</span>
                <span>{gpuMemoryPercent}%</span>
              </div>
              <div className="bg-muted h-2 overflow-hidden rounded-full">
                <div
                  className={`h-full ${getUsageColor(gpuMemoryPercent)} transition-all duration-300`}
                  style={{ width: `${gpuMemoryPercent}%` }}
                />
              </div>
            </div>
          </div>
          <div className="mt-3 grid grid-cols-2 gap-2 text-xs">
            {gpuMetrics?.memoryUsed && gpuMetrics?.memoryTotal && (
              <div>
                <span className="text-muted-foreground">VRAM: </span>
                {(gpuMetrics.memoryUsed / 1024).toFixed(1)} / {(gpuMetrics.memoryTotal / 1024).toFixed(1)} GB
              </div>
            )}
            {gpuMetrics?.temperature && (
              <div>
                <span className="text-muted-foreground">Temp: </span>
                {gpuMetrics.temperature}°C
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
