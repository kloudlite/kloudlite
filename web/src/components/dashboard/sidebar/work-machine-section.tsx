'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Monitor, Play, Square, Settings, Cpu, HardDrive, Zap } from 'lucide-react'
import { cn } from '@/lib/utils'

interface WorkMachineStats {
  cpu: string
  memory: string
  uptime: string
}

interface WorkMachineSectionProps {
  className?: string
}

export function WorkMachineSection({ className }: WorkMachineSectionProps) {
  const [machineRunning, setMachineRunning] = useState(true)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  
  const stats: WorkMachineStats = {
    cpu: '45%',
    memory: '2.1GB',
    uptime: '3h 24m'
  }

  const handleMachineToggle = () => {
    setMachineRunning(!machineRunning)
  }

  return (
    <>
      <div className={cn("border-y bg-gradient-to-r from-muted/20 via-muted/30 to-muted/20 py-5 -mx-6 px-6 space-y-4", className)}>
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-background/50 border border-border/50">
              <Monitor className="size-4 text-foreground" />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-foreground">Work Machine</h3>
              <p className="text-xs text-muted-foreground">
                {machineRunning ? 'Running' : 'Stopped'}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-1">
            <Button
              size="sm"
              variant="ghost"
              className="size-8 p-0 hover:bg-background/80"
              onClick={() => setEditDialogOpen(true)}
            >
              <Settings className="size-4" />
            </Button>
            {machineRunning ? (
              <Button
                size="sm"
                variant="ghost"
                className="size-8 p-0 text-destructive hover:bg-destructive/15 hover:text-destructive"
                onClick={handleMachineToggle}
              >
                <Square className="size-4" />
              </Button>
            ) : (
              <Button
                size="sm"
                variant="ghost"
                className="size-8 p-0 text-success hover:bg-success/15 hover:text-success"
                onClick={handleMachineToggle}
              >
                <Play className="size-4" />
              </Button>
            )}
          </div>
        </div>
        
        {/* Stats with collapse animation */}
        <div className={cn(
          "grid grid-cols-3 gap-2 overflow-hidden transition-all duration-200 ease-out",
          machineRunning 
            ? "max-h-24 opacity-100 mt-0" 
            : "max-h-0 opacity-0 -mt-4"
        )}>
          <div className="bg-background/60 backdrop-blur-sm border border-border/40 rounded-lg py-3 px-2 text-center hover:bg-background/80 transition-all duration-200 cursor-default group">
            <div className="flex flex-col items-center gap-1.5">
              <div className="p-1.5 rounded-md bg-primary/10 group-hover:bg-primary/20 transition-colors">
                <Cpu className="size-3.5 text-primary" />
              </div>
              <span className="text-xs font-medium text-muted-foreground">CPU</span>
              <p className="text-sm font-bold text-foreground">{stats.cpu}</p>
            </div>
          </div>
          
          <div className="bg-background/60 backdrop-blur-sm border border-border/40 rounded-lg py-3 px-2 text-center hover:bg-background/80 transition-all duration-200 cursor-default group">
            <div className="flex flex-col items-center gap-1.5">
              <div className="p-1.5 rounded-md bg-purple/10 group-hover:bg-purple/20 transition-colors">
                <HardDrive className="size-3.5 text-purple" />
              </div>
              <span className="text-xs font-medium text-muted-foreground">Memory</span>
              <p className="text-sm font-bold text-foreground">{stats.memory}</p>
            </div>
          </div>
          
          <div className="bg-background/60 backdrop-blur-sm border border-border/40 rounded-lg py-3 px-2 text-center hover:bg-background/80 transition-all duration-200 cursor-default group">
            <div className="flex flex-col items-center gap-1.5">
              <div className="p-1.5 rounded-md bg-warning/10 group-hover:bg-warning/20 transition-colors">
                <Zap className="size-3.5 text-warning" />
              </div>
              <span className="text-xs font-medium text-muted-foreground">Uptime</span>
              <p className="text-sm font-bold text-foreground">{stats.uptime}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Edit Work Machine Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Work Machine Settings</DialogTitle>
            <DialogDescription>
              Configure your work machine resources and settings.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 pt-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">CPU Limit</label>
              <input 
                type="text" 
                className="w-full px-3 py-2 border rounded-md" 
                placeholder="e.g., 2 cores"
                defaultValue="2 cores"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Memory Limit</label>
              <input 
                type="text" 
                className="w-full px-3 py-2 border rounded-md" 
                placeholder="e.g., 4GB"
                defaultValue="4GB"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Storage</label>
              <input 
                type="text" 
                className="w-full px-3 py-2 border rounded-md" 
                placeholder="e.g., 20GB"
                defaultValue="20GB"
              />
            </div>
            <div className="flex justify-end gap-2 pt-4">
              <Button variant="outline" onClick={() => setEditDialogOpen(false)}>
                Cancel
              </Button>
              <Button onClick={() => setEditDialogOpen(false)}>
                Save Changes
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}