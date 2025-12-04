import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { WorkMachinesContent } from './workspaces/_components/work-machines-content'
import { getMyWorkMachine, listAllWorkMachines } from '@/app/actions/work-machine.actions'
import { listMachineTypes } from '@/app/actions/machine-type.actions'
import type { WorkMachine } from '@kloudlite/types'

// Helper to map work machine CR to display format
function transformWorkMachine(wm: WorkMachine) {
  // Use status.state as the source of truth for machine state
  let state = wm.status?.state || wm.spec.state

  // If machine is not ready yet, show as "starting" even if state is "running"
  const isReady = wm.status?.isReady ?? false
  if (!isReady && state === 'running') {
    state = 'starting'
  }

  // Calculate uptime from startedAt timestamp
  let uptime = '0 minutes'
  if (state === 'running' && wm.status?.startedAt) {
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
    currentState: state,
    desiredState: wm.spec.state,
    status:
      state === 'running'
        ? ('active' as const)
        : state === 'stopped'
          ? ('stopped' as const)
          : ('idle' as const),
    cpu: 0,
    memory: 0,
    disk: 0,
    uptime: uptime,
    type: wm.spec.machineType,
    sshPublicKey: wm.status?.sshPublicKey,
    sshAuthorizedKeys: wm.spec.sshPublicKeys || [],
    autoShutdown: wm.spec.autoShutdown
      ? {
          enabled: wm.spec.autoShutdown.enabled,
          idleThresholdMinutes: wm.spec.autoShutdown.idleThresholdMinutes,
        }
      : undefined,
  }
}

// Main dashboard page
export default async function HomePage() {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  const currentUser = session.user?.email || 'user@example.com'
  const userRoles = session.user?.roles || []
  const isSuperAdmin = userRoles.includes('super-admin')
  const isAdmin = userRoles.includes('admin') || isSuperAdmin

  // Fetch machine types
  const machineTypesResult = await listMachineTypes()
  const availableMachineTypes =
    machineTypesResult.success && machineTypesResult.data
      ? machineTypesResult.data.items
          .filter((mt) => mt.spec.active !== false)
          .map((mt) => ({
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
  let workMachines: ReturnType<typeof transformWorkMachine>[] = []

  if (isAdmin) {
    const result = await listAllWorkMachines()
    if (result.success && result.data) {
      workMachines = result.data.items.map(transformWorkMachine)
    }
  } else {
    const result = await getMyWorkMachine()
    if (result.success && result.data) {
      workMachines = [transformWorkMachine(result.data)]
    }
  }

  const pinnedWorkspaces: never[] = []
  const pinnedEnvironments: never[] = []

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
