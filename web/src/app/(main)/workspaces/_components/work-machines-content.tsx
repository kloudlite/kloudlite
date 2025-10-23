'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { WorkMachineMetrics } from './work-machine-metrics'
import { PinnedResources } from '../../environments/_components/pinned-resources'
import { WorkMachineControls } from './work-machine-controls'
import { updateMyWorkMachine, startMyWorkMachine, stopMyWorkMachine } from '@/app/actions/work-machine.actions'
import { toast } from 'sonner'
import { Server, Loader2, ArrowRight } from 'lucide-react'

interface WorkMachine {
  id: string
  owner: string
  name: string
  status: 'active' | 'idle' | 'stopped'
  currentState: string
  desiredState: string
  cpu: number
  memory: number
  disk: number
  uptime: string
  type: string
  sshPublicKey?: string
  sshAuthorizedKeys?: string[]
}

// Helper to get state display info
function getStateDisplay(currentState: string, desiredState: string) {
  const isTransitioning = currentState !== desiredState

  const stateColors: Record<string, string> = {
    running: 'text-green-600',
    stopped: 'text-gray-600',
    starting: 'text-blue-600',
    stopping: 'text-yellow-600',
    disabled: 'text-red-600',
    error: 'text-red-600',
  }

  const stateLabels: Record<string, string> = {
    running: 'Running',
    stopped: 'Stopped',
    starting: 'Starting',
    stopping: 'Stopping',
    disabled: 'Disabled',
    error: 'Error',
  }

  return {
    color: stateColors[currentState] || 'text-gray-600',
    label: stateLabels[currentState] || currentState,
    isTransitioning,
    desiredLabel: stateLabels[desiredState] || desiredState,
  }
}

interface MachineType {
  id: string
  name: string
  description: string
  category: string
  cpu: string
  memory: string
  gpu?: string
}

interface WorkMachinesContentProps {
  initialMachines: WorkMachine[]
  currentUser: string
  isAdmin: boolean
  availableMachineTypes: MachineType[]
  pinnedWorkspaces: never[]
  pinnedEnvironments: never[]
}

