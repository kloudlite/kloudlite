import { NavigationWrapper } from '@/components/navigation-wrapper'
import { ScrollArea } from '@/components/ui/scroll-area'
import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'
import { isSystemReady, SystemSetupPage } from '@/lib/system-check'
import { APP_MODE } from '@/lib/app-mode'

// Main layout - handles both dashboard and website modes
export default async function MainLayout({ children }: { children: React.ReactNode }) {
  // Website mode: Simple layout without navigation with ScrollArea
  if (APP_MODE === 'website') {
    return (
      <ScrollArea className="h-screen">
        <div className="min-h-screen">{children}</div>
      </ScrollArea>
    )
  }

  // Dashboard mode: Full layout with navigation
  const session = await auth()
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
