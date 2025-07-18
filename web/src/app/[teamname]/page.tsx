import { getTeams } from '@/actions/teams'
import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { Activity, Users, FolderOpen, Settings, BarChart3, Shield } from 'lucide-react'

interface TeamDashboardPageProps {
  params: Promise<{ teamname: string }>
}

export default async function TeamDashboardPage({ params }: TeamDashboardPageProps) {
  // Await params before accessing properties
  const { teamname } = await params
  
  // Get team data
  const teams = await getTeams()
  const team = teams.find(t => 
    t.slug === teamname || 
    t.name.toLowerCase().replace(/\s+/g, '-') === teamname
  )
  
  if (!team) {
    return null // Layout will handle 404
  }

  const quickLinks = [
    { href: `/${teamname}/projects`, label: 'Projects', icon: FolderOpen, description: 'Manage your projects' },
    { href: `/${teamname}/environments`, label: 'Environments', icon: Activity, description: 'View environments' },
    { href: `/${teamname}/members`, label: 'Members', icon: Users, description: 'Team members' },
    { href: `/${teamname}/analytics`, label: 'Analytics', icon: BarChart3, description: 'Usage & metrics' },
    { href: `/${teamname}/security`, label: 'Security', icon: Shield, description: 'Security settings' },
    { href: `/${teamname}/settings`, label: 'Settings', icon: Settings, description: 'Team settings' },
  ]

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <div className="bg-card border rounded-lg p-6 shadow-dashboard-card-shadow">
        <h2 className="text-xl font-semibold mb-2">Welcome to {team.name}</h2>
        <p className="text-muted-foreground mb-4">
          {team.description || 'Get started by creating your first project or exploring team settings.'}
        </p>
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          <span>{team.memberCount} members</span>
          <span>•</span>
          <span>Region: {team.region}</span>
          <span>•</span>
          <span>Role: {team.userRole}</span>
        </div>
      </div>

      {/* Quick Actions */}
      <div>
        <h3 className="text-lg font-semibold mb-4">Quick Actions</h3>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {quickLinks.map((link) => {
            const Icon = link.icon
            return (
              <Link
                key={link.href}
                href={link.href}
                className="group bg-card border rounded-lg p-4 hover:border-primary/40 hover:shadow-dashboard-card-shadow transition-all duration-200 hover:scale-[1.02]"
              >
                <div className="flex items-start gap-3">
                  <div className="p-2 bg-muted rounded-lg group-hover:bg-primary/10 dark:group-hover:bg-primary/20 transition-all duration-200">
                    <Icon className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors duration-200" />
                  </div>
                  <div>
                    <h4 className="font-medium group-hover:text-primary transition-colors duration-200">
                      {link.label}
                    </h4>
                    <p className="text-sm text-muted-foreground mt-0.5">
                      {link.description}
                    </p>
                  </div>
                </div>
              </Link>
            )
          })}
        </div>
      </div>

      {/* Recent Activity Placeholder */}
      <div className="bg-card border rounded-lg p-6 shadow-dashboard-card-shadow">
        <h3 className="text-lg font-semibold mb-4">Recent Activity</h3>
        <p className="text-muted-foreground">No recent activity to display.</p>
      </div>
    </div>
  )
}