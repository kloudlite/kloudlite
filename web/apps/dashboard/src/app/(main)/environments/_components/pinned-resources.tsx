'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Code2, Globe, PinOff, ExternalLink, Plus, Folder, Server } from 'lucide-react'
import { Badge, Button } from '@kloudlite/ui'
import type { PinnedWorkspace, PinnedEnvironment } from '@/types/shared'

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
          <h3 className="text-sm font-semibold">Pinned Workspaces</h3>
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
                className="bg-card hover:border-accent rounded-lg border p-3 transition-colors"
                onMouseEnter={() => setHoveredWorkspace(workspace.id)}
                onMouseLeave={() => setHoveredWorkspace(null)}
              >
                <div className="flex items-center justify-between">
                  <Link href={`/workspaces/${workspace.hash}`} className="min-w-0 flex-1">
                    <div className="flex items-center gap-3">
                      <div className="bg-info/10 dark:bg-info/20 rounded-lg p-2">
                        <Code2 className="text-info h-4 w-4" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <h4 className="truncate text-sm font-medium">{workspace.name}</h4>
                          <Badge variant={workspace.status === 'active' ? 'success' : 'secondary'}>
                            {workspace.status}
                          </Badge>
                        </div>
                        {workspace.environment !== '-' && (
                          <span className="text-muted-foreground flex items-center gap-1 text-xs mt-1">
                            <Globe className="h-3 w-3" />
                            {workspace.environment}
                          </span>
                        )}
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
                    <Link href={`/workspaces/${workspace.hash}`}>
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
          <h3 className="text-sm font-semibold">Pinned Environments</h3>
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
                className="bg-card hover:border-accent rounded-lg border p-3 transition-colors"
                onMouseEnter={() => setHoveredEnvironment(env.id)}
                onMouseLeave={() => setHoveredEnvironment(null)}
              >
                <div className="flex items-center justify-between">
                  <Link href={`/environments/${env.hash}`} className="min-w-0 flex-1">
                    <div className="flex items-center gap-3">
                      <div className="bg-success/10 dark:bg-success/20 rounded-lg p-2">
                        <Server className="text-success h-4 w-4" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <h4 className="truncate text-sm font-medium">{env.name}</h4>
                          <Badge variant={env.status === 'active' ? 'success' : 'secondary'}>
                            {env.status}
                          </Badge>
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
                    <Link href={`/environments/${env.hash}`}>
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
