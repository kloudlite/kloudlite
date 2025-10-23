'use client'

import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { MoreHorizontal, GitBranch, Users, Clock } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface Workspace {
  id: string
  name: string
  description: string
  status: 'active' | 'idle'
  lastActivity: string
  branch: string
  team: number
  environment: string
  language: string
  framework: string
}

interface WorkspaceCardProps {
  workspace: Workspace
}

export function WorkspaceCard({ workspace }: WorkspaceCardProps) {
  return (
    <Link href={`/workspaces/${workspace.id}`} className="group block">
      <div className="h-full cursor-pointer overflow-hidden rounded-lg border border-gray-200 bg-white transition-all hover:border-gray-300 hover:shadow-md">
        {/* Card Header */}
        <div className="border-b border-gray-100 px-6 py-4">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <h3 className="text-lg font-medium text-gray-900 transition-colors group-hover:text-blue-600">
                {workspace.name}
              </h3>
              <p className="mt-1 line-clamp-1 text-sm text-gray-500">{workspace.description}</p>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-8 w-8 p-0"
                  onClick={(e) => e.preventDefault()}
                >
                  <MoreHorizontal className="h-4 w-4" />
                  <span className="sr-only">Open menu</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={(e) => e.stopPropagation()}>
                  Open in Editor
                </DropdownMenuItem>
                <DropdownMenuItem onClick={(e) => e.stopPropagation()}>View Logs</DropdownMenuItem>
                <DropdownMenuItem className="text-red-600" onClick={(e) => e.stopPropagation()}>
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* Card Body */}
        <div className="px-6 py-4">
          {/* Tech Stack */}
          <div className="mb-3 flex items-center gap-2">
            <span className="inline-flex items-center rounded bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-800">
              {workspace.language}
            </span>
            <span className="inline-flex items-center rounded bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700">
              {workspace.framework}
            </span>
          </div>

          {/* Stats */}
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm">
              <GitBranch className="h-4 w-4 text-gray-400" />
              <span className="text-gray-600">Branch:</span>
              <span className="font-medium text-gray-900">{workspace.branch}</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Users className="h-4 w-4 text-gray-400" />
              <span className="text-gray-600">Team:</span>
              <span className="font-medium text-gray-900">{workspace.team} members</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Clock className="h-4 w-4 text-gray-400" />
              <span className="text-gray-600">Last activity:</span>
              <span className="font-medium text-gray-900">{workspace.lastActivity}</span>
            </div>
          </div>

          {/* Environment */}
          <div className="mt-3 border-t border-gray-100 pt-3">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-gray-500">Connected to</p>
                <p className="mt-0.5 text-sm font-medium text-gray-900">{workspace.environment}</p>
              </div>
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
          </div>
        </div>
      </div>
    </Link>
  )
}
