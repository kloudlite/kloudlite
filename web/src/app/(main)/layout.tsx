import { NavigationWrapper } from '@/components/navigation-wrapper'
import { ScrollArea } from '@/components/ui/scroll-area'
import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'
import { isSystemReady, SystemSetupPage } from '@/lib/system-check'

// Main layout - middleware ensures only users with 'user' role can access this
export default async function MainLayout({
  children,
}: {
  children: React.ReactNode
}) {
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
    <div className="h-screen flex flex-col bg-background">
      <div className="flex-shrink-0">
        <NavigationWrapper />
      </div>
      <ScrollArea className="flex-1 overflow-auto">
        {children}
      </ScrollArea>
    </div>
  )
}