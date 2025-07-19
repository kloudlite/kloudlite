'use client'

import { useState } from 'react'
import { cn } from '@/lib/utils'
import { LAYOUT } from '@/lib/constants/layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Separator } from '@/components/ui/separator'
import { CompleteSection } from '@/components/ui/section'
import { OverviewCard, OverviewGrid } from '@/components/ui/overview-card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
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
import {
  Server,
  Database,
  Network,
  Shield,
  Settings,
  Plus,
  Trash2,
  CheckCircle,
  AlertTriangle,
  XCircle,
  Cpu,
  HardDrive,
  Activity,
  Clock,
  MapPin,
  TrendingUp,
} from 'lucide-react'
import type { Team, TeamRole } from '@/lib/teams/types'

interface TeamInfrastructureSettingsProps {
  team: Team & { userRole: TeamRole }
  isOwner: boolean
}

// Mock infrastructure data
const mockWorkMachines = [
  {
    id: 'wm-01',
    developer: 'Sarah Chen',
    status: 'running' as const,
    region: 'us-west-2',
    instanceType: 'm5.2xlarge',
    cpuCores: 8,
    memoryGB: 32,
    storageGB: 100,
    createdAt: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
  },
  {
    id: 'wm-02',
    developer: 'Alex Kumar',
    status: 'stopped' as const,
    region: 'us-west-2',
    instanceType: 'm5.xlarge',
    cpuCores: 4,
    memoryGB: 16,
    storageGB: 80,
    createdAt: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000),
  },
]

const mockNodePools = [
  {
    id: 'np-01',
    name: 'development-pool',
    status: 'healthy' as const,
    nodeCount: 8,
    instanceType: 'm5.2xlarge',
    region: 'us-west-2',
    autoscaleEnabled: true,
    minNodes: 4,
    maxNodes: 12,
    createdAt: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000),
  },
  {
    id: 'np-02',
    name: 'production-pool',
    status: 'healthy' as const,
    nodeCount: 12,
    instanceType: 'm5.4xlarge',
    region: 'us-east-1',
    autoscaleEnabled: true,
    minNodes: 8,
    maxNodes: 20,
    createdAt: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000),
  },
]

const mockSharedServices = [
  {
    id: 'svc-01',
    name: 'redis-cache',
    type: 'Redis',
    status: 'healthy' as const,
    endpoint: 'redis.internal.com:6379',
    region: 'us-west-2',
    createdAt: new Date(Date.now() - 45 * 24 * 60 * 60 * 1000),
  },
  {
    id: 'svc-02',
    name: 'postgres-db',
    type: 'PostgreSQL',
    status: 'healthy' as const,
    endpoint: 'postgres.internal.com:5432',
    region: 'us-west-2',
    createdAt: new Date(Date.now() - 120 * 24 * 60 * 60 * 1000),
  },
]

function getStatusIcon(status: 'running' | 'stopped' | 'healthy' | 'degraded' | 'failed') {
  switch (status) {
    case 'running':
    case 'healthy':
      return <CheckCircle className="h-4 w-4 text-current" />
    case 'degraded':
      return <AlertTriangle className="h-4 w-4 text-current" />
    case 'stopped':
    case 'failed':
      return <XCircle className="h-4 w-4 text-current" />
    default:
      return <Activity className="h-4 w-4 text-current" />
  }
}

function getStatusBadgeVariant(status: string): "default" | "secondary" | "destructive" | "outline" {
  switch (status) {
    case 'running':
    case 'healthy':
      return 'default'
    case 'stopped':
      return 'secondary'
    case 'degraded':
      return 'outline'
    case 'failed':
      return 'destructive'
    default:
      return 'outline'
  }
}

