import { getTeams } from '@/actions/teams'
import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { Activity, Users, FolderOpen, Settings, BarChart3, Shield, TrendingUp, Clock, AlertCircle } from 'lucide-react'
import { ActionGrid, MetricCard } from '@/components/ui/action-grid'
import { GridLayout, Card } from '@/components/layout/grid-layout'

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

  // Mock data for demonstration - in real app, fetch from API
  const teamMetrics = {
    projects: { total: 12, active: 8, inactive: 4 },
    environments: { total: 24, running: 18, stopped: 6 },
    activity: { deploymentsToday: 5, lastDeployment: '2 hours ago' },
    resources: { cpu: 65, memory: 78, storage: 45 }
  }

  const quickActions = [
    { 
      href: `/${teamname}/projects`, 
      label: 'Projects', 
      icon: FolderOpen, 
      description: 'Manage your projects',
      metric: `${teamMetrics.projects.total} total`,
      status: `${teamMetrics.projects.active} active`
    },
    { 
      href: `/${teamname}/environments`, 
      label: 'Environments', 
      icon: Activity, 
      description: 'View environments',
      metric: `${teamMetrics.environments.total} total`,
      status: `${teamMetrics.environments.running} running`
    },
    { 
      href: `/${teamname}/members`, 
      label: 'Members', 
      icon: Users, 
      description: 'Team members',
      metric: `${team.memberCount} members`,
      status: 'Active team'
    },
    { 
      href: `/${teamname}/analytics`, 
      label: 'Analytics', 
      icon: BarChart3, 
      description: 'Usage & metrics',
      metric: '65% CPU usage',
      status: 'Normal load'
    },
    { 
      href: `/${teamname}/security`, 
      label: 'Security', 
      icon: Shield, 
      description: 'Security settings',
      metric: 'All secure',
      status: 'No issues'
    },
    { 
      href: `/${teamname}/settings`, 
      label: 'Settings', 
      icon: Settings, 
      description: 'Team settings',
      metric: 'Configure team',
      status: 'Up to date'
    },
  ]

  return (
    <div className="space-y-8">
      {/* Welcome Section with Metrics */}
      <GridLayout columns={2}>
        <Card padding="large">
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
        </Card>

        <GridLayout columns={1} className="space-y-4">
          <MetricCard
            title="Deployments Today"
            value={teamMetrics.activity.deploymentsToday}
            subtitle="Active deployment pipeline"
            icon={TrendingUp}
            trend="up"
          />
          <MetricCard
            title="Resource Usage"
            value={`${teamMetrics.resources.cpu}%`}
            subtitle="CPU utilization"
            icon={AlertCircle}
            trend="neutral"
          />
        </GridLayout>
      </GridLayout>

      {/* Quick Actions with Metrics */}
      <ActionGrid 
        title="Quick Actions"
        actions={quickActions}
        columns={3}
      />

      {/* Recent Activity */}
      <Card padding="large">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Recent Activity</h3>
          <Button variant="ghost" size="sm" asChild>
            <Link href={`/${teamname}/analytics`}>View All</Link>
          </Button>
        </div>
        <div className="space-y-3">
          {[
            { action: 'Deployed', target: 'frontend-app', time: '2 hours ago', user: 'John Doe', status: 'success' },
            { action: 'Updated', target: 'backend-service', time: '4 hours ago', user: 'Jane Smith', status: 'success' },
            { action: 'Created', target: 'new-environment', time: '1 day ago', user: 'Mike Johnson', status: 'info' },
            { action: 'Deleted', target: 'old-database', time: '2 days ago', user: 'Sarah Wilson', status: 'warning' },
          ].map((activity, index) => (
            <div key={index} className="flex items-center gap-3 p-3 rounded-sm bg-muted/30">
              <div className={`w-2 h-2 rounded-full flex-shrink-0 ${
                activity.status === 'success' ? 'bg-green-600' :
                activity.status === 'warning' ? 'bg-orange-600' :
                'bg-blue-600'
              }`} />
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium">
                  {activity.action} <span className="font-mono text-xs bg-muted px-1.5 py-0.5 rounded">{activity.target}</span>
                </p>
                <p className="text-xs text-muted-foreground">
                  by {activity.user} • {activity.time}
                </p>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )
}