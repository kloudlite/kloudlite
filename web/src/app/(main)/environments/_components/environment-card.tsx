'use client'

import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { MoreHorizontal, Box, FileCode, Lock } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { formatResourceName } from '@/lib/utils'

interface Environment {
  id: string
  name: string
  owner: string
  status: 'active' | 'inactive'
  created: string
  services: number
  configs: number
  secrets: number
  workspaces: string[]
  lastDeployed: string
}

interface EnvironmentCardProps {
  environment: Environment
}

export function EnvironmentCard({ environment: env }: EnvironmentCardProps) {
  return (
    <Link href={`/environments/${env.id}`} className="group block">
      <div className="bg-card hover:border-border cursor-pointer overflow-hidden rounded-lg border transition-all hover:shadow-md">
        {/* Card Header */}
        <div className="border-b px-6 py-4">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <div className="flex items-center gap-3">
                <h3 className="group-hover:text-info text-lg font-medium transition-colors">
                  {formatResourceName(env.name)}
                </h3>
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
              <div className="mt-1 flex flex-col gap-0.5">
                <p className="text-muted-foreground text-sm">
                  Owned by {env.owner.includes('@') ? env.owner.split('@')[0] : env.owner}
                </p>
                <p className="text-muted-foreground text-sm">Last deployed {env.lastDeployed}</p>
              </div>
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
                  Clone Environment
                </DropdownMenuItem>
                <DropdownMenuItem onClick={(e) => e.stopPropagation()}>
                  Export Config
                </DropdownMenuItem>
                <DropdownMenuItem className="text-destructive" onClick={(e) => e.stopPropagation()}>
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* Card Body */}
        <div className="px-6 py-4">
          {/* Stats */}
          <div className="mb-4 grid grid-cols-3 gap-4">
            <div className="flex items-center gap-2">
              <Box className="text-muted-foreground h-4 w-4" />
              <div>
                <p className="text-sm font-medium">{env.services}</p>
                <p className="text-muted-foreground text-xs">Services</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <FileCode className="text-muted-foreground h-4 w-4" />
              <div>
                <p className="text-sm font-medium">{env.configs}</p>
                <p className="text-muted-foreground text-xs">Configs</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Lock className="text-muted-foreground h-4 w-4" />
              <div>
                <p className="text-sm font-medium">{env.secrets}</p>
                <p className="text-muted-foreground text-xs">Secrets</p>
              </div>
            </div>
          </div>

          {/* Connected Workspaces */}
          <div className="border-t pt-4">
            <div className="mb-2 flex items-center justify-between">
              <p className="text-muted-foreground text-xs font-medium">Connected Workspaces</p>
              {env.workspaces.length > 0 && (
                <span className="text-muted-foreground text-xs">{env.workspaces.length}</span>
              )}
            </div>
            {env.workspaces.length > 0 ? (
              <div className="flex flex-wrap gap-1">
                {env.workspaces.map((workspace) => (
                  <span
                    key={workspace}
                    className="bg-muted inline-flex items-center rounded px-2 py-0.5 text-xs"
                  >
                    {workspace}
                  </span>
                ))}
              </div>
            ) : (
              <p className="text-muted-foreground text-xs">No workspaces connected</p>
            )}
          </div>

          {/* Footer */}
          <div className="mt-3 border-t pt-3">
            <span className="text-muted-foreground text-xs">Created {env.created}</span>
          </div>
        </div>
      </div>
    </Link>
  )
}
