'use client'

import { useState } from 'react'
import { WorkMachineMetrics } from '@/components/work-machine-metrics'
import { PinnedResources } from '@/components/pinned-resources'
import { WorkMachineControls } from '@/components/work-machine-controls'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { ChevronDown, Server } from 'lucide-react'

interface WorkMachine {
  id: string
  owner: string
  name: string
  status: 'active' | 'idle' | 'stopped'
  cpu: number
  memory: number
  disk: number
  uptime: string
  type: string
}

interface WorkMachinesContentProps {
  initialMachines: WorkMachine[]
  currentUser: string
  isAdmin: boolean
  pinnedWorkspaces: any[]
  pinnedEnvironments: any[]
}

export function WorkMachinesContent({
  initialMachines,
  currentUser,
  isAdmin,
  pinnedWorkspaces,
  pinnedEnvironments
}: WorkMachinesContentProps) {
  const [workMachines, setWorkMachines] = useState(initialMachines)
  const [selectedMachineId, setSelectedMachineId] = useState(workMachines[0]?.id)

  const selectedMachine = workMachines.find(m => m.id === selectedMachineId) || workMachines[0]

  const handleStart = () => {
    setWorkMachines(machines =>
      machines.map(m =>
        m.id === selectedMachineId
          ? { ...m, status: 'active' as const, cpu: 45, memory: 62 }
          : m
      )
    )
  }

  const handleStop = () => {
    setWorkMachines(machines =>
      machines.map(m =>
        m.id === selectedMachineId
          ? { ...m, status: 'stopped' as const, cpu: 0, memory: 0, uptime: '0 minutes' }
          : m
      )
    )
  }

  const handleTypeChange = (typeId: string) => {
    setWorkMachines(machines =>
      machines.map(m =>
        m.id === selectedMachineId
          ? { ...m, type: typeId }
          : m
      )
    )
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
              <p className="text-xs text-gray-500">Status</p>
              <p className={`text-sm font-medium ${
                selectedMachine.status === 'active' ? 'text-green-600' :
                selectedMachine.status === 'idle' ? 'text-yellow-600' :
                'text-gray-600'
              }`}>
                {selectedMachine.status}
              </p>
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
            currentType={selectedMachine.type}
            onStart={handleStart}
            onStop={handleStop}
            onTypeChange={handleTypeChange}
          />
        </div>
      </div>

      {/* Metrics Section - Always show, but CPU/Memory are 0 when stopped */}
      <div className="mb-8">
        <WorkMachineMetrics
          cpu={selectedMachine.cpu}
          memory={selectedMachine.memory}
          disk={selectedMachine.disk}
        />
      </div>

      {/* Stopped State Message */}
      {selectedMachine.status === 'stopped' && (
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