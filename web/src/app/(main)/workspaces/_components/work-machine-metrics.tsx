'use client'

import { Cpu, HardDrive, MemoryStick } from 'lucide-react'

interface WorkMachineMetricsProps {
  cpu: number
  memory: number
  disk: number
}

export function WorkMachineMetrics({ cpu, memory, disk }: WorkMachineMetricsProps) {
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

  return (
    <div className="grid gap-4 md:grid-cols-3">
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
          <span className={`text-2xl font-medium ${getUsageTextColor(cpu)}`}>
            {cpu}%
          </span>
        </div>
        <div className="space-y-2">
          <div className="h-2 bg-muted rounded-full overflow-hidden">
            <div
              className={`h-full ${getUsageColor(cpu)} transition-all duration-300`}
              style={{ width: `${cpu}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>0%</span>
            <span>100%</span>
          </div>
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
          <span className={`text-2xl font-medium ${getUsageTextColor(memory)}`}>
            {memory}%
          </span>
        </div>
        <div className="space-y-2">
          <div className="h-2 bg-muted rounded-full overflow-hidden">
            <div
              className={`h-full ${getUsageColor(memory)} transition-all duration-300`}
              style={{ width: `${memory}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>0%</span>
            <span>100%</span>
          </div>
        </div>
        <div className="mt-3 text-xs">
          {Math.round((memory / 100) * 16)} GB / 16 GB
        </div>
      </div>

      {/* Disk Usage */}
      <div className="bg-card rounded-lg border p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <div className="p-2 bg-orange-50 dark:bg-orange-900/30 rounded-lg">
              <HardDrive className="h-5 w-5 text-orange-600 dark:text-orange-400" />
            </div>
            <div>
              <h3 className="text-sm font-semibold">Disk Usage</h3>
              <p className="text-xs text-muted-foreground">Storage space</p>
            </div>
          </div>
          <span className={`text-2xl font-medium ${getUsageTextColor(disk)}`}>
            {disk}%
          </span>
        </div>
        <div className="space-y-2">
          <div className="h-2 bg-muted rounded-full overflow-hidden">
            <div
              className={`h-full ${getUsageColor(disk)} transition-all duration-300`}
              style={{ width: `${disk}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>0%</span>
            <span>100%</span>
          </div>
        </div>
        <div className="mt-3 text-xs">
          {Math.round((disk / 100) * 500)} GB / 500 GB
        </div>
      </div>
    </div>
  )
}