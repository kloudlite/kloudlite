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
    <div className="flex h-full flex-col bg-dashboard-sidebar">
        {/* Header - Team Section */}
        <header className="flex-shrink-0 bg-gradient-to-b from-muted/80 to-muted/40 border-b border-border shadow-dashboard-card-shadow">
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
        <section className="flex-shrink-0 bg-dashboard-section border-t border-border">
          <SimpleWorkMachine className="px-4 py-5" />
        </section>

        {/* Footer - Always visible */}
        <footer className="flex-shrink-0 bg-gradient-to-t from-muted/80 to-muted/40 border-t border-border">
          {/* Bottom Action Bar with Profile */}
          <div className="px-4 py-3 flex items-center justify-between">
            <SidebarActions />
            {user && (
              <UserProfileDropdown 
                variant="default" 
                userRole="owner" 
                user={user} 
              />
            )}
          </div>
        </footer>
      </div>
  )
}