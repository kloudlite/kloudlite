'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { WorkMachineMetrics } from '@/components/work-machine-metrics'
import { PinnedResources } from '@/components/pinned-resources'
import { WorkMachineControls } from '@/components/work-machine-controls'
import { updateMyWorkMachine, startMyWorkMachine, stopMyWorkMachine } from '@/app/actions/work-machine.actions'
import { toast } from 'sonner'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { ChevronDown, Server, Loader2, ArrowRight } from 'lucide-react'

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

interface WorkMachinesContentProps {
  initialMachines: WorkMachine[]
  currentUser: string
  isAdmin: boolean
  availableMachineTypes: any[]
  pinnedWorkspaces: any[]
  pinnedEnvironments: any[]
}

export function WorkMachinesContent({
  initialMachines,
  currentUser,
  isAdmin,
  availableMachineTypes,
  pinnedWorkspaces,
  pinnedEnvironments
}: WorkMachinesContentProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [workMachines, setWorkMachines] = useState(initialMachines)
  const [selectedMachineId, setSelectedMachineId] = useState(workMachines[0]?.id)
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
  }, [selectedMachine?.currentState, selectedMachine?.desiredState, router])

  // Handle case where user has no work machine
  if (!selectedMachine) {
    return (
      <main className="mx-auto max-w-7xl px-6 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-light tracking-tight">Welcome back!</h1>
          <p className="text-sm text-gray-600 mt-2">
            Monitor your development machine and manage resources
          </p>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 p-12 text-center">
          <Server className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No work machine found</h3>
          <p className="text-sm text-gray-500 mb-4">
            Your work machine is being created. Please refresh the page in a moment.
          </p>
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
    } catch (error) {
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
    } catch (error) {
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
    } catch (error) {
      toast.error('An error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
      {/* Page Header */}
      <div className="mb-8">
        <div>
          <h1 className="text-3xl font-light tracking-tight">Welcome back!</h1>
          <p className="text-sm text-gray-600 mt-2">
            Monitor your development machine and manage resources
          </p>
        </div>
      </div>

      {/* Machine Info Bar with Controls */}
      <div className="bg-white rounded-lg border border-gray-200 p-4 mb-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-6">
            <div>
              <p className="text-xs text-gray-500">Machine</p>
              <p className="text-sm font-medium">{selectedMachine.name}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500">Owner</p>
              <p className="text-sm font-medium">{selectedMachine.owner.split('@')[0]}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500">State</p>
              <div className="flex items-center gap-2">
                {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).isTransitioning && (
                  <Loader2 className="h-3 w-3 animate-spin text-blue-600" />
                )}
                <p className={`text-sm font-medium ${getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).color}`}>
                  {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).label}
                </p>
                {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).isTransitioning && (
                  <>
                    <ArrowRight className="h-3 w-3 text-gray-400" />
                    <span className="text-sm font-medium text-gray-600">
                      {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).desiredLabel}
                    </span>
                  </>
                )}
              </div>
            </div>
            <div>
              <p className="text-xs text-gray-500">Uptime</p>
              <p className="text-sm font-medium">{selectedMachine.uptime}</p>
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
            onStart={handleStart}
            onStop={handleStop}
            onTypeChange={handleTypeChange}
            isLoading={isLoading}
          />
        </div>
      </div>

      {/* Transitioning State Banner */}
      {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).isTransitioning && (
        <div className="mb-6 bg-blue-50 rounded-lg border border-blue-200 p-4">
          <div className="flex items-center gap-3">
            <Loader2 className="h-5 w-5 animate-spin text-blue-600" />
            <div>
              <p className="text-sm font-medium text-blue-900">
                Machine is transitioning from{' '}
                <span className="font-semibold">
                  {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).label}
                </span>
                {' '}to{' '}
                <span className="font-semibold">
                  {getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).desiredLabel}
                </span>
              </p>
              <p className="text-xs text-blue-700 mt-1">
                This may take a few moments. The page will refresh automatically when complete.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Metrics Section - Always show, but CPU/Memory are 0 when stopped */}
      <div className="mb-8">
        <WorkMachineMetrics
          cpu={selectedMachine.cpu}
          memory={selectedMachine.memory}
          disk={selectedMachine.disk}
        />
      </div>

      {/* Stopped State Message */}
      {selectedMachine.currentState === 'stopped' && !getStateDisplay(selectedMachine.currentState, selectedMachine.desiredState).isTransitioning && (
        <div className="mb-8 bg-yellow-50 rounded-lg border border-yellow-200 p-6 text-center">
          <p className="text-sm text-yellow-800">
            Machine is stopped. CPU and Memory are not consuming resources, but disk storage is preserved.
          </p>
        </div>
      )}

      {/* Pinned Resources Section */}
      <div>
        <h2 className="text-lg font-medium mb-4">Pinned Resources</h2>
        <PinnedResources
          workspaces={pinnedWorkspaces}
          environments={pinnedEnvironments}
        />
      </div>
    </main>
  )
}