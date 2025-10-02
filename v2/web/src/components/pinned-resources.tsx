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
  GitBranch
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
  onUnpinEnvironment
}: PinnedResourcesProps) {
  const [hoveredWorkspace, setHoveredWorkspace] = useState<string | null>(null)
  const [hoveredEnvironment, setHoveredEnvironment] = useState<string | null>(null)

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      {/* Pinned Workspaces */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-2">
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
                className="bg-white rounded-lg border border-gray-200 p-4 hover:border-gray-300 transition-colors"
                onMouseEnter={() => setHoveredWorkspace(workspace.id)}
                onMouseLeave={() => setHoveredWorkspace(null)}
              >
                <div className="flex items-start justify-between">
                  <Link
                    href={`/workspaces/${workspace.id}`}
                    className="flex-1 min-w-0"
                  >
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-blue-50 rounded-lg">
                        <Code2 className="h-4 w-4 text-blue-600" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <h4 className="text-sm font-semibold text-gray-900 truncate">
                            {workspace.name}
                          </h4>
                          <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                            workspace.status === 'active'
                              ? 'bg-green-100 text-green-800'
                              : 'bg-gray-100 text-gray-600'
                          }`}>
                            {workspace.status}
                          </span>
                        </div>
                        <div className="flex items-center gap-4 mt-1.5">
                          <span className="text-xs text-gray-500 flex items-center gap-1">
                            <Globe className="h-3 w-3" />
                            {workspace.environment}
                          </span>
                          <span className="text-xs text-gray-500 flex items-center gap-1">
                            <GitBranch className="h-3 w-3" />
                            {workspace.branch}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 mt-1">
                          <span className="text-xs text-gray-500">
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
          <div className="bg-gray-50 rounded-lg border border-gray-200 p-8 text-center">
            <Folder className="h-8 w-8 text-gray-400 mx-auto mb-2" />
            <p className="text-sm text-gray-600">No pinned workspaces</p>
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
          <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-2">
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
                className="bg-white rounded-lg border border-gray-200 p-4 hover:border-gray-300 transition-colors"
                onMouseEnter={() => setHoveredEnvironment(env.id)}
                onMouseLeave={() => setHoveredEnvironment(null)}
              >
                <div className="flex items-start justify-between">
                  <Link
                    href={`/environments/${env.id}`}
                    className="flex-1 min-w-0"
                  >
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-green-50 rounded-lg">
                        <Server className="h-4 w-4 text-green-600" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <h4 className="text-sm font-semibold text-gray-900 truncate">
                            {env.name}
                          </h4>
                          <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                            env.status === 'active'
                              ? 'bg-green-100 text-green-800'
                              : 'bg-gray-100 text-gray-600'
                          }`}>
                            {env.status}
                          </span>
                        </div>
                        <div className="grid grid-cols-2 gap-x-4 gap-y-1 mt-2">
                          <span className="text-xs text-gray-500">
                            {env.services} services
                          </span>
                          <span className="text-xs text-gray-500">
                            {env.workspaces} workspaces
                          </span>
                          <span className="text-xs text-gray-500">
                            {env.configs} configs
                          </span>
                          <span className="text-xs text-gray-500">
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
          <div className="bg-gray-50 rounded-lg border border-gray-200 p-8 text-center">
            <Server className="h-8 w-8 text-gray-400 mx-auto mb-2" />
            <p className="text-sm text-gray-600">No pinned environments</p>
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