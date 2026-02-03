'use client'

import { useState, useTransition, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { WorkMachineControls } from './work-machine-controls'
import {
  startMyWorkMachine,
  stopMyWorkMachine,
} from '@/app/actions/work-machine.actions'
import { toast } from 'sonner'
import { Server, Loader2, Activity, Clock, User, Cpu } from 'lucide-react'

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
  autoShutdown?: {
    enabled: boolean
    idleThresholdMinutes: number
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

interface WorkMachineCardProps {
  machine: WorkMachine
  availableMachineTypes: MachineType[]
}

// Helper to get state display info
function getStateDisplay(currentState: string, desiredState: string) {
  const isTransitioning = currentState !== desiredState

  const stateColors: Record<string, string> = {
    running: 'text-emerald-600 dark:text-emerald-400',
    stopped: 'text-muted-foreground',
    starting: 'text-blue-600 dark:text-blue-400',
    stopping: 'text-amber-600 dark:text-amber-400',
    disabled: 'text-destructive',
    errored: 'text-destructive',
  }

  const stateBgColors: Record<string, string> = {
    running: 'bg-emerald-500/10',
    stopped: 'bg-muted',
    starting: 'bg-blue-500/10',
    stopping: 'bg-amber-500/10',
    disabled: 'bg-destructive/10',
    errored: 'bg-destructive/10',
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
    bgColor: stateBgColors[currentState] || 'bg-muted',
    label: stateLabels[currentState] || currentState,
    isTransitioning,
    desiredLabel: stateLabels[desiredState] || desiredState,
  }
}

export function WorkMachineCard({ machine, availableMachineTypes }: WorkMachineCardProps) {
  const router = useRouter()
  const [_isPending, startTransition] = useTransition()
  const [isLoading, setIsLoading] = useState(false)
  const [optimisticState, setOptimisticState] = useState<string | null>(null)

  // Determine the current display state (optimistic or actual)
  const displayState = optimisticState || machine.currentState
  const isTransitioning = displayState !== machine.desiredState || optimisticState !== null

  // Clear optimistic state when actual state matches desired state
  useEffect(() => {
    if (optimisticState && machine.currentState === machine.desiredState) {
      setOptimisticState(null)
    }
  }, [machine.currentState, machine.desiredState, optimisticState])

  // Auto-refresh when machine is transitioning
  useEffect(() => {
    if (isTransitioning) {
      // Poll every 1 second during transitions
      const interval = setInterval(() => {
        router.refresh()
      }, 1000)

      return () => clearInterval(interval)
    }
  }, [machine.currentState, machine.desiredState, router, isTransitioning])

  const handleStart = async () => {
    // Optimistically update UI immediately
    setOptimisticState('starting')
    setIsLoading(true)

    try {
      const result = await startMyWorkMachine()
      if (result.success) {
        toast.success('Work machine starting')
        // Trigger immediate refresh to get updated state
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to start work machine')
        // Revert optimistic state on error
        setOptimisticState(null)
      }
    } catch (_error) {
      toast.error('An error occurred')
      // Revert optimistic state on error
      setOptimisticState(null)
    } finally {
      setIsLoading(false)
    }
  }

  const handleStop = async () => {
    // Optimistically update UI immediately
    setOptimisticState('stopping')
    setIsLoading(true)

    try {
      const result = await stopMyWorkMachine()
      if (result.success) {
        toast.success('Work machine stopping')
        // Trigger immediate refresh to get updated state
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to stop work machine')
        // Revert optimistic state on error
        setOptimisticState(null)
      }
    } catch (_error) {
      toast.error('An error occurred')
      // Revert optimistic state on error
      setOptimisticState(null)
    } finally {
      setIsLoading(false)
    }
  }

  const handleTypeChange = async (typeId: string) => {
    // Type change not implemented yet
    toast.info('Machine type change coming soon')
  }

  const stateDisplay = getStateDisplay(displayState, machine.desiredState)

  return (
    <div className="group relative overflow-hidden rounded-xl border bg-card transition-all duration-300 mb-8">
      {/* Gradient decoration */}
      <div className="absolute top-0 right-0 h-32 w-32 bg-gradient-to-br from-primary/20 to-primary/5 blur-3xl opacity-0 transition-opacity duration-300 group-hover:opacity-100" />

      <div className="relative p-6 sm:p-8 space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="flex items-center gap-4">
            <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-gradient-to-br from-primary/20 to-primary/5 border border-primary/20">
              <Server className="h-7 w-7 text-primary" />
            </div>
            <div>
              <h2 className="text-xl font-semibold tracking-tight">
                {machine.owner}&apos;s WorkMachine
              </h2>
              <p className="text-muted-foreground text-sm mt-0.5">{machine.type}</p>
            </div>
          </div>

          {/* Machine Controls */}
          <WorkMachineControls
            machineId={machine.id}
            machineName={machine.name}
            status={machine.status}
            currentState={displayState}
            desiredState={machine.desiredState}
            currentType={machine.type}
            availableMachineTypes={availableMachineTypes}
            sshPublicKey={machine.sshPublicKey}
            sshAuthorizedKeys={machine.sshAuthorizedKeys}
            autoShutdown={machine.autoShutdown}
            onStart={handleStart}
            onStop={handleStop}
            onTypeChange={handleTypeChange}
            isLoading={isLoading}
          />
        </div>

        {/* Machine Stats Grid */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 pt-6 border-t">
          {/* Status */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-muted-foreground" />
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Status
              </p>
            </div>
            <div className="flex items-center gap-2">
              {(displayState === 'starting' || displayState === 'stopping') && (
                <Loader2 className={`h-4 w-4 animate-spin ${stateDisplay.color}`} />
              )}
              <span className={`inline-flex items-center gap-2 px-3 py-1 rounded-lg ${stateDisplay.bgColor} ${stateDisplay.color} text-sm font-semibold`}>
                {stateDisplay.label}
              </span>
            </div>
          </div>

          {/* Uptime/Status Message */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-muted-foreground" />
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                {displayState === 'running' ? 'Uptime' : 'Status'}
              </p>
            </div>
            <p className="text-sm font-semibold">
              {displayState === 'running'
                ? machine.uptime
                : displayState === 'stopped'
                  ? 'Not consuming resources'
                  : displayState === 'starting'
                    ? 'Starting up...'
                    : displayState === 'stopping'
                      ? 'Shutting down...'
                      : 'Transitioning...'}
            </p>
          </div>

          {/* Owner */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <User className="h-4 w-4 text-muted-foreground" />
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Owner
              </p>
            </div>
            <p className="text-sm font-semibold">{machine.owner}</p>
          </div>

          {/* Type */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Cpu className="h-4 w-4 text-muted-foreground" />
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Type
              </p>
            </div>
            <p className="text-sm font-semibold">{machine.type}</p>
          </div>
        </div>
      </div>
    </div>
  )
}