export function WorkMachinesContent({
  initialMachines,
  currentUser: _currentUser,
  isAdmin: _isAdmin,
  availableMachineTypes,
  pinnedWorkspaces,
  pinnedEnvironments
}: WorkMachinesContentProps) {
  const router = useRouter()
  const [_isPending, startTransition] = useTransition()
  const [workMachines, setWorkMachines] = useState(initialMachines)
  const [selectedMachineId, _setSelectedMachineId] = useState(workMachines[0]?.id)
  const [isLoading, setIsLoading] = useState(false)

  const selectedMachine = workMachines.find(m => m.id === selectedMachineId) || workMachines[0]

  // Sync local state with prop changes when page refreshes
  useEffect(() => {
    setWorkMachines(initialMachines)
  }, [initialMachines])

  // Auto-refresh when machine is transitioning
  useEffect(() => {
    if (!selectedMachine) return

    const isTransitioning = selectedMachine.currentState !== selectedMachine.desiredState

    if (isTransitioning) {
      // Poll every 1 second during transitions
      const interval = setInterval(() => {
        router.refresh()
      }, 1000)

      return () => clearInterval(interval)
    }

    return undefined
  }, [selectedMachine, router])

  // Handle case where user has no work machine
  if (!selectedMachine) {
    return (
      <main className="min-h-screen">
        <div className="mx-auto max-w-7xl px-6 py-8">
          <div className="mb-8">
            <h1 className="text-2xl font-semibold">Dashboard</h1>
            <p className="text-sm text-muted-foreground mt-1.5">
              Monitor and manage your development environment
            </p>
          </div>

          <div className="bg-card border p-12 text-center">
            <Server className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-base font-semibold mb-2">No work machine found</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Your work machine is being created. Please refresh the page in a moment.
            </p>
          </div>
        </div>
      </main>
    )
  }

  const handleStart = async () => {
    setIsLoading(true)
    try {
      const result = await startMyWorkMachine()
      if (result.success) {
        toast.success('Work machine starting')
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to start work machine')
      }
    } catch (_error) {
      toast.error('An error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  const handleStop = async () => {
    setIsLoading(true)
    try {
      const result = await stopMyWorkMachine()
      if (result.success) {
        toast.success('Work machine stopping')
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to stop work machine')
      }
    } catch (_error) {
      toast.error('An error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  const handleTypeChange = async (typeId: string) => {
    setIsLoading(true)
    try {
      const result = await updateMyWorkMachine({ machineType: typeId })
      if (result.success) {
        toast.success('Work machine type updated')
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to update work machine')
      }
    } catch (_error) {
      toast.error('An error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <main className="min-h-screen">
      <div className="mx-auto max-w-7xl px-6 py-8">
        {/* Page Header */}
        <div className="mb-8">
          <h1 className="text-2xl font-semibold">Dashboard</h1>
          <p className="text-sm text-muted-foreground mt-1.5">
            Monitor and manage your development environment
          </p>
        </div>

        {/* Machine Info Card */}
        <div className="bg-card border p-6 mb-6">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 bg-primary flex items-center justify-center">
                <Server className="h-5 w-5 text-primary-foreground" />
              </div>
              <div>
                <h2 className="text-base font-semibold">{selectedMachine.name}</h2>
                <p className="text-xs text-muted-foreground">{selectedMachine.type}</p>
              </div>
            </div>

            {/* Machine Controls */}
            <WorkMachineControls
              machineId={selectedMachine.id}
              machineName={selectedMachine.name}
              status={selectedMachine.status}
              currentState={selectedMachine.currentState}
              desiredState={selectedMachine.desiredState}
              currentType={selectedMachine.type}
              availableMachineTypes={availableMachineTypes}
              sshPublicKey={selectedMachine.sshPublicKey}
              sshAuthorizedKeys={selectedMachine.sshAuthorizedKeys}
              onStart={handleStart}
              onStop={handleStop}
              onTypeChange={handleTypeChange}
              isLoading={isLoading}
            />
          </div>

          {/* Machine Stats */}
          <div className="grid grid-cols-4 gap-6 pt-6 border-t">
            <div>
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Status</p>
              <div className="flex items-center gap-2 mt-2">
                {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).isTransitioning && (
                  <Loader2 className="h-4 w-4 animate-spin text-blue-600 dark:text-blue-400" />
                )}
                <p className={`text-sm font-medium ${getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).color}`}>
                  {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).label}
                </p>
                {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).isTransitioning && (
                  <>
                    <ArrowRight className="h-3 w-3 text-muted-foreground" />
                    <span className="text-sm font-medium text-muted-foreground">
                      {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).desiredLabel}
                    </span>
                  </>
                )}
              </div>
            </div>
            <div>
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Uptime</p>
              <p className="text-sm font-medium mt-2">{selectedMachine.uptime}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Owner</p>
              <p className="text-sm font-medium mt-2">{selectedMachine.owner.split('@')[0]}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Type</p>
              <p className="text-sm font-medium mt-2">{selectedMachine.type}</p>
            </div>
          </div>
        </div>

        {/* Transitioning State Banner */}
        {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).isTransitioning && (
          <div className="mb-6 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 p-4">
            <div className="flex items-center gap-3">
              <Loader2 className="h-5 w-5 animate-spin text-blue-600 dark:text-blue-400" />
              <div>
                <p className="text-sm font-medium text-blue-900 dark:text-blue-200">
                  Machine is transitioning from{' '}
                  <span className="font-semibold">
                    {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).label}
                  </span>
                  {' '}to{' '}
                  <span className="font-semibold">
                    {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).desiredLabel}
                  </span>
                </p>
                <p className="text-xs text-blue-700 dark:text-blue-300 mt-1">
                  This may take a few moments. The page will refresh automatically when complete.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Stopped State Message */}
        {selectedMachine.currentState === 'stopped' && !getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).isTransitioning && (
          <div className="mb-6 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 p-4 text-center">
            <p className="text-sm text-yellow-800 dark:text-yellow-200">
              Machine is stopped. CPU and Memory are not consuming resources, but disk storage is preserved.
            </p>
          </div>
        )}

        {/* Metrics Section - Always show, but CPU/Memory are 0 when stopped */}
        <div className="mb-6">
          <h2 className="text-base font-semibold mb-4">Resource Usage</h2>
          <WorkMachineMetrics />
        </div>

        {/* Pinned Resources Section */}
        <div>
          <h2 className="text-base font-semibold mb-4">Quick Access</h2>
          <PinnedResources
            workspaces={pinnedWorkspaces}
            environments={pinnedEnvironments}
          />
        </div>
      </div>
    </main>
  )
}