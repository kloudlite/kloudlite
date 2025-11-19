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
import { formatWorkspaceName } from '@/lib/utils'

interface Workspace {
  id: string
  name: string
  ownedBy: string
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
      <div className="border-border bg-card hover:border-border h-full cursor-pointer overflow-hidden rounded-lg border transition-all hover:shadow-md">
        {/* Card Header */}
        <div className="border-border border-b px-6 py-4">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <h3 className="text-foreground group-hover:text-info text-lg font-medium transition-colors">
                {formatWorkspaceName(workspace.ownedBy, workspace.name)}
              </h3>
              <p className="text-muted-foreground mt-1 line-clamp-1 text-sm">
                {workspace.description}
              </p>
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
                <DropdownMenuItem className="text-destructive" onClick={(e) => e.stopPropagation()}>
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
            <span className="bg-info/10 text-info inline-flex items-center rounded px-2 py-0.5 text-xs font-medium">
              {workspace.language}
            </span>
            <span className="bg-muted text-foreground inline-flex items-center rounded px-2 py-0.5 text-xs font-medium">
              {workspace.framework}
            </span>
          </div>

          {/* Stats */}
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm">
              <GitBranch className="text-muted-foreground h-4 w-4" />
              <span className="text-muted-foreground">Branch:</span>
              <span className="text-foreground font-medium">{workspace.branch}</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Users className="text-muted-foreground h-4 w-4" />
              <span className="text-muted-foreground">Team:</span>
              <span className="text-foreground font-medium">{workspace.team} members</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Clock className="text-muted-foreground h-4 w-4" />
              <span className="text-muted-foreground">Last activity:</span>
              <span className="text-foreground font-medium">{workspace.lastActivity}</span>
            </div>
          </div>

          {/* Environment */}
          <div className="border-border mt-3 border-t pt-3">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-muted-foreground text-xs">Connected to</p>
                <p className="text-foreground mt-0.5 text-sm font-medium">
                  {workspace.environment}
                </p>
              </div>
              <span
                className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                  workspace.status === 'active'
                    ? 'bg-success/10 text-success'
                    : 'bg-muted text-muted-foreground'
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
