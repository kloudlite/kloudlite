import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { WorkMachinesContent } from './workspaces/_components/work-machines-content'
import { getMyWorkMachine, listAllWorkMachines } from '@/app/actions/work-machine.actions'
import { listMachineTypes } from '@/app/actions/machine-type.actions'
import { getMyPreferences } from '@/app/actions/user-preferences.actions'
import { workspaceService } from '@/lib/services/workspace.service'
import { environmentService } from '@/lib/services/environment.service'
import type { WorkMachine } from '@kloudlite/types'

// Helper to map work machine CR to display format
function transformWorkMachine(wm: WorkMachine) {
  // Use status.state as the source of truth for machine state
  let state = wm.status?.state || wm.spec.state
  const desiredState = wm.spec.state

  // Handle transitional states when machine is not ready
  const isReady = wm.status?.isReady ?? false
  if (!isReady && state === 'running') {
    // If desired state is 'stopped', show as 'stopping' (user clicked stop)
    // Otherwise show as 'starting' (machine still initializing)
    state = desiredState === 'stopped' ? 'stopping' : 'starting'
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

  // Fetch machine types, work machines, and preferences in parallel
  const [machineTypesResult, workMachinesResult, prefsResult] = await Promise.all([
    listMachineTypes(),
    isAdmin ? listAllWorkMachines() : getMyWorkMachine(),
    getMyPreferences(),
  ])

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

  // Transform work machines result
  let workMachines: ReturnType<typeof transformWorkMachine>[] = []
  if (workMachinesResult.success && workMachinesResult.data) {
    if (isAdmin && 'items' in workMachinesResult.data) {
      workMachines = workMachinesResult.data.items.map(transformWorkMachine)
    } else if (!isAdmin) {
      workMachines = [transformWorkMachine(workMachinesResult.data as WorkMachine)]
    }
  }

  const prefs = prefsResult.success ? prefsResult.data : null

  // Fetch pinned workspaces and environments in parallel
  interface PinnedWorkspace {
    id: string
    name: string
    environment: string
    status: 'active' | 'idle'
  }
  interface PinnedEnvironment {
    id: string
    name: string
    status: 'active' | 'idle'
  }

  // Fetch all pinned workspaces in parallel
  const pinnedWorkspacePromises = (prefs?.spec.pinnedWorkspaces || []).map(async (ref): Promise<PinnedWorkspace | null> => {
    try {
      const ws = await workspaceService.get(ref.name, ref.namespace || '')
      return {
        id: `${ref.namespace}/${ref.name}`,
        name: `${ws.spec.ownedBy}/${ws.spec.displayName || ws.metadata.name}`,
        environment: ws.status?.connectedEnvironment?.name || '-',
        status: ws.status?.phase === 'Running' ? 'active' : 'idle',
      }
    } catch {
      return null // Workspace may have been deleted
    }
  })

  // Fetch all pinned environments in parallel
  const pinnedEnvironmentPromises = (prefs?.spec.pinnedEnvironments || []).map(async (envName): Promise<PinnedEnvironment | null> => {
    try {
      const env = await environmentService.getEnvironment(envName)
      return {
        id: envName,
        name: `${env.spec.ownedBy}/${env.spec.name || env.metadata.name}`,
        status: env.status?.state === 'active' ? 'active' : 'idle',
      }
    } catch {
      return null // Environment may have been deleted
    }
  })

  // Wait for all pinned resources to load in parallel
  const [pinnedWorkspacesResults, pinnedEnvironmentsResults] = await Promise.all([
    Promise.all(pinnedWorkspacePromises),
    Promise.all(pinnedEnvironmentPromises),
  ])

  const pinnedWorkspaces: PinnedWorkspace[] = pinnedWorkspacesResults.filter((ws): ws is PinnedWorkspace => ws !== null)
  const pinnedEnvironments: PinnedEnvironment[] = pinnedEnvironmentsResults.filter((env): env is PinnedEnvironment => env !== null)

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
