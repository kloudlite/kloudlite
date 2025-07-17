'use client'

import { Button } from '@/components/ui/button'
import { Settings, Users, Layers } from 'lucide-react'
import { useRouter } from 'next/navigation'

interface TeamPreviewProps {
  teamSlug: string
  stats?: {
    environments: { online: number; total: number }
    users: { online: number; total: number }
  }
}

export function TeamPreview({ teamSlug, stats }: TeamPreviewProps) {
  const router = useRouter()
  
  // Mock data if stats not provided
  const teamStats = stats || {
    environments: { online: 2, total: 4 },
    users: { online: 5, total: 12 }
  }

  return (
    <div className="px-6 py-4 space-y-4 border-b border-border">
      {/* Team Stats */}
      <div className="space-y-3">
        {/* Users */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Users className="size-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">Users</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-success">{teamStats.users.online}</span>
            <span className="text-sm text-muted-foreground">/</span>
            <span className="text-sm font-medium">{teamStats.users.total}</span>
          </div>
        </div>
        
        {/* Environments */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Layers className="size-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">Environments</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-success">{teamStats.environments.online}</span>
            <span className="text-sm text-muted-foreground">/</span>
            <span className="text-sm font-medium">{teamStats.environments.total}</span>
          </div>
        </div>
      </div>
      
      {/* Settings Button */}
      <Button 
        variant="outline" 
        size="sm" 
        className="w-full"
        onClick={() => router.push(`/${teamSlug}/settings`)}
      >
        <Settings className="size-4 mr-2" />
        Team Settings
      </Button>
    </div>
  )
}