import { Box, Plus, Search, Filter, Terminal, Globe, Cpu } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"

interface WorkspacesPageProps {
  params: Promise<{ teamSlug: string }>
}

export default async function WorkspacesPage({ params }: WorkspacesPageProps) {
  const { teamSlug } = await params

  // TODO: Fetch actual workspaces data
  const workspaces = [
    {
      id: "1",
      name: "frontend-dev",
      status: "active",
      owner: { name: "John Doe", email: "john@example.com", avatar: "" },
      type: "vscode",
      resources: { cpu: "2 vCPU", memory: "4GB RAM", storage: "20GB" },
      lastActive: "Currently active",
      environment: "Development"
    },
    {
      id: "2",
      name: "backend-api",
      status: "active",
      owner: { name: "Jane Smith", email: "jane@example.com", avatar: "" },
      type: "terminal",
      resources: { cpu: "4 vCPU", memory: "8GB RAM", storage: "50GB" },
      lastActive: "2 hours ago",
      environment: "Staging"
    },
    {
      id: "3",
      name: "data-analysis",
      status: "stopped",
      owner: { name: "Mike Wilson", email: "mike@example.com", avatar: "" },
      type: "jupyter",
      resources: { cpu: "8 vCPU", memory: "16GB RAM", storage: "100GB" },
      lastActive: "1 day ago",
      environment: "Development"
    }
  ]

  const getWorkspaceIcon = (type: string) => {
    switch(type) {
      case "terminal": return <Terminal className="h-4 w-4" />
      case "jupyter": return <Globe className="h-4 w-4" />
      default: return <Box className="h-4 w-4" />
    }
  }

  return (
    <div className="space-y-4 md:space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-extralight tracking-tight">Workspaces</h1>
          <p className="text-sm md:text-base text-muted-foreground mt-2">
            Cloud development environments for your team
          </p>
        </div>
        <Button className="gap-2 w-full sm:w-auto">
          <Plus className="h-4 w-4" />
          <span className="hidden sm:inline">New Workspace</span>
          <span className="sm:hidden">New</span>
        </Button>
      </div>

      {/* Search and Filters */}
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="Search workspaces..." className="pl-9" />
        </div>
        <Button variant="outline" size="icon">
          <Filter className="h-4 w-4" />
        </Button>
      </div>

      {/* Workspaces Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {workspaces.map((workspace) => (
          <Card key={workspace.id} className="hover:border-primary/50 transition-colors cursor-pointer">
            <CardHeader>
              <div className="flex items-start justify-between">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    {getWorkspaceIcon(workspace.type)}
                    <CardTitle className="text-lg">{workspace.name}</CardTitle>
                  </div>
                  <CardDescription>{workspace.environment}</CardDescription>
                </div>
                <Badge 
                  variant={workspace.status === "active" ? "default" : "secondary"}
                  className="capitalize"
                >
                  {workspace.status}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {/* Owner */}
                <div className="flex items-center gap-3">
                  <Avatar className="h-8 w-8">
                    <AvatarImage src={workspace.owner.avatar} />
                    <AvatarFallback>
                      {workspace.owner.name.split(' ').map(n => n[0]).join('')}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">{workspace.owner.name}</p>
                    <p className="text-xs text-muted-foreground">{workspace.lastActive}</p>
                  </div>
                </div>
                
                {/* Resources */}
                <div className="flex flex-wrap items-center gap-2 sm:gap-4 text-xs text-muted-foreground">
                  <div className="flex items-center gap-1">
                    <Cpu className="h-3 w-3" />
                    {workspace.resources.cpu}
                  </div>
                  <div>{workspace.resources.memory}</div>
                  <div>{workspace.resources.storage}</div>
                </div>

                {/* Actions */}
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" className="flex-1">
                    Open
                  </Button>
                  <Button size="sm" variant="ghost">
                    Settings
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}