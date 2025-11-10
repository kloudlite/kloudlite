import { getSession } from '@/lib/get-session'
import { WorkMachinesContent } from './workspaces/_components/work-machines-content'
import { getMyWorkMachine, listAllWorkMachines } from '@/app/actions/work-machine.actions'
import { listMachineTypes } from '@/app/actions/machine-type.actions'
import type { WorkMachine } from '@/types/work-machine'
import { APP_MODE } from '@/lib/app-mode'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { TypingText } from './_components/typing-text'
import { GetStartedButton } from '@/components/get-started-button'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'

// Type for transformed work machine display format
interface TransformedWorkMachine {
  id: string
  owner: string
  name: string
  currentState: string
  desiredState: string
  status: 'active' | 'stopped' | 'idle'
  cpu: number
  memory: number
  disk: number
  uptime: string
  type: string
  sshPublicKey?: string
  sshAuthorizedKeys: string[]
}

// Helper to map work machine CR to display format
function transformWorkMachine(wm: WorkMachine): TransformedWorkMachine {
  // Use status.state as the source of truth for machine state
  // Controller updates status.state to reflect actual machine state
  let state = wm.status?.state || wm.spec.state

  // If machine is not ready yet (DomainRequest or host-manager not ready),
  // show as "starting" even if state is "running"
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
    desiredState: wm.spec.state, // Desired state is what the user wants
    // Legacy status for backward compatibility
    status:
      state === 'running'
        ? ('active' as const)
        : state === 'stopped'
          ? ('stopped' as const)
          : ('idle' as const),
    cpu: 0, // Will be updated by metrics
    memory: 0, // Will be updated by metrics
    disk: 0, // Will be updated by metrics
    uptime: uptime,
    type: wm.spec.machineType,
    sshPublicKey: wm.status?.sshPublicKey,
    sshAuthorizedKeys: wm.spec.sshPublicKeys || [],
  }
}

// Website Landing Page Component
function WebsiteLandingPage() {
  return (
    <div className="bg-background flex min-h-screen flex-col overflow-x-hidden">
      <WebsiteHeader currentPage="home" />

      {/* Hero Section */}
      <main className="flex flex-1 items-center py-12 sm:py-16 w-full">
        <div className="mx-auto w-full max-w-[90rem] px-4 sm:px-6 lg:px-8">
          {/* Main Heading */}
          <div className="text-center">
            <h1 className="text-foreground text-2xl leading-tight font-bold tracking-tight sm:text-4xl md:text-5xl lg:text-6xl xl:text-7xl break-words">
              Platform of Development Environments
            </h1>
            <p className="text-muted-foreground mx-auto mt-4 max-w-3xl text-sm sm:text-base md:text-xl lg:text-2xl break-words">
              No Setup. No build. No Deploy. Just Code.
            </p>

            <p className="text-muted-foreground mx-auto mt-6 text-sm sm:text-base md:text-xl lg:text-2xl">
              For <TypingText />
            </p>

            {/* Visual Flow - Hidden on mobile and tablet */}
            <div className="mx-auto mt-14 w-full hidden md:block">
              <div className="flex flex-row items-center justify-center gap-3 lg:gap-4">
                <div className="bg-muted text-muted-foreground decoration-muted-foreground/60 rounded-xl border px-5 py-2.5 text-sm font-medium line-through decoration-2 lg:px-6 lg:py-3 lg:text-base text-center">
                  Setup
                </div>
                <div className="bg-border h-0.5 w-6 lg:w-10 flex-shrink-0"></div>
                <div className="border-info bg-info/10 text-info rounded-xl border-2 px-6 py-2.5 text-sm font-bold lg:px-9 lg:py-4 lg:text-lg text-center">
                  Code
                </div>
                <div className="bg-border h-0.5 w-6 lg:w-10 flex-shrink-0"></div>
                <div className="bg-muted text-muted-foreground decoration-muted-foreground/60 rounded-xl border px-5 py-2.5 text-sm font-medium line-through decoration-2 lg:px-6 lg:py-3 lg:text-base text-center">
                  Build
                </div>
                <div className="bg-border h-0.5 w-6 lg:w-10 flex-shrink-0"></div>
                <div className="bg-muted text-muted-foreground decoration-muted-foreground/60 rounded-xl border px-5 py-2.5 text-sm font-medium line-through decoration-2 lg:px-6 lg:py-3 lg:text-base text-center">
                  Deploy
                </div>
                <div className="bg-border h-0.5 w-6 lg:w-10 flex-shrink-0"></div>
                <div className="border-success bg-success/10 text-success rounded-xl border-2 px-6 py-2.5 text-sm font-bold lg:px-9 lg:py-4 lg:text-lg text-center">
                  Test
                </div>
              </div>
            </div>

            <p className="text-muted-foreground mt-8 sm:mt-12 text-sm sm:text-base px-4">
              Designed to reduce development loop
            </p>

            {/* CTAs */}
            <div className="mt-6 flex flex-col items-center justify-center gap-4 sm:flex-row px-4">
              <GetStartedButton
                size="lg"
                className="w-full px-8 text-base font-semibold sm:w-auto"
              />
              <Button
                asChild
                variant="outline"
                size="lg"
                className="w-full px-8 text-base font-semibold sm:w-auto"
              >
                <Link href="/docs">Read Documentation</Link>
              </Button>
            </div>
          </div>
        </div>
      </main>

      <WebsiteFooter />
    </div>
  )
}

// Main dashboard page - middleware ensures only users with 'user' role can access this
export default async function HomePage() {
  // If in website mode, show website landing page
  if (APP_MODE === 'website') {
    return <WebsiteLandingPage />
  }

  // Otherwise, show dashboard (for dashboard mode)
  const session = await getSession()

  // Session is guaranteed to exist due to middleware checks
  const currentUser = session!.user?.email || 'user@example.com'
  const userRoles = session!.user?.roles || []
  const isSuperAdmin = userRoles.includes('super-admin')
  const isAdmin = userRoles.includes('admin') || isSuperAdmin

  // Fetch machine types
  const machineTypesResult = await listMachineTypes()
  const availableMachineTypes =
    machineTypesResult.success && machineTypesResult.data
      ? machineTypesResult.data.items
          .filter((mt) => mt.spec.active !== false) // Only active types
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
  let workMachines: TransformedWorkMachine[] = []

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
