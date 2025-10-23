'use client'

import { useState } from 'react'
import Link from 'next/link'
import {
  Code2,
  Globe,
  Pin,
  PinOff,
  ExternalLink,
  Plus,
  Folder,
  Server,
  GitBranch,
} from 'lucide-react'
import { Button } from '@/components/ui/button'

interface PinnedWorkspace {
  id: string
  name: string
  environment: string
  status: 'active' | 'idle'
  branch: string
  language: string
  framework: string
}

interface PinnedEnvironment {
  id: string
  name: string
  status: 'active' | 'idle'
  services: number
  workspaces: number
  configs: number
  secrets: number
}

interface PinnedResourcesProps {
  workspaces: PinnedWorkspace[]
  environments: PinnedEnvironment[]
  onUnpinWorkspace?: (id: string) => void
  onUnpinEnvironment?: (id: string) => void
}

export function PinnedResources({
  workspaces,
  environments,
  onUnpinWorkspace,
  onUnpinEnvironment,
}: PinnedResourcesProps) {
  const [hoveredWorkspace, setHoveredWorkspace] = useState<string | null>(null)
  const [hoveredEnvironment, setHoveredEnvironment] = useState<string | null>(null)

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      {/* Pinned Workspaces */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="flex items-center gap-2 text-sm font-semibold">
            <Pin className="h-4 w-4" />
            Pinned Workspaces
          </h3>
          <Link href="/workspaces">
            <Button variant="ghost" size="sm" className="gap-1">
              <Plus className="h-3 w-3" />
              Add
            </Button>
          </Link>
        </div>

        {workspaces.length > 0 ? (
          <div className="space-y-2">
            {workspaces.map((workspace) => (
              <div
                key={workspace.id}
                className="bg-card hover:border-accent rounded-lg border p-4 transition-colors"
                onMouseEnter={() => setHoveredWorkspace(workspace.id)}
                onMouseLeave={() => setHoveredWorkspace(null)}
              >
                <div className="flex items-start justify-between">
                  <Link href={`/workspaces/${workspace.id}`} className="min-w-0 flex-1">
                    <div className="flex items-center gap-3">
                      <div className="rounded-lg bg-blue-50 p-2 dark:bg-blue-900/30">
                        <Code2 className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <h4 className="truncate text-sm font-semibold">{workspace.name}</h4>
                          <span
                            className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                              workspace.status === 'active'
                                ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                                : 'bg-secondary text-secondary-foreground'
                            }`}
                          >
                            {workspace.status}
                          </span>
                        </div>
                        <div className="mt-1.5 flex items-center gap-4">
                          <span className="text-muted-foreground flex items-center gap-1 text-xs">
                            <Globe className="h-3 w-3" />
                            {workspace.environment}
                          </span>
                          <span className="text-muted-foreground flex items-center gap-1 text-xs">
                            <GitBranch className="h-3 w-3" />
                            {workspace.branch}
                          </span>
                        </div>
                        <div className="mt-1 flex items-center gap-2">
                          <span className="text-muted-foreground text-xs">
                            {workspace.language} • {workspace.framework}
                          </span>
                        </div>
                      </div>
                    </div>
                  </Link>
                  <div className="flex items-center gap-1">
                    {hoveredWorkspace === workspace.id && onUnpinWorkspace && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onUnpinWorkspace(workspace.id)}
                        className="h-7 w-7 p-0"
                      >
                        <PinOff className="h-3 w-3" />
                      </Button>
                    )}
                    <Link href={`/workspaces/${workspace.id}`}>
                      <Button variant="ghost" size="sm" className="h-7 w-7 p-0">
                        <ExternalLink className="h-3 w-3" />
                      </Button>
                    </Link>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="bg-muted/50 rounded-lg border p-8 text-center">
            <Folder className="text-muted-foreground mx-auto mb-2 h-8 w-8" />
            <p className="text-muted-foreground text-sm">No pinned workspaces</p>
            <Link href="/workspaces">
              <Button variant="outline" size="sm" className="mt-3">
                Browse Workspaces
              </Button>
            </Link>
          </div>
        )}
      </div>

      {/* Pinned Environments */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="flex items-center gap-2 text-sm font-semibold">
            <Pin className="h-4 w-4" />
            Pinned Environments
          </h3>
          <Link href="/environments">
            <Button variant="ghost" size="sm" className="gap-1">
              <Plus className="h-3 w-3" />
              Add
            </Button>
          </Link>
        </div>

        {environments.length > 0 ? (
          <div className="space-y-2">
            {environments.map((env) => (
              <div
                key={env.id}
                className="bg-card hover:border-accent rounded-lg border p-4 transition-colors"
                onMouseEnter={() => setHoveredEnvironment(env.id)}
                onMouseLeave={() => setHoveredEnvironment(null)}
              >
                <div className="flex items-start justify-between">
                  <Link href={`/environments/${env.id}`} className="min-w-0 flex-1">
                    <div className="flex items-center gap-3">
                      <div className="rounded-lg bg-green-50 p-2 dark:bg-green-900/30">
                        <Server className="h-4 w-4 text-green-600 dark:text-green-400" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <h4 className="truncate text-sm font-semibold">{env.name}</h4>
                          <span
                            className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                              env.status === 'active'
                                ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                                : 'bg-secondary text-secondary-foreground'
                            }`}
                          >
                            {env.status}
                          </span>
                        </div>
                        <div className="mt-2 grid grid-cols-2 gap-x-4 gap-y-1">
                          <span className="text-muted-foreground text-xs">
                            {env.services} services
                          </span>
                          <span className="text-muted-foreground text-xs">
                            {env.workspaces} workspaces
                          </span>
                          <span className="text-muted-foreground text-xs">
                            {env.configs} configs
                          </span>
                          <span className="text-muted-foreground text-xs">
                            {env.secrets} secrets
                          </span>
                        </div>
                      </div>
                    </div>
                  </Link>
                  <div className="flex items-center gap-1">
                    {hoveredEnvironment === env.id && onUnpinEnvironment && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onUnpinEnvironment(env.id)}
                        className="h-7 w-7 p-0"
                      >
                        <PinOff className="h-3 w-3" />
                      </Button>
                    )}
                    <Link href={`/environments/${env.id}`}>
                      <Button variant="ghost" size="sm" className="h-7 w-7 p-0">
                        <ExternalLink className="h-3 w-3" />
                      </Button>
                    </Link>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="bg-muted/50 rounded-lg border p-8 text-center">
            <Server className="text-muted-foreground mx-auto mb-2 h-8 w-8" />
            <p className="text-muted-foreground text-sm">No pinned environments</p>
            <Link href="/environments">
              <Button variant="outline" size="sm" className="mt-3">
                Browse Environments
              </Button>
            </Link>
          </div>
        )}
      </div>
    </div>
  )
}
