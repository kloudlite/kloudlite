'use client'

import Link from 'next/link'
import { Code2, Globe, Activity, Terminal, Settings, Package, Cloud } from 'lucide-react'

interface RecentWorkspace {
  id: string
  name: string
  environment: string
  status: 'active' | 'idle'
  lastActivity: string
  action: string
}

interface RecentEnvironment {
  id: string
  name: string
  status: 'active' | 'idle'
  lastActivity: string
  action: string
  services: number
  workspaces: number
}

interface RecentActivityProps {
  workspaces: RecentWorkspace[]
  environments: RecentEnvironment[]
}

export function RecentActivity({ workspaces, environments }: RecentActivityProps) {
  const getActionIcon = (action: string) => {
    if (action.includes('VS Code') || action.includes('Code')) return <Code2 className="h-4 w-4" />
    if (action.includes('Terminal')) return <Terminal className="h-4 w-4" />
    if (action.includes('deployed')) return <Package className="h-4 w-4" />
    if (action.includes('Configuration')) return <Settings className="h-4 w-4" />
    if (action.includes('scaled')) return <Cloud className="h-4 w-4" />
    return <Activity className="h-4 w-4" />
  }

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      {/* Recent Workspaces */}
      <div className="bg-card rounded-lg border">
        <div className="border-b p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium">Recent Workspaces</h3>
            <Link href="/workspaces" className="text-info hover:text-info/80 text-sm">
              View all →
            </Link>
          </div>
        </div>
        <div className="divide-y">
          {workspaces.map((workspace) => (
            <Link
              key={workspace.id}
              href={`/workspaces/${workspace.id}`}
              className="block p-4 transition-colors hover:bg-muted/50"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-3">
                  <div className="bg-muted mt-0.5 rounded-lg p-2">
                    {getActionIcon(workspace.action)}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="text-sm font-medium">{workspace.name}</h4>
                      <span
                        className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                          workspace.status === 'active'
                            ? 'bg-success/10 text-success dark:bg-success/20'
                            : 'bg-muted text-muted-foreground'
                        }`}
                      >
                        {workspace.status}
                      </span>
                    </div>
                    <p className="text-muted-foreground mt-1 text-xs">{workspace.action}</p>
                    <div className="mt-2 flex items-center gap-4">
                      <span className="text-muted-foreground flex items-center gap-1 text-xs">
                        <Globe className="h-3 w-3" />
                        {workspace.environment}
                      </span>
                      <span className="text-muted-foreground text-xs">{workspace.lastActivity}</span>
                    </div>
                  </div>
                </div>
              </div>
            </Link>
          ))}
        </div>
      </div>

      {/* Recent Environments */}
      <div className="bg-card rounded-lg border">
        <div className="border-b p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium">Recent Environments</h3>
            <Link href="/environments" className="text-info hover:text-info/80 text-sm">
              View all →
            </Link>
          </div>
        </div>
        <div className="divide-y divide-gray-200">
          {environments.map((env) => (
            <Link
              key={env.id}
              href={`/environments/${env.id}`}
              className="block p-4 transition-colors hover:bg-muted/50"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-3">
                  <div className="bg-muted mt-0.5 rounded-lg p-2">
                    {getActionIcon(env.action)}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="text-sm font-medium">{env.name}</h4>
                      <span
                        className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                          env.status === 'active'
                            ? 'bg-success/10 text-success dark:bg-success/20'
                            : 'bg-muted text-muted-foreground'
                        }`}
                      >
                        {env.status}
                      </span>
                    </div>
                    <p className="text-muted-foreground mt-1 text-xs">{env.action}</p>
                    <div className="mt-2 flex items-center gap-4">
                      <span className="text-muted-foreground text-xs">{env.services} services</span>
                      <span className="text-muted-foreground text-xs">{env.workspaces} workspaces</span>
                      <span className="text-muted-foreground text-xs">{env.lastActivity}</span>
                    </div>
                  </div>
                </div>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </div>
  )
}
