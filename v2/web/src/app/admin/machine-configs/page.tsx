import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
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
    gpu: mt.spec.resources?.gpu ? parseResourceValue(mt.spec.resources.gpu) : mt.spec.gpu,
    description: mt.spec.description || '',
    category: mt.spec.category || 'general',
    active: mt.spec.active !== false
  }))
}

export default async function MachineConfigsPage() {
  // Check authentication and permissions
  const session = await auth()
  if (!session || !session.user?.email) {
    redirect('/auth/signin')
  }

  // Check if user has admin or super-admin role
  const userRoles = session.user?.roles || []
  const hasAdminAccess = userRoles.includes('admin') || userRoles.includes('super-admin')
  const isSuperAdmin = userRoles.includes('super-admin')

  if (!hasAdminAccess) {
    redirect('/')
  }

  // Fetch machine types from the API
  const result = await listMachineTypes()

  const machineTypes = result.success && result.data
    ? result.data.items || []
    : []

  const transformedConfigs = transformMachineTypes(machineTypes)

  return (
    <div className="mx-auto max-w-7xl px-6 py-8 space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-light tracking-tight">Machine Configurations</h1>
        <p className="text-sm text-gray-600 mt-2">
          Define machine types and resource allocations
        </p>
      </div>

      {/* Machine Configs List Component */}
      <MachineConfigsList configs={transformedConfigs} isReadOnly={!isSuperAdmin} />
    </div>
  )
}