'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Server, Rocket, Cpu, HardDrive } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { createMyWorkMachine } from '@/app/actions/work-machine.actions'
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
}

export function WorkMachineSetup({ availableMachineTypes }: WorkMachineSetupProps) {
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
    <main className="min-h-screen">
      <div className="mx-auto max-w-4xl px-6 py-16">
        {/* Welcome Header */}
        <div className="mb-12 text-center">
          <div className="bg-primary/10 mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-full">
            <Rocket className="text-primary h-10 w-10" />
          </div>
          <h1 className="mb-3 text-3xl font-bold">Welcome to Kloudlite!</h1>
          <p className="text-muted-foreground text-lg">
            Let's set up your development environment
          </p>
        </div>

        {/* Setup Card */}
        <div className="bg-card border p-8 shadow-sm">
          <div className="mb-8">
            <h2 className="mb-2 text-xl font-semibold">Choose Your Machine Configuration</h2>
            <p className="text-muted-foreground text-sm">
              Select the resources that match your development needs. You can change this later.
            </p>
          </div>

          {/* Machine Type Selector */}
          <div className="mb-6">
            <label className="mb-2 block text-sm font-medium">Machine Type</label>
            <Select value={selectedType} onValueChange={setSelectedType} disabled={isCreating}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Select a machine type" />
              </SelectTrigger>
              <SelectContent>
                {availableMachineTypes.map((mt) => (
                  <SelectItem key={mt.id} value={mt.id}>
                    <div className="flex items-center gap-2">
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
            <div className="bg-muted/50 mb-8 space-y-4 border p-6">
              <div>
                <h3 className="mb-1 text-base font-semibold">{selectedMachineType.name}</h3>
                <p className="text-muted-foreground text-sm">{selectedMachineType.description}</p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="flex items-center gap-3">
                  <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded">
                    <Cpu className="text-primary h-5 w-5" />
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">CPU</p>
                    <p className="text-sm font-medium">{selectedMachineType.cpu} Cores</p>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded">
                    <HardDrive className="text-primary h-5 w-5" />
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Memory</p>
                    <p className="text-sm font-medium">{selectedMachineType.memory}</p>
                  </div>
                </div>
              </div>

              <div className="bg-info/10 border-info mt-4 border-l-4 p-3">
                <p className="text-sm">
                  <span className="font-medium">Category:</span>{' '}
                  <span className="capitalize">{selectedMachineType.category}</span>
                </p>
              </div>
            </div>
          )}

          {/* No Types Available */}
          {availableMachineTypes.length === 0 && (
            <div className="bg-destructive/10 border-destructive mb-6 border p-4 text-center">
              <Server className="text-destructive mx-auto mb-2 h-8 w-8" />
              <p className="text-sm font-medium">No machine types available</p>
              <p className="text-muted-foreground mt-1 text-xs">
                Please contact your administrator to configure machine types.
              </p>
            </div>
          )}

          {/* Create Button */}
          <Button
            onClick={handleCreate}
            disabled={!selectedType || isCreating || availableMachineTypes.length === 0}
            className="w-full"
            size="lg"
          >
            {isCreating ? (
              <>
                <span className="mr-2">Creating...</span>
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
              </>
            ) : (
              'Create Work Machine'
            )}
          </Button>

          <p className="text-muted-foreground mt-4 text-center text-xs">
            This will take a few moments to provision your development environment
          </p>
        </div>
      </div>
    </main>
  )
}
