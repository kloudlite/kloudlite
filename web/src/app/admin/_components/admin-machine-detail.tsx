'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import {
  ArrowLeft,
  Server,
  Play,
  Square,
  RefreshCw,
  Settings,
  Activity,
  Clock,
  Cpu,
  MemoryStick,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Terminal,
  FolderOpen,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { WorkMachineMetrics } from '../../(main)/workspaces/_components/work-machine-metrics'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface Workspace {
  id: string
  name: string
  status: 'running' | 'stopped'
  resources: { cpu: number; memory: number }
  lastAccess: string
}

interface Environment {
  id: string
  name: string
  status: 'running' | 'stopped'
  resources: { cpu: number; memory: number }
}

interface ActivityLog {
  timestamp: Date
  action: string
  user: string
}

interface Machine {
  id: string
  owner: string
  name: string
  status: 'active' | 'idle' | 'stopped'
  cpu: number
  memory: number
  disk: number
  uptime: string
  type: string
  lastActive: Date
  workspaces: Workspace[]
  environments: Environment[]
  activityLog: ActivityLog[]
}

interface AdminMachineDetailProps {
  machine: Machine
}

export function AdminMachineDetail({ machine }: AdminMachineDetailProps) {
  const router = useRouter()
  const [isRestarting, setIsRestarting] = useState(false)

  const handleStart = () => {
    // TODO: Implement machine start logic
  }

  const handleStop = () => {
    // TODO: Implement machine stop logic
  }

  const handleRestart = async () => {
    setIsRestarting(true)
    // Simulate restart
    await new Promise((resolve) => setTimeout(resolve, 3000))
    setIsRestarting(false)
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="text-success h-5 w-5" />
      case 'idle':
        return <AlertTriangle className="text-warning h-5 w-5" />
      case 'stopped':
        return <XCircle className="text-muted-foreground h-5 w-5" />
      default:
        return null
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="sm" onClick={() => router.push('/administration')}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Machines
          </Button>
        </div>
        <div className="flex items-center gap-2">
          {machine.status === 'stopped' ? (
            <Button onClick={handleStart} className="gap-2">
              <Play className="h-4 w-4" />
              Start Machine
            </Button>
          ) : (
            <>
              <Button
                onClick={handleRestart}
                variant="outline"
                className="gap-2"
                disabled={isRestarting}
              >
                <RefreshCw className={`h-4 w-4 ${isRestarting ? 'animate-spin' : ''}`} />
                Restart
              </Button>
              <Button onClick={handleStop} variant="destructive" className="gap-2">
                <Square className="h-4 w-4" />
                Stop Machine
              </Button>
            </>
          )}
          <Button variant="outline" className="gap-2">
            <Settings className="h-4 w-4" />
            Settings
          </Button>
        </div>
      </div>

      {/* Machine Info */}
      <div className="border-border bg-card rounded-lg border p-6">
        <div className="mb-6 flex items-start justify-between">
          <div className="flex items-start gap-4">
            <Server className="text-muted-foreground mt-1 h-10 w-10" />
            <div>
              <h1 className="flex items-center gap-2 text-2xl font-semibold">
                {machine.name}
                {getStatusIcon(machine.status)}
              </h1>
              <p className="text-muted-foreground mt-1 text-sm">Owner: {machine.owner}</p>
              <div className="text-muted-foreground mt-2 flex items-center gap-4 text-sm">
                <span className="flex items-center gap-1">
                  <Activity className="h-4 w-4" />
                  Type: {machine.type}
                </span>
                <span className="flex items-center gap-1">
                  <Clock className="h-4 w-4" />
                  Uptime: {machine.uptime}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Metrics */}
        <WorkMachineMetrics />
      </div>

      {/* Tabs */}
      <Tabs defaultValue="workspaces" className="space-y-4">
        <TabsList>
          <TabsTrigger value="workspaces">Workspaces ({machine.workspaces.length})</TabsTrigger>
          <TabsTrigger value="environments">
            Environments ({machine.environments.length})
          </TabsTrigger>
          <TabsTrigger value="activity">Activity Log</TabsTrigger>
        </TabsList>

        <TabsContent value="workspaces">
          <div className="border-border bg-card rounded-lg border">
            <div className="border-border border-b px-6 py-4">
              <h2 className="text-lg font-medium">Active Workspaces</h2>
            </div>
            <div className="divide-border divide-y">
              {machine.workspaces.map((workspace) => (
                <div key={workspace.id} className="flex items-center justify-between px-6 py-4">
                  <div className="flex items-center gap-4">
                    <FolderOpen className="text-muted-foreground h-5 w-5" />
                    <div>
                      <p className="font-medium">{workspace.name}</p>
                      <p className="text-muted-foreground text-sm">
                        Last accessed {workspace.lastAccess}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-6">
                    <div className="flex items-center gap-4 text-sm">
                      <span className="flex items-center gap-1">
                        <Cpu className="text-muted-foreground h-4 w-4" />
                        {workspace.resources.cpu}%
                      </span>
                      <span className="flex items-center gap-1">
                        <MemoryStick className="text-muted-foreground h-4 w-4" />
                        {workspace.resources.memory}%
                      </span>
                    </div>
                    <span
                      className={`rounded-full px-2 py-1 text-xs font-medium ${
                        workspace.status === 'running'
                          ? 'bg-success/10 text-success'
                          : 'bg-muted text-muted-foreground'
                      }`}
                    >
                      {workspace.status}
                    </span>
                    <Button variant="ghost" size="sm">
                      <Terminal className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
              {machine.workspaces.length === 0 && (
                <div className="text-muted-foreground px-6 py-12 text-center">
                  No active workspaces
                </div>
              )}
            </div>
          </div>
        </TabsContent>

        <TabsContent value="environments">
          <div className="border-border bg-card rounded-lg border">
            <div className="border-border border-b px-6 py-4">
              <h2 className="text-lg font-medium">Active Environments</h2>
            </div>
            <div className="divide-border divide-y">
              {machine.environments.map((environment) => (
                <div key={environment.id} className="flex items-center justify-between px-6 py-4">
                  <div className="flex items-center gap-4">
                    <Server className="text-muted-foreground h-5 w-5" />
                    <div>
                      <p className="font-medium">{environment.name}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-6">
                    <div className="flex items-center gap-4 text-sm">
                      <span className="flex items-center gap-1">
                        <Cpu className="text-muted-foreground h-4 w-4" />
                        {environment.resources.cpu}%
                      </span>
                      <span className="flex items-center gap-1">
                        <MemoryStick className="text-muted-foreground h-4 w-4" />
                        {environment.resources.memory}%
                      </span>
                    </div>
                    <span
                      className={`rounded-full px-2 py-1 text-xs font-medium ${
                        environment.status === 'running'
                          ? 'bg-success/10 text-success'
                          : 'bg-muted text-muted-foreground'
                      }`}
                    >
                      {environment.status}
                    </span>
                  </div>
                </div>
              ))}
              {machine.environments.length === 0 && (
                <div className="text-muted-foreground px-6 py-12 text-center">
                  No active environments
                </div>
              )}
            </div>
          </div>
        </TabsContent>

        <TabsContent value="activity">
          <div className="border-border bg-card rounded-lg border">
            <div className="border-border border-b px-6 py-4">
              <h2 className="text-lg font-medium">Activity Log</h2>
            </div>
            <div className="divide-border divide-y">
              {machine.activityLog.map((log, index) => (
                <div key={index} className="px-6 py-4">
                  <div className="flex items-start justify-between">
                    <div>
                      <p className="text-sm font-medium">{log.action}</p>
                      <p className="text-muted-foreground mt-1 text-sm">
                        by {log.user === 'system' ? 'System' : log.user}
                      </p>
                    </div>
                    <span className="text-muted-foreground text-sm">
                      {new Date(log.timestamp).toLocaleString()}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </TabsContent>
      </Tabs>

      {/* Alerts */}
      {machine.cpu > 80 || machine.memory > 80 ? (
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            This machine is experiencing high resource usage. Consider upgrading to a higher tier or
            optimizing resource consumption.
          </AlertDescription>
        </Alert>
      ) : null}
    </div>
  )
}
