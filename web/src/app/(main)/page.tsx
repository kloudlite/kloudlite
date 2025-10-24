import { auth } from '@/lib/auth'
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
    status:
      currentState === 'running'
        ? ('active' as const)
        : currentState === 'stopped'
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
    <div className="flex min-h-screen flex-col bg-background">
      {/* Navigation Header */}
      <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <nav className="mx-auto flex h-16 max-w-[90rem] items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-6 lg:gap-8">
            <KloudliteLogo showText={true} linkToHome={true} />
            <div className="hidden items-center gap-6 md:flex">
              <Link
                href="/docs"
                className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                Docs
              </Link>
              <Link
                href="/pricing"
                className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
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
            <h1 className="text-5xl font-bold leading-tight tracking-tight text-foreground sm:text-6xl lg:text-7xl">
              Platform of Development Environments
            </h1>
            <p className="mx-auto mt-6 max-w-3xl text-xl text-muted-foreground sm:text-2xl">
              No Setup. No build. No Deploy. Just Code.
            </p>

            <p className="mx-auto mt-8 text-xl text-muted-foreground sm:text-2xl">
              For <TypingText />
            </p>

            {/* Visual Flow */}
            <div className="mx-auto mt-14 flex max-w-5xl items-center justify-center gap-3 sm:gap-4">
              <div className="rounded-xl border bg-muted px-5 py-2.5 text-sm font-medium text-muted-foreground line-through decoration-muted-foreground/60 decoration-2 sm:px-6 sm:py-3 sm:text-base">
                Setup
              </div>
              <div className="h-0.5 w-6 bg-border sm:w-10"></div>
              <div className="rounded-xl border-2 border-info bg-info/10 px-7 py-3 text-base font-bold text-info sm:px-9 sm:py-4 sm:text-lg">
                Code
              </div>
              <div className="h-0.5 w-6 bg-border sm:w-10"></div>
              <div className="rounded-xl border bg-muted px-5 py-2.5 text-sm font-medium text-muted-foreground line-through decoration-muted-foreground/60 decoration-2 sm:px-6 sm:py-3 sm:text-base">
                Build
              </div>
              <div className="h-0.5 w-6 bg-border sm:w-10"></div>
              <div className="rounded-xl border bg-muted px-5 py-2.5 text-sm font-medium text-muted-foreground line-through decoration-muted-foreground/60 decoration-2 sm:px-6 sm:py-3 sm:text-base">
                Deploy
              </div>
              <div className="h-0.5 w-6 bg-border sm:w-10"></div>
              <div className="rounded-xl border-2 border-success bg-success/10 px-7 py-3 text-base font-bold text-success sm:px-9 sm:py-4 sm:text-lg">
                Test
              </div>
            </div>

            {/* CTAs */}
            <div className="mt-12 flex flex-col items-center justify-center gap-4 sm:flex-row">
              <GetStartedButton size="lg" className="w-full px-8 text-base font-semibold sm:w-auto" />
              <Button asChild variant="outline" size="lg" className="w-full px-8 text-base font-semibold sm:w-auto">
                <Link href="/docs">Read Documentation</Link>
              </Button>
            </div>

            <p className="mt-6 text-base text-muted-foreground">
              Designed to reduce development loop
            </p>
          </div>

        </div>
      </main>

      {/* Footer */}
      <footer className="border-t bg-muted">
        <div className="mx-auto max-w-[90rem] px-4 py-12 sm:px-6 lg:px-8">
          <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
            <p className="text-sm text-muted-foreground">© 2024 Kloudlite. All rights reserved.</p>
            <div className="flex items-center gap-6">
              <Link href="/docs" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
                Docs
              </Link>
              <Link href="/pricing" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
                Pricing
              </Link>
              <Link href="/contact" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
                Contact
              </Link>
              <Link href="https://github.com/kloudlite/kloudlite" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
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
  const session = await auth()

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
