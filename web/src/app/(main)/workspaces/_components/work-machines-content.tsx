'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { WorkMachineMetrics } from './work-machine-metrics'
import { PinnedResources } from '../../environments/_components/pinned-resources'
import { WorkMachineControls } from './work-machine-controls'
import { WorkMachineSetup } from './work-machine-setup'
import {
  updateMyWorkMachine,
  startMyWorkMachine,
  stopMyWorkMachine,
} from '@/app/actions/work-machine.actions'
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
    running: 'text-success',
    stopped: 'text-muted-foreground',
    starting: 'text-info',
    stopping: 'text-warning',
    disabled: 'text-destructive',
    errored: 'text-destructive',
  }

  const stateLabels: Record<string, string> = {
    running: 'Running',
    stopped: 'Stopped',
    starting: 'Starting',
    stopping: 'Stopping',
    disabled: 'Disabled',
    errored: 'Error',
  }

  return {
    color: stateColors[currentState] || 'text-muted-foreground',
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
  pinnedEnvironments,
}: WorkMachinesContentProps) {
  const router = useRouter()
  const [_isPending, startTransition] = useTransition()
  const [workMachines, setWorkMachines] = useState(initialMachines)
  const [selectedMachineId, _setSelectedMachineId] = useState(workMachines[0]?.id)
  const [isLoading, setIsLoading] = useState(false)

  const selectedMachine = workMachines.find((m) => m.id === selectedMachineId) || workMachines[0]

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
    return <WorkMachineSetup availableMachineTypes={availableMachineTypes} />
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
          <p className="text-muted-foreground mt-1.5 text-sm">
            Monitor and manage your development environment
          </p>
        </div>

        {/* Machine Info Card */}
        <div className="bg-card mb-6 border p-6">
          <div className="mb-6 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="bg-primary flex h-10 w-10 items-center justify-center">
                <Server className="text-primary-foreground h-5 w-5" />
              </div>
              <div>
                <h2 className="text-base font-semibold">{selectedMachine.name}</h2>
                <p className="text-muted-foreground text-xs">{selectedMachine.type}</p>
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
          <div className="grid grid-cols-4 gap-6 border-t pt-6">
            <div>
              <p className="text-muted-foreground text-xs font-medium tracking-wider uppercase">
                Status
              </p>
              <div className="mt-2 flex items-center gap-2">
                {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState)
                  .isTransitioning && <Loader2 className="text-info h-4 w-4 animate-spin" />}
                <p
                  className={`text-sm font-medium ${getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).color}`}
                >
                  {
                    getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState)
                      .label
                  }
                </p>
              </div>
            </div>
            <div>
              <p className="text-muted-foreground text-xs font-medium tracking-wider uppercase">
                {selectedMachine.currentState === 'running' ? 'Uptime' : 'Status'}
              </p>
              <p className="mt-2 text-sm font-medium">
                {selectedMachine.currentState === 'running'
                  ? selectedMachine.uptime
                  : selectedMachine.currentState === 'stopped'
                    ? 'Not consuming resources'
                    : selectedMachine.currentState === 'starting'
                      ? 'Starting up...'
                      : selectedMachine.currentState === 'stopping'
                        ? 'Shutting down...'
                        : 'Transitioning...'}
              </p>
            </div>
            <div>
              <p className="text-muted-foreground text-xs font-medium tracking-wider uppercase">
                Owner
              </p>
              <p className="mt-2 text-sm font-medium">{selectedMachine.owner.split('@')[0]}</p>
            </div>
            <div>
              <p className="text-muted-foreground text-xs font-medium tracking-wider uppercase">
                Type
              </p>
              <p className="mt-2 text-sm font-medium">{selectedMachine.type}</p>
            </div>
          </div>
        </div>



        {/* Metrics Section - Only show when machine is running */}
        {selectedMachine.currentState === 'running' && (
          <div className="mb-6">
            <h2 className="mb-4 text-base font-semibold">Resource Usage</h2>
            <WorkMachineMetrics machineState={selectedMachine.currentState} />
          </div>
        )}

        {/* Pinned Resources Section */}
        <div>
          <h2 className="mb-4 text-base font-semibold">Quick Access</h2>
          <PinnedResources workspaces={pinnedWorkspaces} environments={pinnedEnvironments} />
        </div>
      </div>
    </main>
  )
}
