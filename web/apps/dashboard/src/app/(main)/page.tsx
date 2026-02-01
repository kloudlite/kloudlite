import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { WorkMachinesContent } from './workspaces/_components/work-machines-content'
import { listMachineTypes } from '@/app/actions/machine-type.actions'
import { getMyWorkMachine } from '@/app/actions/work-machine.actions'
import { getMyPreferences } from '@/app/actions/user-preferences.actions'
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
  const username = session.user?.username || currentUser.split('@')[0]

  // Fetch data directly from Kubernetes
  const [machineTypesResult, myWorkMachineResult, preferencesResult] = await Promise.all([
    listMachineTypes(),
    getMyWorkMachine(),
    getMyPreferences(),
  ])

  // Transform machine types for the component
  const availableMachineTypes = machineTypesResult.success && machineTypesResult.data
    ? machineTypesResult.data
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

  // Transform work machine (single machine for current user)
  const workMachines = myWorkMachineResult.success && myWorkMachineResult.data
    ? [transformWorkMachine(myWorkMachineResult.data)]
    : []

  // Pinned items from preferences (TODO: implement fetching pinned workspaces/environments)
  const pinnedWorkspaces: Array<{id: string; name: string; environment: string; status: 'active' | 'idle'}> = []
  const pinnedEnvironments: Array<{id: string; name: string; status: 'active' | 'idle'}> = []

  // Check if user is admin
  const userRoles = session.user?.roles || []
  const isAdmin = userRoles.includes('admin') || userRoles.includes('super-admin')

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
