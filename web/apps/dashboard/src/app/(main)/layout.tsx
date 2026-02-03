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

  // Check if system is configured
  const systemReady = await isSystemReady()

  // If system not ready
  if (!systemReady) {
    // Only super-admin can access system during setup - redirect to admin page
    if (isSuperAdmin) {
      redirect('/admin')
    }

    // All other users (including admin) see under construction page
    return <SystemSetupPage />
  }

  // Check if user has a work machine
  const workMachineResult = await getMyWorkMachine()
  const hasWorkMachine = workMachineResult.success && !!workMachineResult.data

  // If no work machine, render children full-screen without navigation
  if (!hasWorkMachine) {
    return (
      <div className="bg-background flex h-screen flex-col">
        {children}
      </div>
    )
  }

  // Normal layout with navigation for users with work machines
  return (
    <div className="bg-background flex h-screen flex-col">
      <NavigationWrapper />
      <ScrollArea className="flex-1">
        <main className="mx-auto max-w-7xl px-6 lg:px-8 py-10">
          {children}
        </main>
        {/* Footer spacer for better visual balance */}
        <div className="h-16" />
      </ScrollArea>
    </div>
  )
}
