import { auth } from '@/auth'
import { TeamSwitcher } from '@/components/dashboard/sidebar/team-switcher'
import { TeamStats } from '@/components/dashboard/sidebar/team-stats'
import { SimpleWorkMachine } from '@/components/dashboard/sidebar/simple-work-machine'
import { NavSection } from '@/components/dashboard/sidebar/nav-section'
import { SidebarActions } from '@/components/dashboard/sidebar/sidebar-actions'
import { UserProfileDropdown } from '@/components/teams/user-profile-dropdown'

interface DashboardSidebarProps {
  teamSlug: string
  teamName: string
}

export async function DashboardSidebar({ teamSlug, teamName }: DashboardSidebarProps) {
  const session = await auth()
  const user = session?.user
  
  return (
    <div className="flex h-full flex-col bg-sidebar">
        {/* Header - Team Section */}
        <header className="flex-shrink-0 bg-gradient-to-b from-gray-50 to-gray-100/50 dark:from-gray-900/50 dark:to-gray-900/30 border-b border-gray-200 dark:border-gray-700">
          <div className="px-4 py-4 pb-2">
            <TeamSwitcher teamSlug={teamSlug} teamName={teamName} />
          </div>
          <div>
            <TeamStats teamSlug={teamSlug} />
          </div>
        </header>

        {/* Main Navigation - Only this scrolls if needed */}
        <NavSection teamSlug={teamSlug} />

        {/* Work Machine Section - Moved to bottom */}
        <section className="flex-shrink-0 bg-white dark:bg-gray-800/50 border-t border-gray-200 dark:border-gray-700">
          <SimpleWorkMachine className="px-4 py-5" />
        </section>

        {/* Footer - Always visible */}
        <footer className="flex-shrink-0 bg-gradient-to-t from-gray-50 to-gray-100/50 dark:from-gray-900/50 dark:to-gray-900/30 border-t border-gray-200 dark:border-gray-700">
          {/* Profile Section */}
          <div className="px-4 py-4 border-b border-gray-100 dark:border-gray-800/50">
            {user && <UserProfileDropdown variant="sidebar" userRole="owner" user={user} />}
          </div>
          
          {/* Bottom Action Bar */}
          <div className="px-4 py-3">
            <SidebarActions />
          </div>
        </footer>
      </div>
  )
}