'use client'

import { useState } from 'react'
import {
  Play,
  Square,
  Settings,
  ChevronDown,
  Cpu,
  MemoryStick,
  HardDrive,
  Zap,
  AlertCircle,
  Check
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface MachineType {
  id: string
  name: string
  cpu: number
  memory: number
  disk: number
  description: string
}

interface WorkMachineControlsProps {
  machineId: string
  machineName: string
  status: 'active' | 'idle' | 'stopped'
  currentType?: string
  onStart?: () => void
  onStop?: () => void
  onTypeChange?: (typeId: string) => void
}

const machineTypes: MachineType[] = [
  {
    id: 'basic',
    name: 'Basic',
    cpu: 2,
    memory: 4,
    disk: 100,
    description: 'For light development work'
  },
  {
    id: 'standard',
    name: 'Standard',
    cpu: 4,
    memory: 8,
    disk: 250,
    description: 'For regular development'
  },
  {
    id: 'performance',
    name: 'Performance',
    cpu: 8,
    memory: 16,
    disk: 500,
    description: 'For heavy workloads'
  },
  {
    id: 'premium',
    name: 'Premium',
    cpu: 16,
    memory: 32,
    disk: 1000,
    description: 'For enterprise workloads'
  }
]

export function WorkMachineControls({
  machineId,
  machineName,
  status,
  currentType = 'standard',
  onStart,
  onStop,
  onTypeChange
}: WorkMachineControlsProps) {
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [selectedType, setSelectedType] = useState(currentType)
  const [autoStop, setAutoStop] = useState(true)
  const [idleTimeout, setIdleTimeout] = useState('30')
  const [startupScript, setStartupScript] = useState('')
  const [showTypeChangeWarning, setShowTypeChangeWarning] = useState(false)
  const [isChangingType, setIsChangingType] = useState(false)

  const currentMachineType = machineTypes.find(t => t.id === selectedType) || machineTypes[1]

  const handleStart = () => {
    if (onStart) onStart()
  }

  const handleStop = () => {
    if (onStop) onStop()
  }

  const handleTypeSelect = (typeId: string) => {
    if (typeId !== currentType && status === 'active') {
      setSelectedType(typeId)
      setShowTypeChangeWarning(true)
    } else {
      setSelectedType(typeId)
      if (onTypeChange) onTypeChange(typeId)
    }
  }

  const confirmTypeChange = () => {
    setIsChangingType(true)
    // Simulate type change
    setTimeout(() => {
      if (onTypeChange) onTypeChange(selectedType)
      setShowTypeChangeWarning(false)
      setIsChangingType(false)
    }, 2000)
  }

  const saveSettings = () => {
    // Save settings logic here
    setSettingsOpen(false)
  }

  return (
    <>
      <div className="flex items-center gap-2">
        {/* Start/Stop Button */}
        {status === 'stopped' ? (
          <Button
            onClick={handleStart}
            size="sm"
            className="gap-2"
            variant="default"
          >
            <Play className="h-4 w-4" />
            Start Machine
          </Button>
        ) : (
          <Button
            onClick={handleStop}
            size="sm"
            className="gap-2"
            variant={status === 'active' ? 'destructive' : 'outline'}
            disabled={status === 'idle'}
          >
            <Square className="h-4 w-4" />
            Stop Machine
          </Button>
        )}

        {/* Machine Type Selector */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm" className="gap-2">
              <Zap className="h-4 w-4" />
              {currentMachineType.name}
              <ChevronDown className="h-3 w-3" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-80">
            <DropdownMenuLabel>Machine Type</DropdownMenuLabel>
            <DropdownMenuSeparator />
            {machineTypes.map((type) => (
              <DropdownMenuItem
                key={type.id}
                className="flex items-start gap-3 p-3"
                onClick={() => handleTypeSelect(type.id)}
              >
                <div className="mt-0.5">
                  {selectedType === type.id && <Check className="h-4 w-4 text-green-600" />}
                </div>
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <span className="font-medium">{type.name}</span>
                  </div>
                  <div className="text-xs text-gray-500 mt-1">{type.description}</div>
                  <div className="flex items-center gap-4 text-xs text-gray-600 mt-2">
                    <span className="flex items-center gap-1">
                      <Cpu className="h-3 w-3" />
                      {type.cpu} vCPU
                    </span>
                    <span className="flex items-center gap-1">
                      <MemoryStick className="h-3 w-3" />
                      {type.memory} GB
                    </span>
                    <span className="flex items-center gap-1">
                      <HardDrive className="h-3 w-3" />
                      {type.disk} GB
                    </span>
                  </div>
                </div>
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Settings Button */}
        <Button
          variant="outline"
          size="sm"
          onClick={() => setSettingsOpen(true)}
        >
          <Settings className="h-4 w-4" />
        </Button>
      </div>

      {/* Type Change Warning Dialog */}
      <Dialog open={showTypeChangeWarning} onOpenChange={setShowTypeChangeWarning}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Change Machine Type</DialogTitle>
            <DialogDescription>
              Changing the machine type requires stopping and restarting your work machine.
              All running workspaces will be temporarily suspended.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Your work machine will be unavailable for approximately 2-3 minutes during the resize operation.
              </AlertDescription>
            </Alert>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setShowTypeChangeWarning(false)
                setSelectedType(currentType)
              }}
              disabled={isChangingType}
            >
              Cancel
            </Button>
            <Button
              onClick={confirmTypeChange}
              disabled={isChangingType}
            >
              {isChangingType ? 'Changing...' : 'Proceed with Change'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Settings Dialog */}
      <Dialog open={settingsOpen} onOpenChange={setSettingsOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle className="text-xl font-semibold">Work Machine Settings</DialogTitle>
            <DialogDescription className="text-sm text-gray-600">
              Configure automatic behaviors and startup scripts for {machineName}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-8 py-6">
            {/* Auto-stop Settings */}
            <div className="space-y-4">
              <div className="flex items-start justify-between">
                <div className="space-y-1">
                  <Label htmlFor="auto-stop" className="text-base font-medium">
                    Auto-stop when idle
                  </Label>
                  <p className="text-sm text-gray-600">
                    Automatically stop the machine after a period of inactivity
                  </p>
                </div>
                <Switch
                  id="auto-stop"
                  checked={autoStop}
                  onCheckedChange={setAutoStop}
                  className="mt-1"
                />
              </div>

              {autoStop && (
                <div className="ml-0 space-y-2">
                  <Label htmlFor="idle-timeout" className="text-sm font-medium">
                    Idle timeout (minutes)
                  </Label>
                  <Input
                    id="idle-timeout"
                    type="number"
                    min="5"
                    max="120"
                    value={idleTimeout}
                    onChange={(e) => setIdleTimeout(e.target.value)}
                    className="w-32"
                  />
                </div>
              )}
            </div>

            {/* Startup Script */}
            <div className="space-y-2">
              <Label htmlFor="startup-script" className="text-base font-medium">
                Startup script (optional)
              </Label>
              <textarea
                id="startup-script"
                className="w-full min-h-[120px] rounded-md border border-gray-300 px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-gray-900 resize-none"
                placeholder="#!/bin/bash
# Commands to run on machine startup"
                value={startupScript}
                onChange={(e) => setStartupScript(e.target.value)}
              />
              <p className="text-sm text-gray-600">
                This script runs automatically when the machine starts
              </p>
            </div>

            {/* SSH Keys */}
            <div className="space-y-3">
              <div>
                <Label className="text-base font-medium">SSH Public Keys</Label>
                <p className="text-sm text-gray-600 mt-1">
                  Add SSH keys for direct terminal access
                </p>
              </div>
              <Button variant="outline" className="w-fit">
                Manage SSH Keys
              </Button>
            </div>
          </div>

          <DialogFooter className="gap-2">
            <Button
              variant="outline"
              onClick={() => setSettingsOpen(false)}
              className="px-6"
            >
              Cancel
            </Button>
            <Button
              onClick={saveSettings}
              className="px-6 bg-gray-900 hover:bg-gray-800"
            >
              Save Settings
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}