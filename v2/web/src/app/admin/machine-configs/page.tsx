import { MachineConfigsList } from '@/components/machine-configs-list'
import { listMachineTypes } from '@/app/actions/machine-type.actions'
import type { MachineType } from '@/types/machine'

// Parse Kubernetes resource strings (e.g., "4Gi" -> 4)
function parseResourceValue(value?: string | number): number {
  if (typeof value === 'number') return value
  if (!value) return 0

  const match = value.match(/^(\d+(?:\.\d+)?)/);
  if (match) {
    return parseFloat(match[1])
  }
  return 0
}

// Transform MachineType to the format expected by MachineConfigsList
function transformMachineTypes(machineTypes: MachineType[]) {
  return machineTypes.map((mt) => ({
    id: mt.metadata.name,
    name: mt.spec.displayName || mt.metadata.name,
    cpu: parseResourceValue(mt.spec.resources?.cpu || mt.spec.cpu),
    memory: parseResourceValue(mt.spec.resources?.memory || mt.spec.memory),
    storage: parseResourceValue(mt.spec.resources?.storage || mt.spec.storage),
    gpu: mt.spec.resources?.gpu ? parseResourceValue(mt.spec.resources.gpu) : mt.spec.gpu,
    maxInstances: 10, // Default value, can be made configurable
    activeInstances: 0, // Would need to be fetched from actual usage
    pricePerHour: mt.spec.pricePerHour || 0,
    description: mt.spec.description || '',
    category: mt.spec.category || 'general',
    active: mt.spec.active !== false
  }))
}

export default async function MachineConfigsPage() {
  // Fetch machine types from the API
  const result = await listMachineTypes()

  const machineTypes = result.success && result.data
    ? result.data.items || []
    : []

  const transformedConfigs = transformMachineTypes(machineTypes)

  return (
    <main className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-light tracking-tight">Machine Configurations</h1>
        <p className="text-sm text-gray-600 mt-2">
          Define machine types and resource allocations
        </p>
      </div>

      {/* Machine Configs List Component */}
      <MachineConfigsList configs={transformedConfigs} />
    </main>
  )
}