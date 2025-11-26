'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import {
  Play,
  Square,
  Settings,
  ChevronDown,
  Cpu,
  MemoryStick,
  Zap,
  AlertCircle,
  Check,
  Loader2,
  Copy,
  Trash2,
  Plus,
} from 'lucide-react'
import { toast } from 'sonner'
import { updateMyWorkMachine } from '@/app/actions/work-machine.actions'
import { Button } from '@kloudlite/ui'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@kloudlite/ui'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetFooter,
} from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Switch } from '@kloudlite/ui'
import { Alert, AlertDescription } from '@kloudlite/ui'

interface MachineType {
  id: string
  name: string
  description: string
  category: string
  cpu: string
  memory: string
  gpu?: string
}

interface WorkMachineControlsProps {
  machineId: string
  machineName: string
  status: 'active' | 'idle' | 'stopped'
  currentState: string
  desiredState: string
  currentType?: string
  availableMachineTypes: MachineType[]
  sshPublicKey?: string
  sshAuthorizedKeys?: string[]
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

// Parse SSH key to extract algorithm and email/comment
function parseSSHKey(key: string): string {
  const parts = key.trim().split(/\s+/)
  if (parts.length >= 2) {
    const algo = parts[0] // e.g., "ssh-rsa", "ssh-ed25519"
    const comment = parts.length > 2 ? parts[parts.length - 1] : '' // email or comment at the end
    return comment ? `${algo} ${comment}` : algo
  }
  return 'Unknown'
}

export function WorkMachineControls({
  machineId: _machineId,
  machineName,
  status,
  currentState,
  desiredState,
  currentType = 'standard',
  availableMachineTypes,
  sshPublicKey,
  sshAuthorizedKeys = [],
  onStart,
  onStop,
  onTypeChange,
  isLoading = false,
}: WorkMachineControlsProps) {
  const router = useRouter()
  const [_isPending, startTransition] = useTransition()
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [selectedType, setSelectedType] = useState(currentType)
  const [autoStop, setAutoStop] = useState(true)
  const [idleTimeout, setIdleTimeout] = useState('30')
  const [showTypeChangeWarning, setShowTypeChangeWarning] = useState(false)
  const [isChangingType, setIsChangingType] = useState(false)
  const [copiedSSHKey, setCopiedSSHKey] = useState(false)
  const [newSSHKey, setNewSSHKey] = useState('')
  const [isAddingKey, setIsAddingKey] = useState(false)
  const [isDeletingKey, setIsDeletingKey] = useState<string | null>(null)
  const [showStartConfirm, setShowStartConfirm] = useState(false)
  const [showStopConfirm, setShowStopConfirm] = useState(false)

  const currentMachineType =
    availableMachineTypes.find((t) => t.id === currentType) || availableMachineTypes[0]

  const handleStart = () => {
    setShowStartConfirm(true)
  }

  const confirmStart = () => {
    setShowStartConfirm(false)
    if (onStart) onStart()
  }

  const handleStop = () => {
    setShowStopConfirm(true)
  }

  const confirmStop = () => {
    setShowStopConfirm(false)
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

  const handleCopySSHKey = (key: string) => {
    navigator.clipboard.writeText(key)
    setCopiedSSHKey(true)
    toast.success('SSH public key copied to clipboard')
    setTimeout(() => setCopiedSSHKey(false), 2000)
  }

  const handleAddSSHKey = async () => {
    if (!newSSHKey.trim()) {
      toast.error('Please enter an SSH public key')
      return
    }

    // Basic validation for SSH public key format
    if (!newSSHKey.startsWith('ssh-') && !newSSHKey.startsWith('ecdsa-')) {
      toast.error('Invalid SSH public key format')
      return
    }

    setIsAddingKey(true)
    try {
      const updatedKeys = [...sshAuthorizedKeys, newSSHKey.trim()]
      const result = await updateMyWorkMachine({ sshPublicKeys: updatedKeys })

      if (result.success) {
        toast.success('SSH key added successfully')
        setNewSSHKey('')
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to add SSH key')
      }
    } catch (_error) {
      toast.error('An error occurred while adding SSH key')
    } finally {
      setIsAddingKey(false)
    }
  }

  const handleRemoveSSHKey = async (keyToRemove: string) => {
    setIsDeletingKey(keyToRemove)
    try {
      const updatedKeys = sshAuthorizedKeys.filter((k) => k !== keyToRemove)
      const result = await updateMyWorkMachine({ sshPublicKeys: updatedKeys })

      if (result.success) {
        toast.success('SSH key removed successfully')
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to remove SSH key')
      }
    } catch (_error) {
      toast.error('An error occurred while removing SSH key')
    } finally {
      setIsDeletingKey(null)
    }
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
        disabled: isLoading,
      }
    }

    if (currentState === 'starting' || isStarting) {
      return {
        action: 'none',
        label: 'Starting...',
        icon: <Loader2 className="h-4 w-4 animate-spin" />,
        variant: 'default' as const,
        disabled: true,
      }
    }

    if (currentState === 'stopping' || isStopping) {
      return {
        action: 'none',
        label: 'Stopping...',
        icon: <Loader2 className="h-4 w-4 animate-spin" />,
        variant: 'outline' as const,
        disabled: true,
      }
    }

    if (currentState === 'running') {
      return {
        action: 'stop',
        label: 'Stop Machine',
        icon: <Square className="h-4 w-4" />,
        variant: 'destructive' as const,
        disabled: isLoading,
      }
    }

    // Error or disabled state
    return {
      action: 'none',
      label: currentState.charAt(0).toUpperCase() + currentState.slice(1),
      icon: <AlertCircle className="h-4 w-4" />,
      variant: 'outline' as const,
      disabled: true,
    }
  }

