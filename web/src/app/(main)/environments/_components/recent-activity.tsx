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
      <div className="rounded-lg border border-gray-200 bg-white">
        <div className="border-b border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium text-gray-900">Recent Workspaces</h3>
            <Link href="/workspaces" className="text-sm text-blue-600 hover:text-blue-700">
              View all →
            </Link>
          </div>
        </div>
        <div className="divide-y divide-gray-200">
          {workspaces.map((workspace) => (
            <Link
              key={workspace.id}
              href={`/workspaces/${workspace.id}`}
              className="block p-4 transition-colors hover:bg-gray-50"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-3">
                  <div className="mt-0.5 rounded-lg bg-gray-100 p-2">
                    {getActionIcon(workspace.action)}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="text-sm font-medium text-gray-900">{workspace.name}</h4>
                      <span
                        className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                          workspace.status === 'active'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-600'
                        }`}
                      >
                        {workspace.status}
                      </span>
                    </div>
                    <p className="mt-1 text-xs text-gray-600">{workspace.action}</p>
                    <div className="mt-2 flex items-center gap-4">
                      <span className="flex items-center gap-1 text-xs text-gray-500">
                        <Globe className="h-3 w-3" />
                        {workspace.environment}
                      </span>
                      <span className="text-xs text-gray-500">{workspace.lastActivity}</span>
                    </div>
                  </div>
                </div>
              </div>
            </Link>
          ))}
        </div>
      </div>

      {/* Recent Environments */}
      <div className="rounded-lg border border-gray-200 bg-white">
        <div className="border-b border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium text-gray-900">Recent Environments</h3>
            <Link href="/environments" className="text-sm text-blue-600 hover:text-blue-700">
              View all →
            </Link>
          </div>
        </div>
        <div className="divide-y divide-gray-200">
          {environments.map((env) => (
            <Link
              key={env.id}
              href={`/environments/${env.id}`}
              className="block p-4 transition-colors hover:bg-gray-50"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-3">
                  <div className="mt-0.5 rounded-lg bg-gray-100 p-2">
                    {getActionIcon(env.action)}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="text-sm font-medium text-gray-900">{env.name}</h4>
                      <span
                        className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                          env.status === 'active'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-600'
                        }`}
                      >
                        {env.status}
                      </span>
                    </div>
                    <p className="mt-1 text-xs text-gray-600">{env.action}</p>
                    <div className="mt-2 flex items-center gap-4">
                      <span className="text-xs text-gray-500">{env.services} services</span>
                      <span className="text-xs text-gray-500">{env.workspaces} workspaces</span>
                      <span className="text-xs text-gray-500">{env.lastActivity}</span>
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
