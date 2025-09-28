'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Plus, MoreHorizontal, Edit, Trash2, Server, Cpu, HardDrive, DollarSign, Gpu, Power, PowerOff, CheckCircle, Circle } from 'lucide-react'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { toast } from 'sonner'
import {
  createMachineType,
  updateMachineType,
  deleteMachineType,
  activateMachineType,
  deactivateMachineType
} from '@/app/actions/machine-type.actions'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface MachineConfig {
  id: string
  name: string
  cpu: number
  memory: number
  storage: number
  gpu?: number
  maxInstances: number
  activeInstances: number
  pricePerHour: number
  description: string
  category?: 'general' | 'compute' | 'memory' | 'gpu'
  active?: boolean
}

interface MachineConfigsListProps {
  configs: MachineConfig[]
}

const categoryColors: Record<string, string> = {
  general: 'bg-blue-50 text-blue-700 border-blue-200',
  'compute-optimized': 'bg-purple-50 text-purple-700 border-purple-200',
  'memory-optimized': 'bg-orange-50 text-orange-700 border-orange-200',
  gpu: 'bg-green-50 text-green-700 border-green-200',
  development: 'bg-gray-50 text-gray-700 border-gray-200',
  // Fallback for old values
  compute: 'bg-purple-50 text-purple-700 border-purple-200',
  memory: 'bg-orange-50 text-orange-700 border-orange-200',
}

const categoryLabels: Record<string, string> = {
  general: 'General',
  'compute-optimized': 'Compute Optimized',
  'memory-optimized': 'Memory Optimized',
  gpu: 'GPU',
  development: 'Development',
  // Fallback for old values
  compute: 'Compute',
  memory: 'Memory',
}

