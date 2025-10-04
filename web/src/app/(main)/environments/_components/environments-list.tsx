'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Plus, MoreHorizontal, ExternalLink, Power, PowerOff, Edit, Loader2 } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { CreateEnvironmentDialog } from '@/components/dialogs/create-environment'
import { EditEnvironmentDialog } from '@/components/dialogs/edit-environment'
import { DeleteEnvironmentConfirm } from '@/components/dialogs/delete-environment-confirm'
import { activateEnvironment, deactivateEnvironment } from '@/app/actions/environment.actions'
import { toast } from 'sonner'
import type { EnvironmentUIModel } from '@/types/environment'

interface EnvironmentsListProps {
  environments: EnvironmentUIModel[]
  currentUser: string
}

export function EnvironmentsList({ environments: initialEnvironments, currentUser }: EnvironmentsListProps) {
  const [scopeFilter, setScope] = useState<'all' | 'mine'>('all')
  const [statusFilter, setStatus] = useState<'all' | 'active'>('all')
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [selectedEnvironment, setSelectedEnvironment] = useState<EnvironmentUIModel | null>(null)
  const [deleteEnvironmentName, setDeleteEnvironmentName] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()
  const [environments, setEnvironments] = useState<EnvironmentUIModel[]>(initialEnvironments)
  const router = useRouter()

  // Poll for environment updates every 3 seconds
  useEffect(() => {
    const pollInterval = setInterval(() => {
      // Check if any environment is in a transitional state
      const hasTransitionalEnv = environments.some(
        env => env.status === 'deleting' || env.status === 'activating' || env.status === 'deactivating'
      )

      if (hasTransitionalEnv) {
        router.refresh()
      }
    }, 3000)

    return () => clearInterval(pollInterval)
  }, [environments, router])

  // Update local state when server data changes
  useEffect(() => {
    setEnvironments(initialEnvironments)
  }, [initialEnvironments])

  let filteredEnvironments = environments

  // Apply scope filter
  if (scopeFilter === 'mine') {
    filteredEnvironments = filteredEnvironments.filter(env => env.owner === currentUser)
  }

  // Apply status filter
  if (statusFilter === 'active') {
    filteredEnvironments = filteredEnvironments.filter(env => env.status === 'active')
  }

  const counts = {
    all: environments.length,
    mine: environments.filter(env => env.owner === currentUser).length,
    active: environments.filter(env => env.status === 'active').length,
    allActive: environments.filter(env => env.status === 'active').length,
    mineActive: environments.filter(env => env.owner === currentUser && env.status === 'active').length,
  }

  const handleActivate = async (envName: string) => {
    try {
      const result = await activateEnvironment(envName, currentUser)
      if (result.success) {
        toast.success('Environment activated', {
          description: `${envName} has been activated successfully.`,
        })
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error('Failed to activate environment', {
          description: result.error || 'An error occurred',
        })
      }
    } catch (error: any) {
      toast.error('Failed to activate environment', {
        description: error.message || 'An error occurred',
      })
    }
  }

  const handleDeactivate = async (envName: string) => {
    try {
      const result = await deactivateEnvironment(envName, currentUser)
      if (result.success) {
        toast.success('Environment deactivated', {
          description: `${envName} has been deactivated successfully.`,
        })
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error('Failed to deactivate environment', {
          description: result.error || 'An error occurred',
        })
      }
    } catch (error: any) {
      toast.error('Failed to deactivate environment', {
        description: error.message || 'An error occurred',
      })
    }
  }

  const handleEditClick = (env: EnvironmentUIModel) => {
    setSelectedEnvironment(env)
    setEditDialogOpen(true)
  }

  const handleDeleteClick = (envName: string) => {
    setDeleteEnvironmentName(envName)
    setDeleteConfirmOpen(true)
  }

  const handleCreateSuccess = () => {
    toast.success('Environment created', {
      description: 'Your new environment has been created successfully.',
    })
    startTransition(() => {
      router.refresh()
    })
  }

  const handleDeleteSuccess = () => {
    toast.success('Environment deleted', {
      description: 'The environment has been deleted successfully.',
    })
    startTransition(() => {
      router.refresh()
    })
  }

  return (
    <div className="space-y-4">
      {/* Filter and Actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          {/* Scope Filter */}
          <div className="flex items-center gap-1 p-1 bg-gray-100 rounded-md">
            <button
              onClick={() => setScope('all')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                scopeFilter === 'all'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setScope('mine')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                scopeFilter === 'mine'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Mine
            </button>
          </div>

          {/* Status Filter */}
          <div className="flex items-center gap-1 p-1 bg-gray-100 rounded-md">
            <button
              onClick={() => setStatus('all')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                statusFilter === 'all'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setStatus('active')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                statusFilter === 'active'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Active
            </button>
          </div>

          <span className="text-sm text-gray-500">
            {filteredEnvironments.length} {filteredEnvironments.length === 1 ? 'environment' : 'environments'}
          </span>
        </div>
        <Button size="sm" className="gap-2" onClick={() => setCreateDialogOpen(true)}>
          <Plus className="h-4 w-4" />
          New Environment
        </Button>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <table className="min-w-full">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Owner
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Resources
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Last Deployed
              </th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {filteredEnvironments.map((env) => (
              <tr key={env.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <Link
                    href={`/environments/${env.id}`}
                    className="text-sm font-semibold text-gray-900 hover:text-blue-600 flex items-center gap-1"
                  >
                    {env.name}
                    <ExternalLink className="h-3 w-3" />
                  </Link>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                  {env.owner.split('@')[0]}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                    env.status === 'active'
                      ? 'bg-green-100 text-green-800'
                      : env.status === 'deleting'
                      ? 'bg-red-100 text-red-800'
                      : env.status === 'activating' || env.status === 'deactivating'
                      ? 'bg-blue-100 text-blue-800'
                      : 'bg-gray-100 text-gray-600'
                  }`}>
                    {(env.status === 'deleting' || env.status === 'activating' || env.status === 'deactivating') && (
                      <Loader2 className="h-3 w-3 animate-spin" />
                    )}
                    {env.status}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                  <div className="flex items-center gap-4">
                    <span>{env.services} services</span>
                    <span className="text-gray-400">•</span>
                    <span>{env.configs} configs</span>
                    <span className="text-gray-400">•</span>
                    <span>{env.secrets} secrets</span>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                  {env.lastDeployed}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                  {env.status === 'deleting' ? (
                    <span className="text-xs text-gray-500">Deleting...</span>
                  ) : (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem asChild>
                          <Link href={`/environments/${env.id}`}>
                            View Details
                          </Link>
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => handleEditClick(env)}
                        >
                          <Edit className="h-4 w-4 mr-2" />
                          Edit Settings
                        </DropdownMenuItem>
                        {env.status === 'active' ? (
                          <DropdownMenuItem
                            onClick={() => handleDeactivate(env.name)}
                            className="text-orange-600"
                          >
                            <PowerOff className="h-4 w-4 mr-2" />
                            Deactivate
                          </DropdownMenuItem>
                        ) : (
                          <DropdownMenuItem
                            onClick={() => handleActivate(env.name)}
                            className="text-green-600"
                          >
                            <Power className="h-4 w-4 mr-2" />
                            Activate
                          </DropdownMenuItem>
                        )}
                        <DropdownMenuItem>Clone Environment</DropdownMenuItem>
                        <DropdownMenuItem>Export Config</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-red-600"
                          onClick={() => handleDeleteClick(env.name)}
                        >
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {filteredEnvironments.length === 0 && (
        <div className="bg-white rounded-lg border border-gray-200 text-center py-12">
          <p className="text-sm text-gray-500">
            {scopeFilter === 'mine' && statusFilter === 'active'
              ? "You don't have any active environments"
              : scopeFilter === 'mine'
              ? "You don't have any environments yet"
              : statusFilter === 'active'
              ? "No active environments found"
              : "No environments found"}
          </p>
        </div>
      )}

      {/* Dialogs */}
      <CreateEnvironmentDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSuccess={handleCreateSuccess}
        currentUser={currentUser}
      />

      {selectedEnvironment && (
        <EditEnvironmentDialog
          open={editDialogOpen}
          onOpenChange={setEditDialogOpen}
          environment={selectedEnvironment}
          onSuccess={handleCreateSuccess}
          currentUser={currentUser}
        />
      )}

      {deleteEnvironmentName && (
        <DeleteEnvironmentConfirm
          open={deleteConfirmOpen}
          onOpenChange={setDeleteConfirmOpen}
          environmentName={deleteEnvironmentName}
          onSuccess={handleDeleteSuccess}
          currentUser={currentUser}
        />
      )}
    </div>
  )
}