import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { KloudliteLogo } from '@kloudlite/ui'
import { AdminNavigation } from './_components/admin-navigation'
import { AdminProfileDropdown } from './_components/admin-profile-dropdown'
import { isSystemReady, SystemSetupPage } from '@/lib/system-check'

// Admin layout - only users with admin/super-admin roles can access
export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const session = await getSession()

  // Redirect to login if not authenticated
  if (!session) {
    redirect('/auth/signin')
  }

  const userRoles = session.user?.roles || []
  const hasUserRole = userRoles.includes('user')
  const hasAdminRole = userRoles.includes('admin') || userRoles.includes('super-admin')
  const isSuperAdmin = userRoles.includes('super-admin')

  // Admin section - only allow admin/super-admin access
  if (!hasAdminRole) {
    redirect('/')
  }

  // Check if system is configured
  const systemReady = await isSystemReady()

  // If system not ready and not super-admin, show under construction
  if (!systemReady && !isSuperAdmin) {
    return <SystemSetupPage />
  }

  return (
    <div className="bg-background min-h-screen">
      {/* Admin Header */}
      <header className="border-b bg-background">
        <div className="mx-auto max-w-7xl px-6">
          <div className="flex h-16 items-center justify-between">
            {/* Logo / Brand and Navigation */}
            <div className="flex items-center gap-8">
              <div className="flex items-center gap-3">
                <KloudliteLogo className="text-lg font-medium" />
                <span className="text-muted-foreground text-lg font-medium">Admin</span>
              </div>

              {/* Admin Navigation */}
              <AdminNavigation />
            </div>

            {/* User Dropdown */}
            <AdminProfileDropdown
              name={session.user?.name}
              email={session.user?.email}
              hasUserRole={hasUserRole}
            />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main>{children}</main>
    </div>
  )
}