  const buttonConfig = getButtonConfig()

  return (
    <>
      <div className="flex items-center gap-2">
        {/* Start/Stop Button */}
        <Button
          onClick={
            buttonConfig.action === 'start'
              ? handleStart
              : buttonConfig.action === 'stop'
                ? handleStop
                : undefined
          }
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
              title={
                isTransitioning
                  ? 'Machine type cannot be changed during state transitions'
                  : undefined
              }
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
                  {selectedType === type.id && <Check className="text-success h-4 w-4" />}
                </div>
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <span className="font-medium">{type.name}</span>
                  </div>
                  <div className="text-muted-foreground mt-1 text-xs">{type.description}</div>
                  <div className="text-muted-foreground mt-2 flex items-center gap-4 text-xs">
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
        <Button variant="outline" size="sm" onClick={() => setSettingsOpen(true)}>
          <Settings className="h-4 w-4" />
        </Button>
      </div>

      {/* Type Change Warning Dialog */}
      <Dialog open={showTypeChangeWarning} onOpenChange={setShowTypeChangeWarning}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Change Machine Type</DialogTitle>
            <DialogDescription>
              Changing the machine type requires stopping and restarting your work machine. All
              running workspaces will be temporarily suspended.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Your work machine will be unavailable for approximately 2-3 minutes during the
                resize operation.
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
            <Button onClick={confirmTypeChange} disabled={isChangingType}>
              {isChangingType ? 'Changing...' : 'Proceed with Change'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Settings Sheet */}
      <Sheet open={settingsOpen} onOpenChange={setSettingsOpen}>
        <SheetContent side="right" className="w-full sm:max-w-xl">
          <div className="flex h-full flex-col">
            <SheetHeader>
              <SheetTitle>Work Machine Settings</SheetTitle>
              <SheetDescription>
                Configure automatic behaviors and startup scripts for {machineName}
              </SheetDescription>
            </SheetHeader>

            <div className="flex-1 space-y-8 overflow-y-auto p-4">
              {/* Auto-stop Settings */}
              <div className="space-y-4">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <Label htmlFor="auto-stop" className="text-base font-medium">
                      Auto-stop when idle
                    </Label>
                    <p className="text-muted-foreground text-sm">
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

              {/* SSH Keys */}
              <div className="space-y-4 border-t pt-6">
                <div>
                  <Label className="text-base font-medium">SSH Configuration</Label>
                  <p className="text-muted-foreground mt-1 text-sm">
                    Manage SSH keys for workspace access
                  </p>
                </div>

                {/* SSH Public Key */}
                {sshPublicKey ? (
                  <div className="space-y-3">
                    <div>
                      <Label className="text-sm font-medium">WorkMachine SSH Public Key</Label>
                      <p className="text-muted-foreground mt-1 text-xs">
                        This SSH public key is shared across all workspaces in this WorkMachine. Use
                        it to authorize access to external systems.
                      </p>
                    </div>
                    <div className="bg-muted/50 relative rounded-lg border p-4">
                      <pre className="text-muted-foreground overflow-x-auto font-mono text-xs leading-relaxed whitespace-pre-wrap break-all">
                        {sshPublicKey}
                      </pre>
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={() => handleCopySSHKey(sshPublicKey)}
                        className="absolute right-2 top-2"
                      >
                        {copiedSSHKey ? (
                          <>
                            <Check className="mr-1.5 h-3.5 w-3.5" />
                            Copied
                          </>
                        ) : (
                          <>
                            <Copy className="mr-1.5 h-3.5 w-3.5" />
                            Copy
                          </>
                        )}
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="rounded-md border border-dashed py-4 text-center">
                    <p className="text-muted-foreground text-xs">
                      SSH public key is being generated
                    </p>
                  </div>
                )}

                {/* Authorized Keys */}
                <div className="space-y-3 pt-3">
                  <div className="flex items-start justify-between">
                    <Label className="text-base font-medium">Authorized Keys</Label>
                    {sshAuthorizedKeys.length > 0 && (
                      <span className="text-muted-foreground text-xs">
                        {sshAuthorizedKeys.length} {sshAuthorizedKeys.length === 1 ? 'key' : 'keys'}
                      </span>
                    )}
                  </div>
                  <p className="text-muted-foreground text-xs">
                    Add SSH public keys to authorize external access to all workspaces in this
                    WorkMachine
                  </p>

                  {/* List of authorized keys */}
                  {sshAuthorizedKeys.length > 0 ? (
                    <div className="max-h-48 space-y-1 overflow-y-auto rounded-md border">
                      {sshAuthorizedKeys.map((key, index) => (
                        <div
                          key={index}
                          className="hover:bg-muted/50 group flex items-center justify-between px-3 py-2 transition-colors"
                        >
                          <code className="text-muted-foreground font-mono text-xs">
                            {parseSSHKey(key)}
                          </code>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={async () => {
                              await handleRemoveSSHKey(key)
                            }}
                            disabled={isDeletingKey === key}
                            className="h-6 w-6 p-0 opacity-0 transition-opacity group-hover:opacity-100"
                          >
                            {isDeletingKey === key ? (
                              <Loader2 className="h-3 w-3 animate-spin" />
                            ) : (
                              <Trash2 className="h-3 w-3" />
                            )}
                          </Button>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="rounded-md border border-dashed py-6 text-center">
                      <p className="text-muted-foreground text-sm">No authorized keys</p>
                    </div>
                  )}

                  {/* Add new key */}
                  <div className="space-y-2 pt-2">
                    <Label htmlFor="new-ssh-key" className="text-sm font-medium">
                      Add New Key
                    </Label>
                    <div className="flex gap-2">
                      <Input
                        id="new-ssh-key"
                        type="text"
                        placeholder="ssh-rsa AAAAB3NzaC1yc2EA... user@example.com"
                        value={newSSHKey}
                        onChange={(e) => setNewSSHKey(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            handleAddSSHKey()
                          }
                        }}
                        className="flex-1 font-mono text-xs"
                        disabled={isAddingKey}
                      />
                      <Button
                        size="sm"
                        onClick={handleAddSSHKey}
                        disabled={isAddingKey || !newSSHKey.trim()}
                        className="flex-shrink-0"
                      >
                        {isAddingKey ? (
                          <>
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                            Adding...
                          </>
                        ) : (
                          <>
                            <Plus className="mr-2 h-4 w-4" />
                            Add
                          </>
                        )}
                      </Button>
                    </div>
                    <p className="text-muted-foreground text-xs">
                      Paste the full SSH public key (e.g., ssh-rsa, ssh-ed25519)
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <SheetFooter className="p-4">
              <Button variant="outline" onClick={() => setSettingsOpen(false)}>
                Cancel
              </Button>
              <Button onClick={saveSettings}>Save Settings</Button>
            </SheetFooter>
          </div>
        </SheetContent>
      </Sheet>

      {/* Start Machine Confirmation Dialog */}
      <Dialog open={showStartConfirm} onOpenChange={setShowStartConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Start WorkMachine?</DialogTitle>
            <DialogDescription>
              This will start the WorkMachine &quot;{machineName}&quot;. The machine will begin consuming resources
              and you will be charged according to your plan.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowStartConfirm(false)}>
              Cancel
            </Button>
            <Button onClick={confirmStart}>Start Machine</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Stop Machine Confirmation Dialog */}
      <Dialog open={showStopConfirm} onOpenChange={setShowStopConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Stop WorkMachine?</DialogTitle>
            <DialogDescription>
              This will stop the WorkMachine &quot;{machineName}&quot;. All workspaces and environments will be
              suspended. Your disk storage will be preserved, but the machine will stop consuming compute
              resources.
            </DialogDescription>
          </DialogHeader>
          <Alert className="border-warning bg-warning/10">
            <AlertCircle className="text-warning h-4 w-4" />
            <AlertDescription className="text-sm">
              Active workspaces will be suspended and service intercepts will be disconnected.
            </AlertDescription>
          </Alert>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowStopConfirm(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={confirmStop}>
              Stop Machine
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