export function TeamInfrastructureSettings({ team, isOwner }: TeamInfrastructureSettingsProps) {
  const [autoScalingEnabled, setAutoScalingEnabled] = useState(true)
  const [resourceLimitsEnabled, setResourceLimitsEnabled] = useState(true)
  const [monitoringEnabled, setMonitoringEnabled] = useState(true)

  return (
    <div className={cn(LAYOUT.SPACING.SECTION, "overflow-x-hidden")}>
      {/* Infrastructure Overview */}
      <OverviewGrid columns={3}>
        <OverviewCard
          icon={<Server className="h-5 w-5 text-primary flex-shrink-0" />}
          title="Work Machines"
          value={mockWorkMachines.length}
        />
        <OverviewCard
          icon={<Network className="h-5 w-5 text-primary flex-shrink-0" />}
          title="Node Pools"
          value={mockNodePools.length}
        />
        <OverviewCard
          icon={<Database className="h-5 w-5 text-primary flex-shrink-0" />}
          title="Shared Services"
          value={mockSharedServices.length}
          className="sm:col-span-2 md:col-span-1"
        />
      </OverviewGrid>

      {/* Infrastructure Policies */}
      <CompleteSection
        title="Infrastructure Policies"
        icon={<Shield className="h-5 w-5" />}
        spacing="loose"
      >
          <div className="flex items-start justify-between gap-3">
            <div className="space-y-1 flex-1">
              <Label className="text-base font-medium">Auto-scaling</Label>
              <p className="text-sm text-muted-foreground">
                Automatically scale resources based on demand
              </p>
            </div>
            <Switch
              checked={autoScalingEnabled}
              onCheckedChange={setAutoScalingEnabled}
            />
          </div>

          <Separator />

          <div className="flex items-start justify-between gap-3">
            <div className="space-y-1 flex-1">
              <Label className="text-base font-medium">Resource Limits</Label>
              <p className="text-sm text-muted-foreground">
                Enforce CPU and memory limits on workloads
              </p>
            </div>
            <Switch
              checked={resourceLimitsEnabled}
              onCheckedChange={setResourceLimitsEnabled}
            />
          </div>

          <Separator />

          <div className="flex items-start justify-between gap-3">
            <div className="space-y-1 flex-1">
              <Label className="text-base font-medium">Monitoring & Alerts</Label>
              <p className="text-sm text-muted-foreground">
                Enable infrastructure monitoring and alerting
              </p>
            </div>
            <Switch
              checked={monitoringEnabled}
              onCheckedChange={setMonitoringEnabled}
            />
          </div>

          {resourceLimitsEnabled && (
            <>
              <Separator />
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-4">
                <div className="space-y-2">
                  <Label htmlFor="cpu-limit">Default CPU Limit (cores)</Label>
                  <Input id="cpu-limit" type="number" placeholder="4" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="memory-limit">Default Memory Limit (GB)</Label>
                  <Input id="memory-limit" type="number" placeholder="8" />
                </div>
              </div>
            </>
          )}

          <div className="pt-4">
            <Button>Save Infrastructure Policies</Button>
          </div>
      </CompleteSection>

      {/* Work Machines */}
      <CompleteSection
        title="Work Machines"
        icon={<Server className="h-5 w-5" />}
        actions={
          <Button size="sm">
            <Plus className="h-4 w-4 mr-2" />
            Add Machine
          </Button>
        }
        contentClassName="p-0"
      >
        
        {/* Mobile: Card Layout */}
        <div className="block lg:hidden">
          {mockWorkMachines.map((machine, index) => (
            <div key={machine.id} className={`p-6 space-y-4 ${index !== mockWorkMachines.length - 1 ? 'border-b' : ''}`}>
                <div className="flex items-start justify-between">
                  <div className="min-w-0 flex-1">
                    <h4 className="font-medium truncate">{machine.developer}</h4>
                    <p className="text-sm text-muted-foreground mt-0.5">Work Machine</p>
                  </div>
                  <div className="flex items-center gap-2 ml-2">
                    <Badge variant={getStatusBadgeVariant(machine.status)} className="flex items-center gap-1">
                      {getStatusIcon(machine.status)}
                      {machine.status}
                    </Badge>
                    {isOwner && (
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>Delete Work Machine</AlertDialogTitle>
                            <AlertDialogDescription>
                              Are you sure you want to delete {machine.developer}'s work machine? This action cannot be undone.
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                              Delete Machine
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    )}
                  </div>
                </div>
                
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <p className="text-muted-foreground">Instance Type</p>
                    <div>
                      <p className="font-medium">{machine.instanceType}</p>
                      <p className="text-xs text-muted-foreground">
                        {machine.cpuCores}C / {machine.memoryGB}GB
                      </p>
                    </div>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Region</p>
                    <div className="flex items-center gap-1">
                      <MapPin className="h-3 w-3" />
                      <span>{machine.region}</span>
                    </div>
                  </div>
                  <div className="col-span-2">
                    <p className="text-muted-foreground">Created</p>
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      <span>{machine.createdAt.toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

        {/* Desktop: Table Layout */}
        <div className="hidden lg:block overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b text-sm text-muted-foreground">
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Developer</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Status</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Instance Type</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Region</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Created</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3 w-16"></th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {mockWorkMachines.map((machine, index) => (
                <tr key={machine.id} className="hover:bg-muted/50 transition-colors group focus-within:bg-muted/50">
                  <td className="px-6 py-4">
                    <div className="font-medium">{machine.developer}</div>
                    <div className="text-sm text-muted-foreground mt-0.5">Work Machine</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1.5">
                      {getStatusIcon(machine.status)}
                      <span className="text-sm capitalize">{machine.status}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="font-medium">{machine.instanceType}</div>
                    <div className="text-sm text-muted-foreground">
                      {machine.cpuCores}C / {machine.memoryGB}GB
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1">
                      <MapPin className="h-3 w-3" />
                      <span className="text-sm">{machine.region}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-sm text-muted-foreground">
                      {machine.createdAt.toLocaleDateString()}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    {isOwner && (
                      <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity">
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button variant="ghost" size="icon-sm" aria-label={`Delete ${machine.developer}'s work machine`}>
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Delete Work Machine</AlertDialogTitle>
                              <AlertDialogDescription>
                                Are you sure you want to delete {machine.developer}'s work machine? This action cannot be undone.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                                Delete Machine
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </CompleteSection>

      {/* Node Pools */}
      <div className="bg-background border rounded-lg overflow-hidden">
        {/* Table Header */}
        <div className="border-b px-6 py-4">
          <div className="flex items-center justify-between">
            <h2 className="font-semibold flex items-center gap-2">
              <Network className="h-5 w-5" />
              Node Pools
            </h2>
            <div className="flex items-center gap-2">
              <Button size="sm">
                <Plus className="h-4 w-4 mr-2" />
                Add Pool
              </Button>
            </div>
          </div>
        </div>
        
        {/* Mobile: Card Layout */}
        <div className="block lg:hidden">
          {mockNodePools.map((pool, index) => (
            <div key={pool.id} className={`p-6 space-y-4 ${index !== mockNodePools.length - 1 ? 'border-b' : ''}`}>
                <div className="flex items-start justify-between">
                  <div className="min-w-0 flex-1">
                    <h4 className="font-medium truncate">{pool.name}</h4>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge variant={getStatusBadgeVariant(pool.status)} className="flex items-center gap-1">
                        {getStatusIcon(pool.status)}
                        {pool.status}
                      </Badge>
                    </div>
                  </div>
                  {isOwner && (
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Delete Node Pool</AlertDialogTitle>
                          <AlertDialogDescription>
                            Are you sure you want to delete {pool.name}? This action cannot be undone.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                            Delete Pool
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  )}
                </div>
                
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <p className="text-muted-foreground">Nodes</p>
                    <div>
                      <p className="font-medium">{pool.nodeCount} nodes</p>
                      {pool.autoscaleEnabled && (
                        <div className="flex items-center gap-1 text-xs text-primary">
                          <TrendingUp className="h-3 w-3" />
                          <span>{pool.minNodes}-{pool.maxNodes}</span>
                        </div>
                      )}
                    </div>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Instance Type</p>
                    <p className="font-medium">{pool.instanceType}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Region</p>
                    <div className="flex items-center gap-1">
                      <MapPin className="h-3 w-3" />
                      <span>{pool.region}</span>
                    </div>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Created</p>
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      <span>{pool.createdAt.toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

        {/* Desktop: Table Layout */}
        <div className="hidden lg:block overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b text-sm text-muted-foreground">
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Name</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Status</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Nodes</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Instance Type</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Region</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Created</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3 w-16"></th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {mockNodePools.map((pool, index) => (
                <tr key={pool.id} className="hover:bg-muted/50 transition-colors group focus-within:bg-muted/50">
                  <td className="px-6 py-4">
                    <div className="font-medium">{pool.name}</div>
                    <div className="text-sm text-muted-foreground mt-0.5">Kubernetes Node Pool</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1.5">
                      {getStatusIcon(pool.status)}
                      <span className="text-sm capitalize">{pool.status}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="font-medium">{pool.nodeCount} nodes</div>
                    {pool.autoscaleEnabled && (
                      <div className="flex items-center gap-1 text-xs text-primary">
                        <TrendingUp className="h-3 w-3" />
                        <span>{pool.minNodes}-{pool.maxNodes}</span>
                      </div>
                    )}
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-sm">{pool.instanceType}</span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1">
                      <MapPin className="h-3 w-3" />
                      <span className="text-sm">{pool.region}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-sm text-muted-foreground">
                      {pool.createdAt.toLocaleDateString()}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    {isOwner && (
                      <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity">
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button variant="ghost" size="icon-sm" aria-label={`Delete ${pool.name}`}>
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Delete Node Pool</AlertDialogTitle>
                              <AlertDialogDescription>
                                Are you sure you want to delete {pool.name}? This action cannot be undone.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                                Delete Pool
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Shared Services */}
      <div className="bg-background border rounded-lg overflow-hidden">
        {/* Table Header */}
        <div className="border-b px-6 py-4">
          <div className="flex items-center justify-between">
            <h2 className="font-semibold flex items-center gap-2">
              <Database className="h-5 w-5" />
              Shared Services
            </h2>
            <div className="flex items-center gap-2">
              <Button size="sm">
                <Plus className="h-4 w-4 mr-2" />
                Add Service
              </Button>
            </div>
          </div>
        </div>
        
        {/* Mobile: Card Layout */}
        <div className="block lg:hidden">
          {mockSharedServices.map((service, index) => (
            <div key={service.id} className={`p-6 space-y-4 ${index !== mockSharedServices.length - 1 ? 'border-b' : ''}`}>
                <div className="flex items-start justify-between">
                  <div className="min-w-0 flex-1">
                    <h4 className="font-medium truncate">{service.name}</h4>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge variant="outline">{service.type}</Badge>
                      <Badge variant={getStatusBadgeVariant(service.status)} className="flex items-center gap-1">
                        {getStatusIcon(service.status)}
                        {service.status}
                      </Badge>
                    </div>
                  </div>
                  {isOwner && (
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Delete Shared Service</AlertDialogTitle>
                          <AlertDialogDescription>
                            Are you sure you want to delete {service.name}? This action cannot be undone.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                            Delete Service
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  )}
                </div>
                
                <div className="space-y-2 text-sm">
                  <div>
                    <p className="text-muted-foreground">Endpoint</p>
                    <p className="font-mono text-xs break-all">{service.endpoint}</p>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <p className="text-muted-foreground">Region</p>
                      <div className="flex items-center gap-1">
                        <MapPin className="h-3 w-3" />
                        <span>{service.region}</span>
                      </div>
                    </div>
                    <div>
                      <p className="text-muted-foreground">Created</p>
                      <div className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        <span>{service.createdAt.toLocaleDateString()}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

        {/* Desktop: Table Layout */}
        <div className="hidden lg:block overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b text-sm text-muted-foreground">
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Name</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Type</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Status</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Endpoint</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Region</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3">Created</th>
                <th className="text-left font-medium text-muted-foreground px-6 py-3 w-16"></th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {mockSharedServices.map((service, index) => (
                <tr key={service.id} className="hover:bg-muted/50 transition-colors group focus-within:bg-muted/50">
                  <td className="px-6 py-4">
                    <div className="font-medium">{service.name}</div>
                    <div className="text-sm text-muted-foreground mt-0.5">Shared Database Service</div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-sm">{service.type}</span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1.5">
                      {getStatusIcon(service.status)}
                      <span className="text-sm capitalize">{service.status}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="font-mono text-sm max-w-xs truncate">{service.endpoint}</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1">
                      <MapPin className="h-3 w-3" />
                      <span className="text-sm">{service.region}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-sm text-muted-foreground">
                      {service.createdAt.toLocaleDateString()}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    {isOwner && (
                      <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity">
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button variant="ghost" size="icon-sm" aria-label={`Delete ${service.name}`}>
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Delete Shared Service</AlertDialogTitle>
                              <AlertDialogDescription>
                                Are you sure you want to delete {service.name}? This action cannot be undone.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                                Delete Service
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}