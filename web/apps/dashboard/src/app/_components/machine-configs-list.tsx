'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import { Badge } from '@kloudlite/ui'
import {
  Plus,
  MoreHorizontal,
  Edit,
  Trash2,
  Server,
  Cpu,
  HardDrive,
  Gpu,
  Power,
  PowerOff,
  Circle,
} from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'
import { Switch } from '@kloudlite/ui'
import { toast } from 'sonner'
import {
  createMachineType,
  updateMachineType,
  deleteMachineType,
  activateMachineType,
  deactivateMachineType,
} from '@/app/actions/machine-type.actions'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@kloudlite/ui'

interface MachineConfig {
  id: string
  name: string
  cpu: number
  memory: number
  gpu?: number
  description: string
  category?: 'general' | 'compute-optimized' | 'memory-optimized' | 'gpu' | 'development'
  active?: boolean
}

interface MachineConfigsListProps {
  configs: MachineConfig[]
  isReadOnly?: boolean
}

const categoryColors: Record<string, string> = {
  general: 'bg-info/10 text-info border-info/20',
  'compute-optimized': 'bg-purple-50 text-purple-700 border-purple-200',
  'memory-optimized': 'bg-orange-50 text-orange-700 border-orange-200',
  gpu: 'bg-success/10 text-success border-success/20',
  development: 'bg-muted text-foreground border-border',
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

// Helper to convert Kubernetes-friendly names back to AWS instance type format
// e.g., "m5-xlarge" -> "m5.xlarge"
function toDisplayName(k8sName: string): string {
  return k8sName.replace(/-/g, '.')
}

export function MachineConfigsList({
  configs: initialConfigs,
  isReadOnly = false,
}: MachineConfigsListProps) {
  const router = useRouter()
  const [_isPending, startTransition] = useTransition()
  const [configs, _setConfigs] = useState(initialConfigs)
  const [isAddConfigOpen, setIsAddConfigOpen] = useState(false)
  const [editingConfig, setEditingConfig] = useState<MachineConfig | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isActive, setIsActive] = useState(true)
  const [selectedCategory, setSelectedCategory] = useState<
    'general' | 'compute-optimized' | 'memory-optimized' | 'gpu' | 'development'
  >('general')

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
    } catch (_error) {
      toast.error('An error occurred while deleting')
    } finally {
      setIsLoading(false)
    }
  }

  const handleToggleActive = async (id: string, currentActive: boolean) => {
    setIsLoading(true)
    try {
      const result = currentActive ? await deactivateMachineType(id) : await activateMachineType(id)

      if (result.success) {
        toast.success(
          `Machine configuration ${currentActive ? 'deactivated' : 'activated'} successfully`,
        )
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to update machine configuration')
      }
    } catch (_error) {
      toast.error('An error occurred while updating')
    } finally {
      setIsLoading(false)
    }
  }

  const handleSave = async (formData: FormData) => {
    setIsLoading(true)
    try {
      const rawName = formData.get('name') as string
      // Convert dots to hyphens for Kubernetes-friendly resource names (e.g., "m5.xlarge" -> "m5-xlarge")
      const k8sName = rawName.replace(/\./g, '-')

      const data = {
        name: k8sName,
        displayName: formData.get('displayName') as string,
        description: formData.get('description') as string,
        cpu: parseInt(formData.get('cpu') as string),
        memory: parseInt(formData.get('memory') as string),
        gpu: formData.get('gpu') ? parseInt(formData.get('gpu') as string) : undefined,
        category: selectedCategory,
        active: isActive,
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
        toast.error(
          result.error || `Failed to ${editingConfig ? 'update' : 'create'} machine configuration`,
        )
      }
    } catch (_error) {
      toast.error('An error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      {/* Actions Bar */}
      <div className="flex items-center justify-between">
        <div className="text-muted-foreground text-sm">
          {configs.length} configuration{configs.length !== 1 ? 's' : ''} defined
        </div>

        {!isReadOnly && (
          <Button
            onClick={() => {
              setSelectedCategory('general')
              setIsActive(true)
              setIsAddConfigOpen(true)
            }}
          >
            <Plus className="mr-2 h-4 w-4" />
            Add Configuration
          </Button>
        )}
      </div>

      {/* Configurations Table */}
      <div className="bg-card rounded-lg border">
        <table className="w-full">
          <thead>
            <tr className="bg-muted border-b">
              <th className="text-foreground p-4 text-left text-sm font-medium">Configuration</th>
              <th className="text-foreground p-4 text-left text-sm font-medium">Resources</th>
              <th className="text-foreground p-4 text-left text-sm font-medium">Status</th>
              {!isReadOnly && (
                <th className="text-foreground p-4 text-left text-sm font-medium">Actions</th>
              )}
            </tr>
          </thead>
          <tbody>
            {configs.map((config) => (
              <tr key={config.id} className="hover:bg-muted border-b transition-colors">
                <td className="p-4">
                  <div className="space-y-2">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <Server className="text-muted-foreground h-4 w-4" />
                        <span className="text-foreground text-sm font-medium">{toDisplayName(config.name)}</span>
                      </div>
                      <code className="bg-muted text-muted-foreground rounded px-1.5 py-0.5 font-mono text-xs">
                        {config.id}
                      </code>
                    </div>
                    <p className="text-muted-foreground text-sm">{config.description}</p>
                    <Badge
                      variant="outline"
                      className={`${categoryColors[config.category || 'general']} text-xs`}
                    >
                      {categoryLabels[config.category || 'general']}
                    </Badge>
                  </div>
                </td>
                <td className="p-4">
                  <div className="space-y-1.5 text-sm">
                    <div className="flex items-center gap-2">
                      <Cpu className="text-muted-foreground h-3.5 w-3.5" />
                      <span className="text-foreground">{config.cpu} vCPU</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <HardDrive className="text-muted-foreground h-3.5 w-3.5" />
                      <span className="text-foreground">{config.memory}GB RAM</span>
                    </div>
                    {config.gpu && (
                      <div className="flex items-center gap-2">
                        <Gpu className="text-muted-foreground h-3.5 w-3.5" />
                        <span className="text-foreground">{config.gpu} GPU</span>
                      </div>
                    )}
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-1.5">
                    {config.active ? (
                      <>
                        <Circle className="fill-success text-success h-2 w-2" />
                        <span className="text-success text-sm font-medium">Active</span>
                      </>
                    ) : (
                      <>
                        <Circle className="fill-muted-foreground text-muted-foreground h-2 w-2" />
                        <span className="text-muted-foreground text-sm font-medium">Inactive</span>
                      </>
                    )}
                  </div>
                </td>
                {!isReadOnly && (
                  <td className="p-4">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => {
                            setEditingConfig(config)
                            setIsActive(config.active !== false)
                            setSelectedCategory(config.category || 'general')
                            setIsAddConfigOpen(true)
                          }}
                        >
                          <Edit className="mr-2 h-4 w-4" />
                          Edit Configuration
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => handleToggleActive(config.id, config.active || false)}
                        >
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
                          className="text-destructive"
                          onClick={() => handleDelete(config.id)}
                          disabled={isLoading}
                        >
                          <Trash2 className="mr-2 h-4 w-4" />
                          Delete Configuration
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>

        {configs.length === 0 && (
          <div className="p-12 text-center">
            <Server className="text-muted-foreground mx-auto mb-4 h-12 w-12" />
            <h3 className="text-foreground mb-2 text-lg font-medium">No machine configurations</h3>
            <p className="text-muted-foreground mb-4 text-sm">
              {isReadOnly
                ? 'No machine configurations have been created yet.'
                : 'Get started by creating your first machine configuration.'}
            </p>
            {!isReadOnly && (
              <Button
                onClick={() => {
                  setSelectedCategory('general')
                  setIsActive(true)
                  setIsAddConfigOpen(true)
                }}
              >
                <Plus className="mr-2 h-4 w-4" />
                Add Configuration
              </Button>
            )}
          </div>
        )}
      </div>

      {/* Add/Edit Configuration Dialog */}
      <Dialog
        open={isAddConfigOpen || !!editingConfig}
        onOpenChange={(open) => {
          if (!open) {
            setIsAddConfigOpen(false)
            setEditingConfig(null)
            setIsActive(true) // Reset to default
            setSelectedCategory('general') // Reset to default
          } else if (editingConfig) {
            setIsActive(editingConfig.active !== false)
            setSelectedCategory(editingConfig.category || 'general')
          }
        }}
      >
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editingConfig ? 'Edit Configuration' : 'Add Configuration'}</DialogTitle>
            <DialogDescription>
              Define the machine specifications and resource limits
            </DialogDescription>
          </DialogHeader>
          <form
            onSubmit={(e) => {
              e.preventDefault()
              const formData = new FormData(e.currentTarget)
              handleSave(formData)
            }}
          >
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

              {/* Category */}
              <div className="space-y-2">
                <Label htmlFor="category">
                  Category <span className="text-destructive">*</span>
                </Label>
                <Select
                  value={selectedCategory}
                  onValueChange={(value) => setSelectedCategory(value as typeof selectedCategory)}
                >
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

              {/* Resources */}
              <div className="space-y-2">
                <Label className="text-sm font-medium">Resources</Label>
                <div className="grid grid-cols-3 gap-3">
                  <div className="space-y-2">
                    <Label htmlFor="cpu" className="text-muted-foreground text-xs">
                      CPU (vCPU)
                    </Label>
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
                    <Label htmlFor="memory" className="text-muted-foreground text-xs">
                      Memory (GB)
                    </Label>
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
                    <Label htmlFor="gpu" className="text-muted-foreground text-xs">
                      GPU (optional)
                    </Label>
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
                  <Switch id="active" checked={isActive} onCheckedChange={setIsActive} />
                  <Label htmlFor="active" className="text-sm font-normal">
                    Active (available for use)
                  </Label>
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
