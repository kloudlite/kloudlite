import { NavigationWrapper } from '@/components/navigation-wrapper'
import { ScrollArea } from '@kloudlite/ui'
import { getSession } from '@/lib/get-session'
import { redirect } from 'next/navigation'
import { isSystemReady, SystemSetupPage } from '@/lib/system-check'

// Console layout - workspace management interface with full navigation
export default async function MainLayout({ children }: { children: React.ReactNode }) {
  const session = await getSession()
  const userRoles = session?.user?.roles || []
  const isSuperAdmin = userRoles.includes('super-admin')

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

  return (
    <div className="bg-background flex h-screen flex-col">
      <div className="flex-shrink-0">
        <NavigationWrapper />
      </div>
      <ScrollArea className="flex-1 overflow-auto">{children}</ScrollArea>
    </div>
  )
}
