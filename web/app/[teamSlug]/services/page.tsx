import { Database, Plus, Search, Filter, Server, HardDrive, Zap, Shield } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"

interface ServicesPageProps {
  params: Promise<{ teamSlug: string }>
}

export default async function ServicesPage({ params }: ServicesPageProps) {
  const { teamSlug } = await params

  // TODO: Fetch actual services data
  const services = [
    {
      id: "1",
      name: "postgres-main",
      type: "PostgreSQL",
      category: "database",
      status: "healthy",
      version: "15.2",
      resources: { cpu: 25, memory: 60, storage: 80 },
      connections: 12,
      uptime: "99.9%"
    },
    {
      id: "2",
      name: "redis-cache",
      type: "Redis",
      category: "cache",
      status: "healthy",
      version: "7.0",
      resources: { cpu: 10, memory: 30, storage: 20 },
      connections: 45,
      uptime: "100%"
    },
    {
      id: "3",
      name: "mongodb-analytics",
      type: "MongoDB",
      category: "database",
      status: "degraded",
      version: "6.0",
      resources: { cpu: 60, memory: 75, storage: 90 },
      connections: 8,
      uptime: "98.5%"
    },
    {
      id: "4",
      name: "elasticsearch-logs",
      type: "Elasticsearch",
      category: "search",
      status: "healthy",
      version: "8.8",
      resources: { cpu: 40, memory: 50, storage: 70 },
      connections: 5,
      uptime: "99.7%"
    }
  ]

  const getServiceIcon = (category: string) => {
    switch(category) {
      case "database": return <Database className="h-4 w-4" />
      case "cache": return <Zap className="h-4 w-4" />
      case "search": return <HardDrive className="h-4 w-4" />
      default: return <Server className="h-4 w-4" />
    }
  }

  const getStatusColor = (status: string) => {
    switch(status) {
      case "healthy": return "default"
      case "degraded": return "destructive"
      default: return "secondary"
    }
  }

  return (
    <div className="space-y-4 md:space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-extralight tracking-tight">Shared Services</h1>
          <p className="text-sm md:text-base text-muted-foreground mt-2">
            Managed databases, caches, and infrastructure services
          </p>
        </div>
        <Button className="gap-2 w-full sm:w-auto">
          <Plus className="h-4 w-4" />
          <span className="hidden sm:inline">Add Service</span>
          <span className="sm:hidden">Add</span>
        </Button>
      </div>

      {/* Search and Filters */}
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="Search services..." className="pl-9" />
        </div>
        <Button variant="outline" size="icon">
          <Filter className="h-4 w-4" />
        </Button>
      </div>

      {/* Service Categories Summary */}
      <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Databases</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">2</div>
            <p className="text-xs text-muted-foreground">Active instances</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Caches</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">1</div>
            <p className="text-xs text-muted-foreground">Active instances</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Search</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">1</div>
            <p className="text-xs text-muted-foreground">Active instances</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Cost</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">$780</div>
            <p className="text-xs text-muted-foreground">Per month</p>
          </CardContent>
        </Card>
      </div>

      {/* Services Grid */}
      <div className="grid gap-4 lg:grid-cols-2">
        {services.map((service) => (
          <Card key={service.id} className="hover:border-primary/50 transition-colors cursor-pointer">
            <CardHeader>
              <div className="flex items-start justify-between">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    {getServiceIcon(service.category)}
                    <CardTitle className="text-lg">{service.name}</CardTitle>
                  </div>
                  <CardDescription>
                    {service.type} • v{service.version}
                  </CardDescription>
                </div>
                <Badge 
                  variant={getStatusColor(service.status) as any}
                  className="capitalize"
                >
                  {service.status}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {/* Resource Usage */}
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs md:text-sm">
                    <span className="text-muted-foreground">CPU</span>
                    <span className="font-medium">{service.resources.cpu}%</span>
                  </div>
                  <Progress value={service.resources.cpu} className="h-1" />
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs md:text-sm">
                    <span className="text-muted-foreground">Memory</span>
                    <span className="font-medium">{service.resources.memory}%</span>
                  </div>
                  <Progress value={service.resources.memory} className="h-1" />
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs md:text-sm">
                    <span className="text-muted-foreground">Storage</span>
                    <span className="font-medium">{service.resources.storage}%</span>
                  </div>
                  <Progress value={service.resources.storage} className="h-1" />
                </div>

                {/* Stats */}
                <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 pt-2 border-t">
                  <div className="flex flex-wrap items-center gap-2 sm:gap-4 text-xs md:text-sm text-muted-foreground">
                    <span>{service.connections} connections</span>
                    <span>{service.uptime} uptime</span>
                  </div>
                  <Button size="sm" variant="outline" className="w-full sm:w-auto">
                    Manage
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