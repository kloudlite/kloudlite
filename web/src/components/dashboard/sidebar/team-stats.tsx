'use client'

import { Users, Layers, Settings, FolderOpen } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useRouter } from 'next/navigation'

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
        <div className="flex items-center gap-4">
          {/* Users */}
          <div className="flex items-center gap-1.5">
            <Users className="size-3.5 text-muted-foreground" />
            <span className="font-medium text-success">{teamStats.users.online}</span>
            <span className="text-muted-foreground">/</span>
            <span className="font-medium">{teamStats.users.total}</span>
          </div>
          
          {/* Environments */}
          <div className="flex items-center gap-1.5">
            <Layers className="size-3.5 text-muted-foreground" />
            <span className="font-medium text-success">{teamStats.environments.online}</span>
            <span className="text-muted-foreground">/</span>
            <span className="font-medium">{teamStats.environments.total}</span>
          </div>
          
          {/* Workspaces */}
          <div className="flex items-center gap-1.5">
            <FolderOpen className="size-3.5 text-muted-foreground" />
            <span className="font-medium text-success">{teamStats.workspaces.active}</span>
            <span className="text-muted-foreground">/</span>
            <span className="font-medium">{teamStats.workspaces.total}</span>
          </div>
        </div>
        
        <div className="flex items-center gap-3 ml-4">
          {/* Divider */}
          <div className="h-4 w-px bg-border" />
          
          {/* Settings Button */}
          <Button
            variant="ghost"
            size="icon"
            className="size-6 p-0"
            onClick={() => router.push(`/${teamSlug}/settings`)}
          >
            <Settings className="size-3.5" />
            <span className="sr-only">Team Settings</span>
          </Button>
        </div>
      </div>
    </div>
  )
}