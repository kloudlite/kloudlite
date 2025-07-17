'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Monitor, Play, Square, Settings, Cpu, HardDrive, Zap, Loader2, AlertTriangle } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { cva, type VariantProps } from 'class-variance-authority'
import { StatCard } from './stat-card'

type MachineState = 'stopped' | 'starting' | 'running' | 'stopping'

const statusVariants = cva("text-xs", {
  variants: {
    state: {
      running: "text-success",
      stopped: "text-muted-foreground",
      starting: "text-warning",
      stopping: "text-warning"
    }
  }
})

const controlButtonVariants = cva("size-8 p-0", {
  variants: {
    action: {
      start: "text-success hover:bg-success/15 hover:text-success",
      stop: "text-destructive hover:bg-destructive/15 hover:text-destructive",
      loading: "cursor-not-allowed"
    }
  }
})

interface WorkMachineStats {
  cpu: string
  memory: string
  uptime: string
}

interface SimpleWorkMachineProps {
  className?: string
}

export function SimpleWorkMachine({ className }: SimpleWorkMachineProps) {
  const [machineState, setMachineState] = useState<MachineState>('running')
  const [showShutdownDialog, setShowShutdownDialog] = useState(false)
  
  const stats: WorkMachineStats = {
    cpu: '45%',
    memory: '2.1GB',
    uptime: '3h 24m'
  }

  // Mock data for affected resources
  const affectedResources = {
    environments: [
      { name: 'Production', status: 'running' },
      { name: 'Staging', status: 'running' }
    ],
    workspaces: [
      { name: 'Frontend Dev', user: 'John Doe' },
      { name: 'Backend API', user: 'Jane Smith' },
      { name: 'Mobile App', user: 'Bob Wilson' }
    ]
  }

  const handleMachineToggle = () => {
    if (machineState === 'stopped') {
      setMachineState('starting')
      // Simulate startup time
      setTimeout(() => setMachineState('running'), 2000)
    } else if (machineState === 'running') {
      // Show confirmation dialog
      setShowShutdownDialog(true)
    }
  }

  const handleConfirmShutdown = () => {
    setShowShutdownDialog(false)
    setMachineState('stopping')
    // Simulate shutdown time
    setTimeout(() => setMachineState('stopped'), 1500)
  }

  const getStatusText = () => {
    switch (machineState) {
      case 'stopped':
        return 'Stopped'
      case 'starting':
        return 'Starting up...'
      case 'running':
        return 'Running'
      case 'stopping':
        return 'Shutting down...'
    }
  }

  const isTransitioning = machineState === 'starting' || machineState === 'stopping'
  const showStats = machineState === 'running' || machineState === 'stopping'

  return (
    <div className={className}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className={cn(
            "p-2 rounded-lg bg-background transition-colors",
            isTransitioning && "animate-pulse"
          )}>
            <Monitor className="size-4 text-foreground" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-foreground">Work Machine</h3>
            <p className={statusVariants({ state: machineState })}>
              {getStatusText()}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="ghost"
            className="size-8 p-0 hover:bg-background/80"
          >
            <Settings className="size-4" />
          </Button>
          {machineState === 'running' ? (
            <Button
              size="sm"
              variant="ghost"
              className={controlButtonVariants({ action: 'stop' })}
              onClick={handleMachineToggle}
              disabled={isTransitioning}
            >
              <Square className="size-4" />
            </Button>
          ) : machineState === 'stopped' ? (
            <Button
              size="sm"
              variant="ghost"
              className={controlButtonVariants({ action: 'start' })}
              onClick={handleMachineToggle}
              disabled={isTransitioning}
            >
              <Play className="size-4" />
            </Button>
          ) : (
            <Button
              size="sm"
              variant="ghost"
              className={controlButtonVariants({ action: 'loading' })}
              disabled
            >
              <Loader2 className="size-4 animate-spin text-warning" />
            </Button>
          )}
        </div>
      </div>
      
      {/* Stats with collapse animation */}
      <div className={cn(
        "grid grid-cols-3 gap-2 overflow-hidden transition-all duration-300 ease-out",
        showStats 
          ? "max-h-28 opacity-100 mt-4" 
          : "max-h-0 opacity-0 mt-0"
      )}>
        <StatCard type="cpu" icon={Cpu} label="CPU" value={stats.cpu} />
        <StatCard type="memory" icon={HardDrive} label="Memory" value={stats.memory} />
        <StatCard type="uptime" icon={Zap} label="Uptime" value={stats.uptime} />
      </div>

      {/* Shutdown Confirmation Dialog */}
      <Dialog open={showShutdownDialog} onOpenChange={setShowShutdownDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="size-5 text-warning" />
              Shutdown Work Machine?
            </DialogTitle>
            <DialogDescription>
              Shutting down the work machine will stop all running environments and workspaces.
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4 py-4">
            {/* Affected Environments */}
            <div>
              <h4 className="text-sm font-medium mb-2">Affected Environments ({affectedResources.environments.length})</h4>
              <div className="space-y-1">
                {affectedResources.environments.map((env) => (
                  <div key={env.name} className="flex items-center gap-2 text-sm text-muted-foreground">
                    <div className="size-2 rounded-full bg-success" />
                    {env.name}
                  </div>
                ))}
              </div>
            </div>
            
            {/* Affected Workspaces */}
            <div>
              <h4 className="text-sm font-medium mb-2">Active Workspaces ({affectedResources.workspaces.length})</h4>
              <div className="space-y-1">
                {affectedResources.workspaces.map((workspace) => (
                  <div key={workspace.name} className="flex items-center justify-between text-sm text-muted-foreground">
                    <span>{workspace.name}</span>
                    <span className="text-xs">by {workspace.user}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowShutdownDialog(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleConfirmShutdown}>
              Shutdown
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}