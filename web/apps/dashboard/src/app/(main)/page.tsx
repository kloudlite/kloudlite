import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { listMachineTypes } from '@/app/actions/machine-type.actions'
import { getMyWorkMachine } from '@/app/actions/work-machine.actions'
import { getMyPreferences } from '@/app/actions/user-preferences.actions'
import type { WorkMachine } from '@kloudlite/types'
import { WorkMachineCard } from './workspaces/_components/work-machine-card'
import { WorkMachineMetrics } from './workspaces/_components/work-machine-metrics'

// Helper to map work machine CR to display format
function transformWorkMachine(wm: WorkMachine) {
  // Use status.state as the source of truth for machine state
  let state = wm.status?.state || wm.spec.state
  const desiredState = wm.spec.state

  // Handle transitional states
  const isReady = wm.status?.isReady ?? false

  // Machine is starting if desired is 'running' but current is not running or not ready
  if (desiredState === 'running' && (state !== 'running' || !isReady)) {
    state = 'starting'
  }
  // Machine is stopping if desired is 'stopped' but current is still 'running'
  else if (desiredState === 'stopped' && state === 'running') {
    state = 'stopping'
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

// Main dashboard page - THIS PAGE IS ONLY SHOWN WHEN USER HAS A WORK MACHINE
// The layout.tsx handles showing WorkMachineSetup when user doesn't have one
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
  const workMachine = myWorkMachineResult.success && myWorkMachineResult.data
    ? transformWorkMachine(myWorkMachineResult.data)
    : null

  // If no work machine, the layout.tsx will handle showing the setup page
  // This should not happen as layout redirects users without work machines
  if (!workMachine) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-muted-foreground">Loading work machine...</p>
      </div>
    )
  }


  return (
    <>
      {/* Page Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight mb-2">Workmachine</h1>
        <p className="text-muted-foreground text-sm">
          Monitor and control your development workmachine
        </p>
      </div>

      {/* Machine Card - Client Component for Interactivity */}
      <WorkMachineCard
        machine={workMachine}
        availableMachineTypes={availableMachineTypes}
      />

      {/* Monitoring Section - Only show when machine is running */}
      {workMachine.currentState === 'running' && (
        <div className="space-y-4">
          <div>
            <h2 className="text-lg font-semibold tracking-tight">Monitoring</h2>
            <p className="text-muted-foreground mt-1 text-sm">
              Real-time metrics and system health monitoring
            </p>
          </div>

          <WorkMachineMetrics
            workMachineName={workMachine.name}
            machineState={workMachine.currentState}
          />
        </div>
      )}
    </>
  )
}
