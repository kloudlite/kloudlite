'use client'

import { Users, Layers, Settings, FolderOpen } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useRouter } from 'next/navigation'
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from '@/components/ui/hover-card'

interface TeamStatsProps {
  teamSlug: string
  stats?: {
    environments: { online: number; total: number }
    users: { online: number; total: number }
    workspaces: { active: number; total: number }
  }
}

export function TeamStats({ teamSlug, stats }: TeamStatsProps) {
  const router = useRouter()
  
  // Default stats if not provided
  const teamStats = stats || {
    environments: { online: 2, total: 4 },
    users: { online: 5, total: 12 },
    workspaces: { active: 3, total: 8 }
  }

  return (
    <div className="px-4 py-3">
      <div className="flex items-center justify-between text-xs">
        <div className="flex items-center gap-3 flex-1 min-w-0">
          {/* Users */}
          <HoverCard>
            <HoverCardTrigger asChild>
              <div className="flex items-center gap-1.5 px-2 py-1 rounded-md hover:bg-dashboard-hover transition-colors group cursor-pointer">
                <Users className="size-3.5 text-muted-foreground group-hover:text-primary transition-colors" />
                <span className="font-medium text-success">{teamStats.users.online}</span>
                <span className="text-muted-foreground">/</span>
                <span className="font-medium">{teamStats.users.total}</span>
              </div>
            </HoverCardTrigger>
            <HoverCardContent className="w-64 p-3" side="bottom" align="start">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Users className="size-4 text-primary" />
                  <h4 className="font-semibold text-sm">Team Members</h4>
                </div>
                <div className="text-xs space-y-1">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Online now:</span>
                    <span className="font-medium text-success">{teamStats.users.online}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Total members:</span>
                    <span className="font-medium">{teamStats.users.total}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Offline:</span>
                    <span className="font-medium text-muted-foreground">{teamStats.users.total - teamStats.users.online}</span>
                  </div>
                </div>
              </div>
            </HoverCardContent>
          </HoverCard>
          
          {/* Environments */}
          <HoverCard>
            <HoverCardTrigger asChild>
              <div className="flex items-center gap-1.5 px-2 py-1 rounded-md hover:bg-dashboard-hover transition-colors group cursor-pointer">
                <Layers className="size-3.5 text-muted-foreground group-hover:text-primary transition-colors" />
                <span className="font-medium text-success">{teamStats.environments.online}</span>
                <span className="text-muted-foreground">/</span>
                <span className="font-medium">{teamStats.environments.total}</span>
              </div>
            </HoverCardTrigger>
            <HoverCardContent className="w-64 p-3" side="bottom" align="center">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Layers className="size-4 text-primary" />
                  <h4 className="font-semibold text-sm">Environments</h4>
                </div>
                <div className="text-xs space-y-1">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Running:</span>
                    <span className="font-medium text-success">{teamStats.environments.online}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Total environments:</span>
                    <span className="font-medium">{teamStats.environments.total}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Stopped:</span>
                    <span className="font-medium text-muted-foreground">{teamStats.environments.total - teamStats.environments.online}</span>
                  </div>
                </div>
              </div>
            </HoverCardContent>
          </HoverCard>
          
          {/* Workspaces */}
          <HoverCard>
            <HoverCardTrigger asChild>
              <div className="flex items-center gap-1.5 px-2 py-1 rounded-md hover:bg-dashboard-hover transition-colors group cursor-pointer">
                <FolderOpen className="size-3.5 text-muted-foreground group-hover:text-primary transition-colors" />
                <span className="font-medium text-success">{teamStats.workspaces.active}</span>
                <span className="text-muted-foreground">/</span>
                <span className="font-medium">{teamStats.workspaces.total}</span>
              </div>
            </HoverCardTrigger>
            <HoverCardContent className="w-64 p-3" side="bottom" align="end">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <FolderOpen className="size-4 text-primary" />
                  <h4 className="font-semibold text-sm">Workspaces</h4>
                </div>
                <div className="text-xs space-y-1">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Active:</span>
                    <span className="font-medium text-success">{teamStats.workspaces.active}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Total workspaces:</span>
                    <span className="font-medium">{teamStats.workspaces.total}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Inactive:</span>
                    <span className="font-medium text-muted-foreground">{teamStats.workspaces.total - teamStats.workspaces.active}</span>
                  </div>
                </div>
              </div>
            </HoverCardContent>
          </HoverCard>
        </div>
        
        <div className="flex items-center gap-2 ml-3">
          {/* Divider */}
          <div className="h-4 w-px bg-border" />
          
          {/* Settings Button */}
          <Button
            variant="ghost"
            size="icon"
            className="size-6 p-0 hover:bg-dashboard-hover hover:text-primary transition-all duration-200 hover:scale-110"
            onClick={() => router.push(`/${teamSlug}/settings`)}
          >
            <Settings className="size-3.5 transition-transform hover:rotate-45" />
            <span className="sr-only">Team Settings</span>
          </Button>
        </div>
      </div>
    </div>
  )
}