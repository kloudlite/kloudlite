'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Plus, MoreHorizontal, Edit, Trash2, Server, Cpu, HardDrive, DollarSign } from 'lucide-react'
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
  maxInstances: number
  activeInstances: number
  pricePerHour: number
  description: string
}

interface MachineConfigsListProps {
  configs: MachineConfig[]
}

export function MachineConfigsList({ configs: initialConfigs }: MachineConfigsListProps) {
  const [configs, setConfigs] = useState(initialConfigs)
  const [isAddConfigOpen, setIsAddConfigOpen] = useState(false)
  const [editingConfig, setEditingConfig] = useState<MachineConfig | null>(null)

  const handleDelete = (id: string) => {
    setConfigs(configs.filter(c => c.id !== id))
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
            <tr className="border-b">
              <th className="text-left p-4 font-medium text-sm text-gray-900">Configuration</th>
              <th className="text-left p-4 font-medium text-sm text-gray-900">Resources</th>
              <th className="text-left p-4 font-medium text-sm text-gray-900">Instances</th>
              <th className="text-left p-4 font-medium text-sm text-gray-900">Price</th>
              <th className="text-left p-4 font-medium text-sm text-gray-900">Actions</th>
            </tr>
          </thead>
          <tbody>
            {configs.map((config) => (
              <tr key={config.id} className="border-b hover:bg-gray-50">
                <td className="p-4">
                  <div>
                    <div className="font-medium text-sm flex items-center gap-2">
                      <Server className="h-4 w-4 text-gray-400" />
                      {config.name}
                    </div>
                    <div className="text-sm text-gray-600 mt-1">{config.description}</div>
                  </div>
                </td>
                <td className="p-4">
                  <div className="space-y-1 text-sm">
                    <div className="flex items-center gap-2">
                      <Cpu className="h-3 w-3 text-gray-400" />
                      {config.cpu} vCPU
                    </div>
                    <div className="flex items-center gap-2">
                      <HardDrive className="h-3 w-3 text-gray-400" />
                      {config.memory}GB RAM, {config.storage}GB Storage
                    </div>
                  </div>
                </td>
                <td className="p-4">
                  <div className="text-sm">
                    <div className="font-medium">{config.activeInstances} / {config.maxInstances}</div>
                    <div className="text-gray-600">active</div>
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-1 text-sm font-medium text-green-600">
                    <DollarSign className="h-3 w-3" />
                    {config.pricePerHour}/hour
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
                        setIsAddConfigOpen(true)
                      }}>
                        <Edit className="mr-2 h-4 w-4" />
                        Edit Configuration
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        className="text-red-600"
                        onClick={() => handleDelete(config.id)}
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
      </div>

      {/* Add/Edit Configuration Dialog */}
      <Dialog open={isAddConfigOpen || !!editingConfig} onOpenChange={(open) => {
        if (!open) {
          setIsAddConfigOpen(false)
          setEditingConfig(null)
        }
      }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingConfig ? 'Edit Configuration' : 'Add Configuration'}</DialogTitle>
            <DialogDescription>
              Define the machine specifications and limits
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="name">Name</Label>
                <Input id="name" defaultValue={editingConfig?.name} />
              </div>
              <div>
                <Label htmlFor="max-instances">Max Instances</Label>
                <Input
                  id="max-instances"
                  type="number"
                  defaultValue={editingConfig?.maxInstances}
                />
              </div>
            </div>
            <div>
              <Label htmlFor="description">Description</Label>
              <Input id="description" defaultValue={editingConfig?.description} />
            </div>
            <div className="grid grid-cols-3 gap-4">
              <div>
                <Label htmlFor="cpu">CPU (vCPU)</Label>
                <Input
                  id="cpu"
                  type="number"
                  defaultValue={editingConfig?.cpu}
                />
              </div>
              <div>
                <Label htmlFor="memory">Memory (GB)</Label>
                <Input
                  id="memory"
                  type="number"
                  defaultValue={editingConfig?.memory}
                />
              </div>
              <div>
                <Label htmlFor="storage">Storage (GB)</Label>
                <Input
                  id="storage"
                  type="number"
                  defaultValue={editingConfig?.storage}
                />
              </div>
            </div>
            <div>
              <Label htmlFor="price">Price per Hour ($)</Label>
              <Input
                id="price"
                type="number"
                step="0.01"
                defaultValue={editingConfig?.pricePerHour}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setIsAddConfigOpen(false)
              setEditingConfig(null)
            }}>
              Cancel
            </Button>
            <Button>
              {editingConfig ? 'Update' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}