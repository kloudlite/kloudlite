import { Cloud, Plus, Search, Filter } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

interface EnvironmentsPageProps {
  params: Promise<{ teamSlug: string }>
}

export default async function EnvironmentsPage({ params }: EnvironmentsPageProps) {
  const { teamSlug } = await params

  // TODO: Fetch actual environments data
  const environments = [
    {
      id: "1",
      name: "Production",
      status: "running",
      region: "us-west-2",
      resources: 12,
      lastDeployed: "2 hours ago",
      cost: "$450/month"
    },
    {
      id: "2",
      name: "Staging",
      status: "running",
      region: "us-west-2",
      resources: 8,
      lastDeployed: "1 day ago",
      cost: "$230/month"
    },
    {
      id: "3",
      name: "Development",
      status: "stopped",
      region: "us-east-1",
      resources: 5,
      lastDeployed: "3 days ago",
      cost: "$120/month"
    }
  ]

  return (
    <div className="space-y-4 md:space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-extralight tracking-tight">Environments</h1>
          <p className="text-sm md:text-base text-muted-foreground mt-2">
            Deploy and manage your application environments
          </p>
        </div>
        <Button className="gap-2 w-full sm:w-auto">
          <Plus className="h-4 w-4" />
          New Environment
        </Button>
      </div>

      {/* Search and Filters */}
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="Search environments..." className="pl-9" />
        </div>
        <Button variant="outline" size="icon">
          <Filter className="h-4 w-4" />
        </Button>
      </div>

      {/* Environments Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {environments.map((env) => (
          <Card key={env.id} className="hover:border-primary/50 transition-colors cursor-pointer">
            <CardHeader>
              <div className="flex items-start justify-between">
                <div className="space-y-1">
                  <CardTitle className="text-lg">{env.name}</CardTitle>
                  <CardDescription>{env.region}</CardDescription>
                </div>
                <Badge 
                  variant={env.status === "running" ? "default" : "secondary"}
                  className="capitalize"
                >
                  {env.status}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Resources</span>
                  <span className="font-medium">{env.resources}</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Cost</span>
                  <span className="font-medium">{env.cost}</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">Last deployed</span>
                  <span className="font-medium">{env.lastDeployed}</span>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}