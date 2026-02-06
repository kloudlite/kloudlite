import { NavigationWrapper } from '@/components/navigation-wrapper'
import { ScrollArea } from '@kloudlite/ui'
import { getSession } from '@/lib/get-session'
import { redirect } from 'next/navigation'
import { isSystemReady, SystemSetupPage } from '@/lib/system-check'
import { getMyWorkMachine } from '@/app/actions/work-machine.actions'

// Console layout - workspace management interface with full navigation
export default async function MainLayout({ children }: { children: React.ReactNode }) {
  const session = await getSession()

  // Redirect to login if not authenticated
  if (!session) {
    redirect('/auth/signin')
  }

  const userRoles = session.user?.roles || []
  const sessionProvider = (session.user as { provider?: string })?.provider
  const hasUserRole = userRoles.includes('user')
  const hasAdminRole = userRoles.includes('admin') || userRoles.includes('super-admin')
  const isSuperAdmin = userRoles.includes('super-admin')
  const isSuperAdminLogin = sessionProvider === 'superadmin-login' || isSuperAdmin

  // Super-admin logins should always go to admin section
  if (isSuperAdminLogin && hasAdminRole) {
    redirect('/admin')
  }

  // Main dashboard section - redirect admin-only users to admin
  if (!hasUserRole && hasAdminRole) {
    redirect('/admin')
  }

  // Require user role for main section
  if (!hasUserRole && !hasAdminRole) {
    redirect('/auth/signin')
  }

  // Fetch system ready status and work machine in parallel (both independent after session)
  const [systemReady, workMachineResult] = await Promise.all([
    isSystemReady(),
    getMyWorkMachine(),
  ])
  const hasWorkMachine = workMachineResult.success && !!workMachineResult.data

  let isWorkMachineRunning = false
  if (hasWorkMachine && workMachineResult.data) {
    const state = workMachineResult.data.status?.state || workMachineResult.data.spec.state
    const isReady = workMachineResult.data.status?.isReady ?? false
    isWorkMachineRunning = state === 'running' && isReady
  }

  // If system not ready
  if (!systemReady) {
    // Only super-admin can access system during setup - redirect to admin page
    if (isSuperAdmin) {
      redirect('/admin')
    }

    // All other users (including admin) see under construction page
    return <SystemSetupPage />
  }

  // No work machine — render full-screen setup page without navigation chrome
  if (!hasWorkMachine) {
    return <>{children}</>
  }

  // Normal layout with navigation
  return (
    <div className="bg-background flex h-screen flex-col">
      <NavigationWrapper
        email={session.user?.email ?? undefined}
        displayName={session.user?.name ?? undefined}
        isSuperAdmin={isSuperAdmin}
        isAdmin={hasAdminRole}
        userRoles={userRoles}
        hasWorkMachine={hasWorkMachine}
        isWorkMachineRunning={isWorkMachineRunning}
      />
      <ScrollArea className="flex-1">
        <main className="mx-auto max-w-7xl px-6 lg:px-8 py-10">
          {children}
        </main>
        <div className="h-16" />
      </ScrollArea>
    </div>
  )
}
