'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Server, Rocket, Cpu, HardDrive, Sparkles } from 'lucide-react'
import {
  Button,
  KloudliteLogo,
  ThemeSwitcher,
} from '@kloudlite/ui'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'
import { createMyWorkMachine } from '@/app/actions/work-machine.actions'
import { setThemeCookie } from '@/app/actions/theme'
import { UserProfileDropdown } from '@/components/user-profile-dropdown'
import { toast } from 'sonner'

interface MachineType {
  id: string
  name: string
  description: string
  category: string
  cpu: string
  memory: string
  gpu?: string
}

interface WorkMachineSetupProps {
  availableMachineTypes: MachineType[]
  userEmail?: string
  displayName?: string
  isAdmin?: boolean
  isSuperAdmin?: boolean
}

export function WorkMachineSetup({
  availableMachineTypes,
  userEmail,
  displayName,
  isAdmin,
  isSuperAdmin,
}: WorkMachineSetupProps) {
  const router = useRouter()
  const [_isPending, startTransition] = useTransition()
  const [selectedType, setSelectedType] = useState<string>('')
  const [isCreating, setIsCreating] = useState(false)

  // Get the selected machine type details
  const selectedMachineType = availableMachineTypes.find((mt) => mt.id === selectedType)

  const handleCreate = async () => {
    if (!selectedType) {
      toast.error('Please select a machine type')
      return
    }

    setIsCreating(true)
    try {
      const result = await createMyWorkMachine(selectedType)
      if (result.success) {
        toast.success('Work machine created successfully!')
        // Refresh the page to show the new machine
        startTransition(() => {
          router.refresh()
        })
      } else {
        toast.error(result.error || 'Failed to create work machine')
        setIsCreating(false)
      }
    } catch (_error) {
      toast.error('An error occurred while creating work machine')
      setIsCreating(false)
    }
  }

  return (
    <div className="bg-background min-h-screen flex items-center justify-center p-4 sm:p-6 relative">
      {/* Top branding */}
      <div className="absolute top-6 left-6 z-10">
        <KloudliteLogo className="text-lg font-semibold" />
      </div>

      {/* Top right user controls */}
      <div className="absolute top-6 right-6 z-10 flex items-center gap-2">
        <ThemeSwitcher setThemeCookie={setThemeCookie} />
        <div className="h-6 w-px bg-border" />
        <UserProfileDropdown
          email={userEmail}
          displayName={displayName}
          isAdmin={isAdmin}
          isSuperAdmin={isSuperAdmin}
        />
      </div>

      {/* Main setup card - centered */}
      <div className="relative w-full max-w-2xl">
        <div className="bg-card border rounded-lg p-10 sm:p-12 lg:p-14">
          {/* Welcome Header */}
          <div className="mb-10 text-center">
            <div className="bg-primary/10 mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-full ring-8 ring-primary/5">
              <Rocket className="text-primary h-8 w-8" />
            </div>
            <h1 className="mb-2 text-2xl sm:text-3xl font-bold tracking-tight">Welcome to Kloudlite!</h1>
            <p className="text-muted-foreground text-sm sm:text-base leading-relaxed">
              Let&apos;s create your cloud development environment
            </p>
          </div>

          {/* Setup Section */}
          <div className="space-y-6">
            <div className="text-center mb-6">
              <h2 className="text-lg font-semibold mb-2 flex items-center justify-center gap-2">
                <Sparkles className="h-5 w-5 text-primary" />
                Choose Your Machine Configuration
              </h2>
              <p className="text-muted-foreground text-sm leading-relaxed">
                Select the resources that match your development needs
              </p>
            </div>

            {/* Machine Type Selector */}
            <div className="space-y-2">
              <label className="block text-sm font-medium">Machine Type</label>
              <Select value={selectedType} onValueChange={setSelectedType} disabled={isCreating}>
                <SelectTrigger className="w-full h-11">
                  <SelectValue placeholder="Select a machine type" />
                </SelectTrigger>
                <SelectContent>
                  {availableMachineTypes.map((mt) => (
                    <SelectItem key={mt.id} value={mt.id}>
                      <div className="flex items-center gap-2 py-1">
                        <span className="font-medium">{mt.name}</span>
                        <span className="text-muted-foreground text-xs">
                          ({mt.cpu} CPU, {mt.memory} RAM)
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Selected Machine Details */}
            {selectedMachineType && (
              <div className="bg-muted/20 border border-border/50 rounded-lg p-6 space-y-5">
                <div>
                  <h3 className="text-base font-semibold mb-1">{selectedMachineType.name}</h3>
                  <p className="text-muted-foreground text-sm">{selectedMachineType.description}</p>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="flex items-center gap-3 bg-background/80 border border-border/30 rounded-lg p-3">
                    <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg">
                      <Cpu className="text-primary h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground mb-0.5">CPU</p>
                      <p className="text-sm font-semibold">{selectedMachineType.cpu} Cores</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3 bg-background/80 border border-border/30 rounded-lg p-3">
                    <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg">
                      <HardDrive className="text-primary h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground mb-0.5">Memory</p>
                      <p className="text-sm font-semibold">{selectedMachineType.memory}</p>
                    </div>
                  </div>
                </div>

                <div className="bg-primary/5 border-l-4 border-primary rounded p-3">
                  <p className="text-sm">
                    <span className="font-medium">Category:</span>{' '}
                    <span className="capitalize text-muted-foreground">{selectedMachineType.category}</span>
                  </p>
                </div>
              </div>
            )}

            {/* No Types Available */}
            {availableMachineTypes.length === 0 && (
              <div className="bg-destructive/10 border-destructive border rounded-lg p-6 text-center">
                <Server className="text-destructive mx-auto mb-3 h-10 w-10" />
                <p className="text-sm font-semibold mb-1">No machine types available</p>
                <p className="text-muted-foreground text-xs">
                  Please contact your administrator to configure machine types
                </p>
              </div>
            )}

            {/* Create Button */}
            <Button
              onClick={handleCreate}
              disabled={!selectedType || isCreating || availableMachineTypes.length === 0}
              className="w-full h-11 text-base font-semibold"
              size="lg"
            >
              {isCreating ? (
                <>
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent mr-2" />
                  Creating Your Environment...
                </>
              ) : (
                <>
                  <Rocket className="h-4 w-4 mr-2" />
                  Create Work Machine
                </>
              )}
            </Button>

            {selectedType && !isCreating && (
              <p className="text-muted-foreground text-center text-xs">
                This will take a few moments to provision your development environment
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Bottom branding */}
      <div className="absolute bottom-6 left-0 right-0 text-center">
        <p className="text-muted-foreground/60 text-xs tracking-wide">
          Powered by Kloudlite · Cloud Development Environments
        </p>
      </div>
    </div>
  )
}