export function MachineConfigsList({ configs: initialConfigs }: MachineConfigsListProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [configs, setConfigs] = useState(initialConfigs)
  const [isAddConfigOpen, setIsAddConfigOpen] = useState(false)
  const [editingConfig, setEditingConfig] = useState<MachineConfig | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isActive, setIsActive] = useState(true)

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this machine configuration?')) {
      return
    }

    setIsLoading(true)
    try {
      const result = await deleteMachineType(id)
      if (result.success) {
        toast.success('Machine configuration deleted successfully')
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to delete machine configuration')
      }
    } catch (error) {
      toast.error('An error occurred while deleting')
    } finally {
      setIsLoading(false)
    }
  }

  const handleToggleActive = async (id: string, currentActive: boolean) => {
    setIsLoading(true)
    try {
      const result = currentActive
        ? await deactivateMachineType(id)
        : await activateMachineType(id)

      if (result.success) {
        toast.success(`Machine configuration ${currentActive ? 'deactivated' : 'activated'} successfully`)
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to update machine configuration')
      }
    } catch (error) {
      toast.error('An error occurred while updating')
    } finally {
      setIsLoading(false)
    }
  }

  const handleSave = async (formData: FormData) => {
    setIsLoading(true)
    try {
      const data = {
        name: formData.get('name') as string,
        displayName: formData.get('displayName') as string,
        description: formData.get('description') as string,
        cpu: parseInt(formData.get('cpu') as string),
        memory: parseInt(formData.get('memory') as string),
        storage: parseInt(formData.get('storage') as string),
        gpu: formData.get('gpu') ? parseInt(formData.get('gpu') as string) : undefined,
        category: (formData.get('category') as 'general' | 'compute' | 'memory' | 'gpu') || 'general',
        pricePerHour: parseFloat(formData.get('pricePerHour') as string),
        active: isActive
      }

      const result = editingConfig
        ? await updateMachineType(editingConfig.id, data)
        : await createMachineType(data)

      if (result.success) {
        toast.success(`Machine configuration ${editingConfig ? 'updated' : 'created'} successfully`)
        setIsAddConfigOpen(false)
        setEditingConfig(null)
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || `Failed to ${editingConfig ? 'update' : 'create'} machine configuration`)
      }
    } catch (error) {
      toast.error('An error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      {/* Actions Bar */}
      <div className="flex items-center justify-between">
        <div className="text-sm text-gray-600">
          {configs.length} configuration{configs.length !== 1 ? 's' : ''} defined
        </div>

        <Button onClick={() => setIsAddConfigOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Add Configuration
        </Button>
      </div>

      {/* Configurations Table */}
      <div className="bg-white rounded-lg border">
        <table className="w-full">
          <thead>
            <tr className="border-b bg-gray-50/50">
              <th className="text-left p-4 font-medium text-sm text-gray-700">Configuration</th>
              <th className="text-left p-4 font-medium text-sm text-gray-700">Resources</th>
              <th className="text-left p-4 font-medium text-sm text-gray-700">Instances</th>
              <th className="text-left p-4 font-medium text-sm text-gray-700">Price</th>
              <th className="text-left p-4 font-medium text-sm text-gray-700">Status</th>
              <th className="text-left p-4 font-medium text-sm text-gray-700">Actions</th>
            </tr>
          </thead>
          <tbody>
            {configs.map((config) => (
              <tr key={config.id} className="border-b hover:bg-gray-50/50 transition-colors">
                <td className="p-4">
                  <div className="space-y-2">
                    <div className="flex items-center gap-2">
                      <Server className="h-4 w-4 text-gray-400" />
                      <span className="font-medium text-sm text-gray-900">{config.name}</span>
                    </div>
                    <p className="text-sm text-gray-500">{config.description}</p>
                    <Badge variant="outline" className={`${categoryColors[config.category || 'general']} text-xs`}>
                      {categoryLabels[config.category || 'general']}
                    </Badge>
                  </div>
                </td>
                <td className="p-4">
                  <div className="space-y-1.5 text-sm">
                    <div className="flex items-center gap-2">
                      <Cpu className="h-3.5 w-3.5 text-gray-400" />
                      <span className="text-gray-700">{config.cpu} vCPU</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <HardDrive className="h-3.5 w-3.5 text-gray-400" />
                      <span className="text-gray-700">{config.memory}GB RAM, {config.storage}GB Storage</span>
                    </div>
                    {config.gpu && (
                      <div className="flex items-center gap-2">
                        <Gpu className="h-3.5 w-3.5 text-gray-400" />
                        <span className="text-gray-700">{config.gpu} GPU</span>
                      </div>
                    )}
                  </div>
                </td>
                <td className="p-4">
                  <div className="space-y-1">
                    <div className="text-sm font-medium text-gray-900">
                      {config.activeInstances} / {config.maxInstances}
                    </div>
                    <div className="w-24 bg-gray-200 rounded-full h-1.5">
                      <div
                        className="bg-blue-600 h-1.5 rounded-full"
                        style={{ width: `${(config.activeInstances / config.maxInstances) * 100}%` }}
                      />
                    </div>
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-1 text-sm font-semibold text-green-600">
                    <DollarSign className="h-3.5 w-3.5" />
                    {config.pricePerHour.toFixed(2)}/hour
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-1.5">
                    {config.active ? (
                      <>
                        <Circle className="h-2 w-2 fill-green-500 text-green-500" />
                        <span className="text-sm text-green-600 font-medium">Active</span>
                      </>
                    ) : (
                      <>
                        <Circle className="h-2 w-2 fill-gray-400 text-gray-400" />
                        <span className="text-sm text-gray-500 font-medium">Inactive</span>
                      </>
                    )}
                  </div>
                </td>
                <td className="p-4">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => {
                        setEditingConfig(config)
                        setIsActive(config.active !== false)
                        setIsAddConfigOpen(true)
                      }}>
                        <Edit className="mr-2 h-4 w-4" />
                        Edit Configuration
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => handleToggleActive(config.id, config.active || false)}>
                        {config.active ? (
                          <>
                            <PowerOff className="mr-2 h-4 w-4" />
                            Deactivate
                          </>
                        ) : (
                          <>
                            <Power className="mr-2 h-4 w-4" />
                            Activate
                          </>
                        )}
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        className="text-red-600"
                        onClick={() => handleDelete(config.id)}
                        disabled={isLoading}
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Delete Configuration
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {configs.length === 0 && (
          <div className="p-12 text-center">
            <Server className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No machine configurations</h3>
            <p className="text-sm text-gray-500 mb-4">Get started by creating your first machine configuration.</p>
            <Button onClick={() => setIsAddConfigOpen(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Add Configuration
            </Button>
          </div>
        )}
      </div>

      {/* Add/Edit Configuration Dialog */}
      <Dialog open={isAddConfigOpen || !!editingConfig} onOpenChange={(open) => {
        if (!open) {
          setIsAddConfigOpen(false)
          setEditingConfig(null)
          setIsActive(true) // Reset to default
        } else if (editingConfig) {
          setIsActive(editingConfig.active !== false)
        }
      }}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editingConfig ? 'Edit Configuration' : 'Add Configuration'}</DialogTitle>
            <DialogDescription>
              Define the machine specifications and resource limits
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={(e) => {
            e.preventDefault()
            const formData = new FormData(e.currentTarget)
            handleSave(formData)
          }}>
            <div className="space-y-5 py-4">
              {/* Basic Information */}
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="name">Machine Type ID</Label>
                    <Input
                      id="name"
                      name="name"
                      placeholder="e.g. small-4x8"
                      defaultValue={editingConfig?.id}
                      disabled={!!editingConfig}
                      required
                      pattern="^[a-z0-9-]+$"
                      title="Only lowercase letters, numbers, and hyphens allowed"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="displayName">Display Name</Label>
                    <Input
                      id="displayName"
                      name="displayName"
                      placeholder="e.g. Small Instance"
                      defaultValue={editingConfig?.name}
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="description">Description</Label>
                  <Input
                    id="description"
                    name="description"
                    placeholder="e.g. Suitable for light workloads and development"
                    defaultValue={editingConfig?.description}
                  />
                </div>
              </div>

              {/* Category and Pricing */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="category">Category <span className="text-red-500">*</span></Label>
                  <Select name="category" defaultValue={editingConfig?.category || 'general'} required>
                    <SelectTrigger>
                      <SelectValue placeholder="Select a category" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="general">General Purpose</SelectItem>
                      <SelectItem value="compute-optimized">Compute Optimized</SelectItem>
                      <SelectItem value="memory-optimized">Memory Optimized</SelectItem>
                      <SelectItem value="gpu">GPU Accelerated</SelectItem>
                      <SelectItem value="development">Development</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="pricePerHour">Price per Hour ($)</Label>
                  <Input
                    id="pricePerHour"
                    name="pricePerHour"
                    type="number"
                    step="0.01"
                    placeholder="0.00"
                    defaultValue={editingConfig?.pricePerHour}
                    required
                  />
                </div>
              </div>

              {/* Resources */}
              <div className="space-y-2">
                <Label className="text-sm font-medium">Resources</Label>
                <div className="grid grid-cols-4 gap-3">
                  <div className="space-y-2">
                    <Label htmlFor="cpu" className="text-xs text-gray-600">CPU (vCPU)</Label>
                    <Input
                      id="cpu"
                      name="cpu"
                      type="number"
                      placeholder="4"
                      defaultValue={editingConfig?.cpu}
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="memory" className="text-xs text-gray-600">Memory (GB)</Label>
                    <Input
                      id="memory"
                      name="memory"
                      type="number"
                      placeholder="8"
                      defaultValue={editingConfig?.memory}
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="storage" className="text-xs text-gray-600">Storage (GB)</Label>
                    <Input
                      id="storage"
                      name="storage"
                      type="number"
                      placeholder="100"
                      defaultValue={editingConfig?.storage}
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="gpu" className="text-xs text-gray-600">GPU (optional)</Label>
                    <Input
                      id="gpu"
                      name="gpu"
                      type="number"
                      placeholder="0"
                      defaultValue={editingConfig?.gpu}
                    />
                  </div>
                </div>
              </div>

              {/* Status */}
              <div className="border-t pt-4">
                <div className="flex items-center space-x-2">
                  <Switch
                    id="active"
                    checked={isActive}
                    onCheckedChange={setIsActive}
                  />
                  <Label htmlFor="active" className="text-sm font-normal">Active (available for use)</Label>
                </div>
              </div>
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setIsAddConfigOpen(false)
                  setEditingConfig(null)
                }}
                disabled={isLoading}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Saving...' : editingConfig ? 'Update' : 'Create'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}