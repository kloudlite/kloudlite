'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@kloudlite/ui'
import {
  Plus,
  MoreHorizontal,
  ExternalLink,
  Power,
  PowerOff,
  Edit,
  Loader2,
  Copy,
  Pin,
  PinOff,
} from 'lucide-react'
import { VisibilityBadge } from '@/components/visibility-selector'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@kloudlite/ui'
import { CreateEnvironmentDialog } from '@/components/dialogs/create-environment'
import { EditEnvironmentDialog } from '@/components/dialogs/edit-environment'
import { DeleteEnvironmentConfirm } from '@/components/dialogs/delete-environment-confirm'
import { ForkEnvironmentSheet } from './fork-environment-sheet'
import { ImportEnvironmentDialog } from '@/components/dialogs/import-environment'
import { activateEnvironment, deactivateEnvironment, exportEnvironmentConfig } from '@/app/actions/environment.actions'
import { pinEnvironment, unpinEnvironment } from '@/app/actions/user-preferences.actions'
import { Download, Upload } from 'lucide-react'
import { toast } from 'sonner'
import type { EnvironmentUIModel } from '@kloudlite/types'

interface EnvironmentsListProps {
  environments: EnvironmentUIModel[]
  currentUser: string
  workMachineRunning?: boolean
  pinnedEnvironmentIds?: string[]
}

// Format forking phase to user-friendly text
function formatForkingPhase(phase: string | undefined): string {
  if (!phase) return 'Preparing...'

  const phaseMap: Record<string, string> = {
    'Pending': 'Preparing...',
    'Suspending': 'Pausing source environment...',
    'ForkingResources': 'Copying configurations...',
    'ForkingPVCs': 'Creating volumes...',
    'CreatingCopyJobs': 'Starting data transfer...',
    'WaitingForCopyCompletion': 'Copying data...',
    'VerifyingCopies': 'Verifying data...',
    'ForkingCompositions': 'Forking services...',
    'Resuming': 'Resuming source...',
    'Completed': 'Completed',
    'Failed': 'Failed',
  }

  return phaseMap[phase] || phase
}


// Format backend error messages into user-friendly text
function formatErrorMessage(error: string): string {
  if (!error) return 'An error occurred'

  // Handle WorkMachine stopped state error
  const workMachineStoppedMatch = error.match(/WorkMachine '([^']+)' is in '([^']+)' state/)
  if (workMachineStoppedMatch) {
    const state = workMachineStoppedMatch[2]
    if (state === 'stopped') {
      return 'Your workspace is stopped. Please start your workspace first before activating environments.'
    }
    return `Your workspace is in '${state}' state. Please wait for it to be ready.`
  }

  // Handle admission webhook errors - extract the meaningful part
  if (error.includes('admission webhook')) {
    const reasonMatch = error.match(/denied the request: (.+)/)
    if (reasonMatch) {
      return formatErrorMessage(reasonMatch[1])
    }
  }

  // Handle "cannot activate environment" prefix
  if (error.startsWith('cannot activate environment:')) {
    return formatErrorMessage(error.replace('cannot activate environment:', '').trim())
  }

  return error
}

