import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { WorkMachinesContent } from './workspaces/_components/work-machines-content'
import { getDashboardData } from '@/lib/services/dashboard.service'
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

  // Single API call to get all dashboard data
  const dashboardData = await getDashboardData()

  // Transform machine types for the component
  const availableMachineTypes = dashboardData.machineTypes
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

  // Transform work machines
  const workMachines = dashboardData.workMachines.map(transformWorkMachine)

  // Transform pinned workspaces
  interface PinnedWorkspace {
    id: string
    name: string
    environment: string
    status: 'active' | 'idle'
  }
  const pinnedWorkspaces: PinnedWorkspace[] = dashboardData.pinnedWorkspaces.map((ws) => ({
    id: `${ws.metadata.namespace}/${ws.metadata.name}`,
    name: `${ws.spec.ownedBy}/${ws.spec.displayName || ws.metadata.name}`,
    environment: ws.status?.connectedEnvironment?.name || '-',
    status: ws.status?.phase === 'Running' ? 'active' : 'idle',
  }))

  // Transform pinned environments
  interface PinnedEnvironment {
    id: string
    name: string
    status: 'active' | 'idle'
  }
  const pinnedEnvironments: PinnedEnvironment[] = dashboardData.pinnedEnvironments.map((env) => ({
    id: env.metadata.name,
    name: `${env.spec.ownedBy}/${env.spec.name || env.metadata.name}`,
    status: env.status?.state === 'active' ? 'active' : 'idle',
  }))

  return (
    <WorkMachinesContent
      initialMachines={workMachines}
      currentUser={currentUser}
      isAdmin={dashboardData.isAdmin}
      availableMachineTypes={availableMachineTypes}
      pinnedWorkspaces={pinnedWorkspaces}
      pinnedEnvironments={pinnedEnvironments}
    />
  )
}
