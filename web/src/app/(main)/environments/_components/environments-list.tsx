'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import {
  Plus,
  MoreHorizontal,
  ExternalLink,
  Power,
  PowerOff,
  Edit,
  Loader2,
  Copy,
} from 'lucide-react'
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
import { CloneEnvironmentDialog } from '@/components/dialogs/clone-environment'
import { activateEnvironment, deactivateEnvironment } from '@/app/actions/environment.actions'
import { toast } from 'sonner'
import type { EnvironmentUIModel } from '@/types/environment'
import { formatResourceName } from '@/lib/utils'

interface EnvironmentsListProps {
  environments: EnvironmentUIModel[]
  currentUser: string
}

export function EnvironmentsList({
  environments: initialEnvironments,
  currentUser,
}: EnvironmentsListProps) {
  const [scopeFilter, setScope] = useState<'all' | 'mine'>('all')
  const [statusFilter, setStatus] = useState<'all' | 'active'>('all')
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [cloneDialogOpen, setCloneDialogOpen] = useState(false)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [selectedEnvironment, setSelectedEnvironment] = useState<EnvironmentUIModel | null>(null)
  const [cloneSourceEnvironment, setCloneSourceEnvironment] = useState<EnvironmentUIModel | null>(
    null,
  )
  const [deleteEnvironmentName, setDeleteEnvironmentName] = useState<string | null>(null)
  const [, startTransition] = useTransition()
  const [environments, setEnvironments] = useState<EnvironmentUIModel[]>(initialEnvironments)
  const router = useRouter()

  // Poll for environment updates every 3 seconds
  useEffect(() => {
    const pollInterval = setInterval(() => {
      // Check if any environment is in a transitional state
      const hasTransitionalEnv = environments.some(
        (env) =>
          env.status === 'deleting' ||
          env.status === 'activating' ||
          env.status === 'deactivating' ||
          env.status === 'cloning',
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
    filteredEnvironments = filteredEnvironments.filter((env) => env.owner === currentUser)
  }

  // Apply status filter
  if (statusFilter === 'active') {
    filteredEnvironments = filteredEnvironments.filter((env) => env.status === 'active')
  }

  const handleActivate = async (envName: string) => {
    try {
      const result = await activateEnvironment(envName)
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
    } catch (error: unknown) {
      toast.error('Failed to activate environment', {
        description: error instanceof Error ? error.message : 'An error occurred',
      })
    }
  }

  const handleDeactivate = async (envName: string) => {
    try {
      const result = await deactivateEnvironment(envName)
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
    } catch (error: unknown) {
      toast.error('Failed to deactivate environment', {
        description: error instanceof Error ? error.message : 'An error occurred',
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

  const handleCloneClick = (env: EnvironmentUIModel) => {
    setCloneSourceEnvironment(env)
    setCloneDialogOpen(true)
  }

  const handleCreateSuccess = () => {
    toast.success('Environment created', {
      description: 'Your new environment has been created successfully.',
    })
    startTransition(() => {
      router.refresh()
    })
  }

  const handleCloneSuccess = () => {
    toast.success('Environment cloned', {
      description: 'The environment has been cloned successfully.',
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
          <div className="bg-muted flex items-center gap-1 rounded-md p-1">
            <button
              onClick={() => setScope('all')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                scopeFilter === 'all'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setScope('mine')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                scopeFilter === 'mine'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Mine
            </button>
          </div>

          {/* Status Filter */}
          <div className="bg-muted flex items-center gap-1 rounded-md p-1">
            <button
              onClick={() => setStatus('all')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'all'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setStatus('active')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'active'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Active
            </button>
          </div>

          <span className="text-muted-foreground text-sm">
            {filteredEnvironments.length}{' '}
            {filteredEnvironments.length === 1 ? 'environment' : 'environments'}
          </span>
        </div>
        <Button size="sm" className="gap-2" onClick={() => setCreateDialogOpen(true)}>
          <Plus className="h-4 w-4" />
          New Environment
        </Button>
      </div>

      {/* Table */}
      <div className="bg-card overflow-hidden rounded-lg border">
        <table className="min-w-full">
          <thead className="bg-muted/50 border-b">
            <tr>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Name
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Owner
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Status
              </th>
              <th className="text-muted-foreground px-6 py-3 text-right text-xs font-medium tracking-wider uppercase">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {filteredEnvironments.map((env) => (
              <tr key={env.id} className="hover:bg-muted/50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <Link
                    href={`/environments/${env.id}`}
                    className="hover:text-primary flex items-center gap-1 text-sm font-semibold"
                  >
                    {formatResourceName(env.name)}
                    <ExternalLink className="h-3 w-3" />
                  </Link>
                </td>
                <td className="px-6 py-4 text-sm whitespace-nowrap">
                  {env.owner.includes('@') ? env.owner.split('@')[0] : env.owner}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex flex-col gap-1">
                    <span
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${
                        env.status === 'active'
                          ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                          : env.status === 'deleting'
                            ? 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                            : env.status === 'activating' ||
                                env.status === 'deactivating' ||
                                env.status === 'cloning'
                              ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
                              : 'bg-secondary text-secondary-foreground'
                      }`}
                    >
                      {(env.status === 'deleting' ||
                        env.status === 'activating' ||
                        env.status === 'deactivating' ||
                        env.status === 'cloning') && <Loader2 className="h-3 w-3 animate-spin" />}
                      {env.status}
                    </span>
                    {/* Show cloning progress */}
                    {env.status === 'cloning' && env.cloningStatus && (
                      <div className="text-muted-foreground text-xs">
                        {env.sourceCloningStatus ? (
                          <span>Source for: {env.sourceCloningStatus.targetEnvironmentName}</span>
                        ) : (
                          <>
                            <div>{env.cloningStatus.phase}</div>
                            {env.cloningStatus.totalPVCs && env.cloningStatus.totalPVCs > 0 && (
                              <div className="flex items-center gap-1">
                                <span>
                                  {env.cloningStatus.clonedPVCs || 0}/{env.cloningStatus.totalPVCs}{' '}
                                  PVCs
                                </span>
                                <div className="bg-muted h-1 w-12 overflow-hidden rounded-full">
                                  <div
                                    className="bg-primary h-full transition-all"
                                    style={{
                                      width: `${((env.cloningStatus.clonedPVCs || 0) / env.cloningStatus.totalPVCs) * 100}%`,
                                    }}
                                  />
                                </div>
                              </div>
                            )}
                          </>
                        )}
                      </div>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 text-right text-sm whitespace-nowrap">
                  {env.status === 'deleting' ? (
                    <span className="text-muted-foreground text-xs">Deleting...</span>
                  ) : env.status === 'cloning' ? (
                    <span className="text-muted-foreground text-xs">Cloning...</span>
                  ) : (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem asChild>
                          <Link href={`/environments/${env.id}`}>View Details</Link>
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => handleEditClick(env)}>
                          <Edit className="mr-2 h-4 w-4" />
                          Edit Settings
                        </DropdownMenuItem>
                        {env.status === 'active' ? (
                          <DropdownMenuItem
                            onClick={() => handleDeactivate(env.name)}
                            className="text-orange-600"
                          >
                            <PowerOff className="mr-2 h-4 w-4" />
                            Deactivate
                          </DropdownMenuItem>
                        ) : (
                          <DropdownMenuItem
                            onClick={() => handleActivate(env.name)}
                            className="text-green-600"
                          >
                            <Power className="mr-2 h-4 w-4" />
                            Activate
                          </DropdownMenuItem>
                        )}
                        <DropdownMenuItem onClick={() => handleCloneClick(env)}>
                          <Copy className="mr-2 h-4 w-4" />
                          Clone Environment
                        </DropdownMenuItem>
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
        <div className="bg-card rounded-lg border py-12 text-center">
          <p className="text-muted-foreground text-sm">
            {scopeFilter === 'mine' && statusFilter === 'active'
              ? "You don't have any active environments"
              : scopeFilter === 'mine'
                ? "You don't have any environments yet"
                : statusFilter === 'active'
                  ? 'No active environments found'
                  : 'No environments found'}
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

      {cloneSourceEnvironment && (
        <CloneEnvironmentDialog
          open={cloneDialogOpen}
          onOpenChange={setCloneDialogOpen}
          sourceEnvironment={cloneSourceEnvironment}
          onSuccess={handleCloneSuccess}
          currentUser={currentUser}
        />
      )}
    </div>
  )
}