export function EnvironmentsList({
  environments: initialEnvironments,
  currentUser,
  workMachineRunning = false,
  pinnedEnvironmentIds = [],
}: EnvironmentsListProps) {
  const pinnedSet = new Set(pinnedEnvironmentIds)
  const [scopeFilter, setScope] = useState<'all' | 'mine'>('all')
  const [statusFilter, setStatus] = useState<'all' | 'active'>('all')
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [forkDialogOpen, setForkDialogOpen] = useState(false)
  const [forkSourceEnvironment, setForkSourceEnvironment] = useState<string | null>(null)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [importDialogOpen, setImportDialogOpen] = useState(false)
  const [selectedEnvironment, setSelectedEnvironment] = useState<EnvironmentUIModel | null>(null)
  const [deleteEnvironmentId, setDeleteEnvironmentId] = useState<string | null>(null)
  const [deleteEnvironmentDisplayName, setDeleteEnvironmentDisplayName] = useState<string | null>(null)
  const [, startTransition] = useTransition()

  const handlePin = async (envName: string) => {
    const result = await pinEnvironment(envName)
    if (result.success) {
      toast.success('Environment pinned to dashboard')
      router.refresh()
    } else {
      toast.error('Failed to pin environment', { description: result.error })
    }
  }

  const handleUnpin = async (envName: string) => {
    const result = await unpinEnvironment(envName)
    if (result.success) {
      toast.success('Environment unpinned from dashboard')
      router.refresh()
    } else {
      toast.error('Failed to unpin environment', { description: result.error })
    }
  }
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
          env.status === 'snapping' ||
          env.status === 'forking',
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
        toast.success('Environment activated')
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error('Failed to activate environment', {
          description: formatErrorMessage(result.error || 'An error occurred'),
        })
      }
    } catch (error: unknown) {
      toast.error('Failed to activate environment', {
        description: formatErrorMessage(error instanceof Error ? error.message : 'An error occurred'),
      })
    }
  }

  const handleDeactivate = async (envName: string) => {
    try {
      const result = await deactivateEnvironment(envName)
      if (result.success) {
        toast.success('Environment deactivated')
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error('Failed to deactivate environment', {
          description: formatErrorMessage(result.error || 'An error occurred'),
        })
      }
    } catch (error: unknown) {
      toast.error('Failed to deactivate environment', {
        description: formatErrorMessage(error instanceof Error ? error.message : 'An error occurred'),
      })
    }
  }

  const handleEditClick = (env: EnvironmentUIModel) => {
    setSelectedEnvironment(env)
    setEditDialogOpen(true)
  }

  const handleDeleteClick = (envId: string, envDisplayName: string) => {
    setDeleteEnvironmentId(envId)
    setDeleteEnvironmentDisplayName(envDisplayName)
    setDeleteConfirmOpen(true)
  }

  const handleForkClick = (envName: string) => {
    setForkSourceEnvironment(envName)
    setForkDialogOpen(true)
  }

  const handleCreateSuccess = () => {
    toast.success('Environment created', {
      description: 'Your new environment has been created successfully.',
    })
    startTransition(() => {
      router.refresh()
    })
  }

  const handleForkSuccess = () => {
    toast.success('Environment forked', {
      description: 'The environment has been forked successfully.',
    })
    startTransition(() => {
      router.refresh()
    })
  }

  const handleExportConfig = async (env: EnvironmentUIModel) => {
    toast.loading('Exporting environment config...')
    const result = await exportEnvironmentConfig(env.name, env.targetNamespace || '')
    toast.dismiss()

    if (result.success && result.data) {
      // Convert to YAML-like format (JSON for now, can be converted to YAML with a library)
      const jsonString = JSON.stringify(result.data, null, 2)
      const blob = new Blob([jsonString], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${env.name}-config.json`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
      toast.success('Environment config exported')
    } else {
      toast.error('Failed to export config', {
        description: result.error || 'An error occurred',
      })
    }
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
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            className="gap-2"
            onClick={() => setImportDialogOpen(true)}
            disabled={!workMachineRunning}
            title={!workMachineRunning ? 'Start your WorkMachine first' : undefined}
          >
            <Upload className="h-4 w-4" />
            Import
          </Button>
          <Button
            size="sm"
            className="gap-2"
            onClick={() => setCreateDialogOpen(true)}
            disabled={!workMachineRunning}
            title={!workMachineRunning ? 'Start your WorkMachine first' : undefined}
          >
            <Plus className="h-4 w-4" />
            {workMachineRunning ? 'New Environment' : 'New Environment (WorkMachine stopped)'}
          </Button>
        </div>
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
              <th className="text-muted-foreground px-6 py-3 text-center text-xs font-medium tracking-wider uppercase">
                Visibility
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
                    {env.owner}/{env.name || env.id || 'unnamed'}
                    <ExternalLink className="h-3 w-3" />
                  </Link>
                </td>
                <td className="px-6 py-4 text-sm whitespace-nowrap">
                  {env.owner.includes('@') ? env.owner.split('@')[0] : env.owner}
                </td>
                <td className="px-6 py-4 text-center whitespace-nowrap">
                  <VisibilityBadge visibility={env.spec?.visibility} />
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center gap-2">
                    <span
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium shrink-0 ${
                        env.status === 'active'
                          ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                          : env.status === 'inactive'
                            ? 'bg-secondary text-secondary-foreground'
                            : env.status === 'activating'
                              ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
                              : env.status === 'deactivating'
                                ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
                                : env.status === 'snapping'
                                  ? 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400'
                                  : env.status === 'deleting'
                                    ? 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                                    : env.status === 'error'
                                      ? 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                                      : env.status === 'forking'
                                        ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
                                        : 'bg-secondary text-secondary-foreground'
                      }`}
                    >
                      {(env.status === 'deleting' ||
                        env.status === 'activating' ||
                        env.status === 'deactivating' ||
                        env.status === 'snapping' ||
                        env.status === 'forking') && <Loader2 className="h-3 w-3 animate-spin" />}
                      {env.status}
                    </span>
                    {/* Show forking progress inline */}
                    {env.status === 'forking' && env.forkingStatus && (
                      <span className="text-muted-foreground text-xs">
                        {env.sourceForkingStatus ? (
                          <span className="italic">
                            → {env.sourceForkingStatus.targetEnvironmentName}
                          </span>
                        ) : (
                          <span className="flex items-center gap-2">
                            <span>{formatForkingPhase(env.forkingStatus.phase)}</span>
                            {env.forkingStatus.totalPVCs && env.forkingStatus.totalPVCs > 0 && (
                              <>
                                <span className="text-muted-foreground/50">•</span>
                                <span>{env.forkingStatus.forkedPVCs || 0}/{env.forkingStatus.totalPVCs} volumes</span>
                              </>
                            )}
                          </span>
                        )}
                      </span>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 text-right text-sm whitespace-nowrap">
                  {env.status === 'deleting' || env.status === 'forking' || env.status === 'snapping' ? (
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0" disabled>
                      <MoreHorizontal className="h-4 w-4 opacity-30" />
                    </Button>
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
                        {pinnedSet.has(env.name) ? (
                          <DropdownMenuItem onClick={() => handleUnpin(env.name)}>
                            <PinOff className="mr-2 h-4 w-4" />
                            Unpin from Dashboard
                          </DropdownMenuItem>
                        ) : (
                          <DropdownMenuItem onClick={() => handlePin(env.name)}>
                            <Pin className="mr-2 h-4 w-4" />
                            Pin to Dashboard
                          </DropdownMenuItem>
                        )}
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
                            className={workMachineRunning ? "text-green-600" : "text-muted-foreground cursor-not-allowed"}
                            disabled={!workMachineRunning}
                          >
                            <Power className="mr-2 h-4 w-4" />
                            {workMachineRunning ? 'Activate' : 'Activate (WorkMachine stopped)'}
                          </DropdownMenuItem>
                        )}
                        <DropdownMenuItem onClick={() => handleForkClick(env.name)}>
                          <Copy className="mr-2 h-4 w-4" />
                          Create from Snapshot
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => handleExportConfig(env)}>
                          <Download className="mr-2 h-4 w-4" />
                          Export Config
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-red-600"
                          onClick={() => handleDeleteClick(env.id, env.name)}
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

      {deleteEnvironmentId && (
        <DeleteEnvironmentConfirm
          open={deleteConfirmOpen}
          onOpenChange={setDeleteConfirmOpen}
          environmentId={deleteEnvironmentId}
          displayName={deleteEnvironmentDisplayName || undefined}
          onSuccess={handleDeleteSuccess}
          currentUser={currentUser}
        />
      )}

      {forkSourceEnvironment && (
        <ForkEnvironmentSheet
          open={forkDialogOpen}
          onOpenChange={setForkDialogOpen}
          onSuccess={handleForkSuccess}
          sourceEnvironment={forkSourceEnvironment}
        />
      )}

      <ImportEnvironmentDialog
        open={importDialogOpen}
        onOpenChange={setImportDialogOpen}
        onSuccess={handleCreateSuccess}
        currentUser={currentUser}
      />
    </div>
  )
}
