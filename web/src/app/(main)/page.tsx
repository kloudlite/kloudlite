import { auth } from '@/lib/auth'
import { WorkMachinesContent } from './workspaces/_components/work-machines-content'
import { getMyWorkMachine, listAllWorkMachines } from '@/app/actions/work-machine.actions'
import { listMachineTypes } from '@/app/actions/machine-type.actions'
import type { WorkMachine } from '@/types/work-machine'

// Helper to map work machine CR to display format
function transformWorkMachine(wm: WorkMachine) {
  const desiredState = wm.spec.desiredState

  // Use status.state if it exists, otherwise use desiredState
  // Note: Transitions will only be visible once the controller starts updating status
  const currentState = wm.status?.state || desiredState

  // Calculate uptime from startedAt timestamp
  let uptime = '0 minutes'
  if (currentState === 'running' && wm.status?.startedAt) {
    const startTime = new Date(wm.status.startedAt)
    const now = new Date()
    const diffMs = now.getTime() - startTime.getTime()
    const diffMins = Math.floor(diffMs / 60000)

    if (diffMins < 60) {
      uptime = `${diffMins} minutes`
    } else {
      const hours = Math.floor(diffMins / 60)
      const mins = diffMins % 60
      uptime = `${hours}h ${mins}m`
    }
  }

  return {
    id: wm.metadata.name,
    owner: wm.spec.ownedBy,
    name: wm.metadata.name,
    currentState: currentState,
    desiredState: desiredState,
    // Legacy status for backward compatibility
    status: currentState === 'running' ? 'active' as const :
            currentState === 'stopped' ? 'stopped' as const : 'idle' as const,
    cpu: 0, // Will be updated by metrics
    memory: 0, // Will be updated by metrics
    disk: 0, // Will be updated by metrics
    uptime: uptime,
    type: wm.spec.machineType,
  }
}

// Main dashboard page - middleware ensures only users with 'user' role can access this
export default async function WorkMachinesPage() {
  const session = await auth()

  // Session is guaranteed to exist due to middleware checks
  const currentUser = session!.user?.email || 'user@example.com'
  const userRoles = session!.user?.roles || []
  const isSuperAdmin = userRoles.includes('super-admin')
  const isAdmin = userRoles.includes('admin') || isSuperAdmin

  // Fetch machine types
  const machineTypesResult = await listMachineTypes()
  const availableMachineTypes = machineTypesResult.success && machineTypesResult.data
    ? machineTypesResult.data.items
      .filter(mt => mt.spec.active !== false) // Only active types
      .map(mt => ({
        id: mt.metadata.name,
        name: mt.spec.displayName || mt.metadata.name,
        description: mt.spec.description || '',
        category: mt.spec.category || 'general',
        cpu: mt.spec.resources?.cpu || '',
        memory: mt.spec.resources?.memory || '',
        gpu: mt.spec.resources?.gpu,
      }))
    : []

  // Fetch real work machine data from CRs
  let workMachines: any[] = []

  if (isAdmin) {
    // Admin sees all work machines
    const result = await listAllWorkMachines()
    if (result.success && result.data) {
      workMachines = result.data.items.map(transformWorkMachine)
    }
  } else {
    // Regular user sees only their own machine
    const result = await getMyWorkMachine()
    if (result.success && result.data) {
      workMachines = [transformWorkMachine(result.data)]
    }
  }

  // TODO: Fetch pinned resources from actual CRs
  // For now, using empty arrays until workspace/environment CRs are implemented
  const pinnedWorkspaces: any[] = []
  const pinnedEnvironments: any[] = []

  return (
    <WorkMachinesContent
      initialMachines={workMachines}
      currentUser={currentUser}
      isAdmin={isAdmin}
      availableMachineTypes={availableMachineTypes}
      pinnedWorkspaces={pinnedWorkspaces}
      pinnedEnvironments={pinnedEnvironments}
    />
  )
}