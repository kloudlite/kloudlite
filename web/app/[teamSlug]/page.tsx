import { redirect } from "next/navigation"
import { getServerSession } from "next-auth"
import { 
  Cloud, 
  Box, 
  Database, 
  Activity,
  Users,
  Settings2,
  Plus,
  TrendingUp,
  Server
} from "lucide-react"

import { getAuthOptions } from "@/lib/auth/get-auth-options"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"

interface TeamPageProps {
  params: Promise<{ teamSlug: string }>
}

export default async function TeamPage({ params }: TeamPageProps) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  const { teamSlug } = await params

  if (!session) {
    redirect(`/auth/login?callbackUrl=/${teamSlug}`)
  }

  if (!session.user.emailVerified) {
    redirect("/auth/email-verification-required")
  }

  // TODO: Fetch actual team data and statistics
  const stats = {
    environments: { count: 3, limit: 10, active: 2 },
    workspaces: { count: 5, limit: 20, active: 3 },
    services: { count: 8, limit: 50, active: 6 },
    members: { count: 12, active: 10 }
  }

  const recentActivity = [
    { id: 1, action: "Environment deployed", resource: "production", time: "2 hours ago", user: "John Doe" },
    { id: 2, action: "Workspace created", resource: "dev-workspace-3", time: "5 hours ago", user: "Jane Smith" },
    { id: 3, action: "Database provisioned", resource: "postgres-main", time: "1 day ago", user: "Mike Wilson" },
  ]

  return (
    <div className="space-y-8">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl md:text-3xl font-extralight tracking-tight">Team Overview</h1>
        <p className="text-sm md:text-base text-muted-foreground mt-2">
          Manage your team's cloud resources and infrastructure
        </p>
      </div>

      {/* Quick Actions */}
      <div className="flex flex-wrap gap-2">
        <Button size="sm" className="gap-2">
          <Plus className="h-4 w-4" />
          <span className="hidden sm:inline">New Environment</span>
          <span className="sm:hidden">Environment</span>
        </Button>
        <Button size="sm" variant="outline" className="gap-2">
          <Box className="h-4 w-4" />
          <span className="hidden sm:inline">Create Workspace</span>
          <span className="sm:hidden">Workspace</span>
        </Button>
        <Button size="sm" variant="outline" className="gap-2">
          <Database className="h-4 w-4" />
          <span className="hidden sm:inline">Add Service</span>
          <span className="sm:hidden">Service</span>
        </Button>
      </div>

      {/* Resource Statistics */}
      <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Environments
              </CardTitle>
              <Cloud className="h-4 w-4 text-muted-foreground" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.environments.count}</div>
            <div className="mt-3 space-y-2">
              <Progress value={(stats.environments.count / stats.environments.limit) * 100} className="h-1" />
              <p className="text-xs text-muted-foreground">
                {stats.environments.active} active • {stats.environments.limit - stats.environments.count} available
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Workspaces
              </CardTitle>
              <Box className="h-4 w-4 text-muted-foreground" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.workspaces.count}</div>
            <div className="mt-3 space-y-2">
              <Progress value={(stats.workspaces.count / stats.workspaces.limit) * 100} className="h-1" />
              <p className="text-xs text-muted-foreground">
                {stats.workspaces.active} active • {stats.workspaces.limit - stats.workspaces.count} available
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Shared Services
              </CardTitle>
              <Database className="h-4 w-4 text-muted-foreground" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.services.count}</div>
            <div className="mt-3 space-y-2">
              <Progress value={(stats.services.count / stats.services.limit) * 100} className="h-1" />
              <p className="text-xs text-muted-foreground">
                {stats.services.active} running • {stats.services.limit - stats.services.count} available
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Team Members
              </CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.members.count}</div>
            <div className="mt-3">
              <p className="text-xs text-muted-foreground">
                {stats.members.active} active • 2 pending invites
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content Grid */}
      <div className="grid gap-4 lg:gap-6 lg:grid-cols-3">
        {/* Recent Activity */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Recent Activity</CardTitle>
                <CardDescription>Latest actions in your team</CardDescription>
              </div>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {recentActivity.map((activity) => (
                <div key={activity.id} className="flex items-start gap-4 pb-4 last:pb-0 last:border-0 border-b">
                  <div className="h-8 w-8 rounded-full bg-muted flex items-center justify-center">
                    <Activity className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <div className="flex-1 space-y-1">
                    <p className="text-sm">
                      <span className="font-medium">{activity.user}</span>
                      <span className="text-muted-foreground"> {activity.action}</span>
                    </p>
                    <p className="text-sm font-medium text-primary">{activity.resource}</p>
                    <p className="text-xs text-muted-foreground">{activity.time}</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Quick Links */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Quick Links</CardTitle>
                <CardDescription>Common actions and resources</CardDescription>
              </div>
              <Settings2 className="h-4 w-4 text-muted-foreground" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <Button variant="outline" className="w-full justify-start gap-3" size="sm">
                <Cloud className="h-4 w-4" />
                Deploy to Production
              </Button>
              <Button variant="outline" className="w-full justify-start gap-3" size="sm">
                <Server className="h-4 w-4" />
                View Infrastructure
              </Button>
              <Button variant="outline" className="w-full justify-start gap-3" size="sm">
                <TrendingUp className="h-4 w-4" />
                Resource Usage
              </Button>
              <Button variant="outline" className="w-full justify-start gap-3" size="sm">
                <Users className="h-4 w-4" />
                Invite Team Member
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}