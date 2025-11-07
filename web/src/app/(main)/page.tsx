import { getSession } from '@/lib/get-session'
import { WorkMachinesContent } from './workspaces/_components/work-machines-content'
import { getMyWorkMachine, listAllWorkMachines } from '@/app/actions/work-machine.actions'
import { listMachineTypes } from '@/app/actions/machine-type.actions'
import type { WorkMachine } from '@/types/work-machine'
import { APP_MODE } from '@/lib/app-mode'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { TypingText } from './_components/typing-text'
import { ThemeSwitcherServer } from '@/components/theme-switcher-server'
import { GetStartedButton } from '@/components/get-started-button'

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
    <div className="bg-background flex min-h-screen flex-col">
      {/* Navigation Header */}
      <header className="bg-background/95 supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50 border-b backdrop-blur">
        <nav className="mx-auto flex h-16 max-w-[90rem] items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-6 lg:gap-8">
            <KloudliteLogo showText={true} linkToHome={true} />
            <div className="hidden items-center gap-6 md:flex">
              <Link
                href="/docs"
                className="text-muted-foreground hover:text-foreground text-sm font-medium transition-colors"
              >
                Docs
              </Link>
              <Link
                href="/pricing"
                className="text-muted-foreground hover:text-foreground text-sm font-medium transition-colors"
              >
                Pricing
              </Link>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <GetStartedButton size="sm" className="hidden sm:flex" />
          </div>
        </nav>
      </header>

      {/* Hero Section */}
      <main className="flex flex-1 items-center px-4 py-16 sm:px-6 lg:px-8">
        <div className="mx-auto w-full max-w-[90rem]">
          {/* Main Heading */}
          <div className="text-center">
            <h1 className="text-foreground text-5xl leading-tight font-bold tracking-tight sm:text-6xl lg:text-7xl">
              Platform of Development Environments
            </h1>
            <p className="text-muted-foreground mx-auto mt-6 max-w-3xl text-xl sm:text-2xl">
              No Setup. No build. No Deploy. Just Code.
            </p>

            <p className="text-muted-foreground mx-auto mt-8 text-xl sm:text-2xl">
              For <TypingText />
            </p>

            {/* Visual Flow */}
            <div className="mx-auto mt-14 flex max-w-5xl items-center justify-center gap-3 sm:gap-4">
              <div className="bg-muted text-muted-foreground decoration-muted-foreground/60 rounded-xl border px-5 py-2.5 text-sm font-medium line-through decoration-2 sm:px-6 sm:py-3 sm:text-base">
                Setup
              </div>
              <div className="bg-border h-0.5 w-6 sm:w-10"></div>
              <div className="border-info bg-info/10 text-info rounded-xl border-2 px-7 py-3 text-base font-bold sm:px-9 sm:py-4 sm:text-lg">
                Code
              </div>
              <div className="bg-border h-0.5 w-6 sm:w-10"></div>
              <div className="bg-muted text-muted-foreground decoration-muted-foreground/60 rounded-xl border px-5 py-2.5 text-sm font-medium line-through decoration-2 sm:px-6 sm:py-3 sm:text-base">
                Build
              </div>
              <div className="bg-border h-0.5 w-6 sm:w-10"></div>
              <div className="bg-muted text-muted-foreground decoration-muted-foreground/60 rounded-xl border px-5 py-2.5 text-sm font-medium line-through decoration-2 sm:px-6 sm:py-3 sm:text-base">
                Deploy
              </div>
              <div className="bg-border h-0.5 w-6 sm:w-10"></div>
              <div className="border-success bg-success/10 text-success rounded-xl border-2 px-7 py-3 text-base font-bold sm:px-9 sm:py-4 sm:text-lg">
                Test
              </div>
            </div>

            {/* CTAs */}
            <div className="mt-12 flex flex-col items-center justify-center gap-4 sm:flex-row">
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

            <p className="text-muted-foreground mt-6 text-base">
              Designed to reduce development loop
            </p>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-muted border-t">
        <div className="mx-auto max-w-[90rem] px-4 py-12 sm:px-6 lg:px-8">
          <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
            <p className="text-muted-foreground text-sm">© 2024 Kloudlite. All rights reserved.</p>
            <div className="flex items-center gap-6">
              <Link
                href="/docs"
                className="text-muted-foreground hover:text-foreground text-sm transition-colors"
              >
                Docs
              </Link>
              <Link
                href="/pricing"
                className="text-muted-foreground hover:text-foreground text-sm transition-colors"
              >
                Pricing
              </Link>
              <Link
                href="/contact"
                className="text-muted-foreground hover:text-foreground text-sm transition-colors"
              >
                Contact
              </Link>
              <Link
                href="https://github.com/kloudlite/kloudlite"
                className="text-muted-foreground hover:text-foreground text-sm transition-colors"
              >
                GitHub
              </Link>
              <ThemeSwitcherServer />
            </div>
          </div>
        </div>
      </footer>
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
