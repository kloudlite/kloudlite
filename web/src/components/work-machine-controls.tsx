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
  Check,
  Loader2
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

interface WorkMachineControlsProps {
  machineId: string
  machineName: string
  status: 'active' | 'idle' | 'stopped'
  currentState: string
  desiredState: string
  currentType?: string
  availableMachineTypes: any[]
  onStart?: () => void
  onStop?: () => void
  onTypeChange?: (typeId: string) => void
  isLoading?: boolean
}

// Parse K8s resource strings (e.g., "4", "8Gi" -> numbers)
function parseResourceValue(value?: string): number {
  if (!value) return 0
  const match = value.match(/^(\d+(?:\.\d+)?)/)
  return match ? parseFloat(match[1]) : 0
}

export function WorkMachineControls({
  machineId,
  machineName,
  status,
  currentState,
  desiredState,
  currentType = 'standard',
  availableMachineTypes,
  onStart,
  onStop,
  onTypeChange,
  isLoading = false
}: WorkMachineControlsProps) {
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [selectedType, setSelectedType] = useState(currentType)
  const [autoStop, setAutoStop] = useState(true)
  const [idleTimeout, setIdleTimeout] = useState('30')
  const [startupScript, setStartupScript] = useState('')
  const [showTypeChangeWarning, setShowTypeChangeWarning] = useState(false)
  const [isChangingType, setIsChangingType] = useState(false)

  const currentMachineType = availableMachineTypes.find(t => t.id === currentType) || availableMachineTypes[0]

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

  const isTransitioning = currentState !== desiredState
  const isStarting = currentState === 'starting' || (isTransitioning && desiredState === 'running')
  const isStopping = currentState === 'stopping' || (isTransitioning && desiredState === 'stopped')

  // Determine button state based on current and desired states
  const getButtonConfig = () => {
    if (currentState === 'stopped') {
      return {
        action: 'start',
        label: 'Start Machine',
        icon: <Play className="h-4 w-4" />,
        variant: 'default' as const,
        disabled: isLoading
      }
    }

    if (currentState === 'starting' || isStarting) {
      return {
        action: 'none',
        label: 'Starting...',
        icon: <Loader2 className="h-4 w-4 animate-spin" />,
        variant: 'default' as const,
        disabled: true
      }
    }

    if (currentState === 'stopping' || isStopping) {
      return {
        action: 'none',
        label: 'Stopping...',
        icon: <Loader2 className="h-4 w-4 animate-spin" />,
        variant: 'outline' as const,
        disabled: true
      }
    }

    if (currentState === 'running') {
      return {
        action: 'stop',
        label: 'Stop Machine',
        icon: <Square className="h-4 w-4" />,
        variant: 'destructive' as const,
        disabled: isLoading
      }
    }

    // Error or disabled state
    return {
      action: 'none',
      label: currentState.charAt(0).toUpperCase() + currentState.slice(1),
      icon: <AlertCircle className="h-4 w-4" />,
      variant: 'outline' as const,
      disabled: true
    }
  }

  const buttonConfig = getButtonConfig()

  return (
    <>
      <div className="flex items-center gap-2">
        {/* Start/Stop Button */}
        <Button
          onClick={buttonConfig.action === 'start' ? handleStart : buttonConfig.action === 'stop' ? handleStop : undefined}
          size="sm"
          className="gap-2"
          variant={buttonConfig.variant}
          disabled={buttonConfig.disabled}
        >
          {buttonConfig.icon}
          {buttonConfig.label}
        </Button>

        {/* Machine Type Selector */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="outline"
              size="sm"
              className="gap-2"
              disabled={isLoading || isTransitioning}
              title={isTransitioning ? 'Machine type cannot be changed during state transitions' : undefined}
            >
              <Zap className="h-4 w-4" />
              {currentMachineType?.name || 'Select Type'}
              <ChevronDown className="h-3 w-3" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-80">
            <DropdownMenuLabel>Machine Type</DropdownMenuLabel>
            <DropdownMenuSeparator />
            {availableMachineTypes.map((type) => (
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
                      {parseResourceValue(type.cpu)} vCPU
                    </span>
                    <span className="flex items-center gap-1">
                      <MemoryStick className="h-3 w-3" />
                      {parseResourceValue(type.memory)} GB
                    </span>
                    {type.gpu && (
                      <span className="flex items-center gap-1">
                        <Zap className="h-3 w-3" />
                        {parseResourceValue(type.gpu)} GPU
                      </span>
                    )}
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