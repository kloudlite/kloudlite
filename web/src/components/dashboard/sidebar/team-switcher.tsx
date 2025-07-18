'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ChevronDown, Plus, Building2, Check, Globe2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useRouter } from 'next/navigation'
import { getTeams } from '@/actions/teams'
import { Team } from '@/lib/teams/types'
import { TeamAvatar } from './team-avatar'

interface TeamSwitcherProps {
  teamSlug: string
  teamName: string
  className?: string
  stats?: {
    environments: { online: number; total: number }
    users: { online: number; total: number }
  }
}

export function TeamSwitcher({ teamSlug, teamName, className, stats }: TeamSwitcherProps) {
  const router = useRouter()
  const [teams, setTeams] = useState<Team[]>([])
  const [loadingTeams, setLoadingTeams] = useState(true)
  const [machineRunning, setMachineRunning] = useState(true)
  
  // Default stats if not provided
  const teamStats = stats || {
    environments: { online: 2, total: 4 },
    users: { online: 5, total: 12 }
  }

  useEffect(() => {
    const fetchTeams = async () => {
      try {
        const userTeams = await getTeams()
        setTeams(userTeams)
      } catch (error) {
        console.error('Failed to fetch teams:', error)
      } finally {
        setLoadingTeams(false)
      }
    }
    fetchTeams()
  }, [])

  const currentTeam = teams.find(t => 
    t.slug === teamSlug || 
    t.name.toLowerCase().replace(/\s+/g, '-') === teamSlug
  )

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button 
          variant="ghost" 
          className={cn(
            "w-full h-auto py-4 px-4 justify-between bg-card hover:bg-dashboard-hover rounded-lg border shadow-dashboard-card-shadow focus:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2",
            loadingTeams && "cursor-not-allowed opacity-50",
            className
          )}
          disabled={loadingTeams}
          disableActiveTransition={true}
        >
          <div className="flex items-center gap-3">
            <TeamAvatar 
              name={teamName} 
              showStatus={true} 
              status={machineRunning ? 'active' : 'inactive'} 
              size="md"
            />
            <div className="flex-1 text-left">
              <p className="text-sm font-semibold">{teamName}</p>
              <p className="text-xs text-muted-foreground/80 flex items-center gap-1">
                <Globe2 className="size-3" />
                {currentTeam?.region || 'us-west-2'}
              </p>
            </div>
          </div>
          <ChevronDown className="size-4 text-muted-foreground" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-[var(--radix-dropdown-menu-trigger-width)]">
        <DropdownMenuLabel>Switch Team</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {teams.map((team) => {
          const slug = team.slug || team.name.toLowerCase().replace(/\s+/g, '-')
          const isCurrentTeam = slug === teamSlug
          return (
            <DropdownMenuItem
              key={team.id}
              onClick={() => router.push(`/${slug}`)}
              className="cursor-pointer"
            >
              <div className="flex items-center gap-3 w-full">
                <TeamAvatar 
                  name={team.name} 
                  size="sm"
                />
                <div className="flex-1">
                  <p className="text-sm font-medium text-foreground">{team.name}</p>
                  <p className="text-xs text-muted-foreground">{team.memberCount} members</p>
                </div>
                {isCurrentTeam && (
                  <Check className="size-4 text-muted-foreground" />
                )}
              </div>
            </DropdownMenuItem>
          )
        })}
        <DropdownMenuSeparator />
        <DropdownMenuItem className="cursor-pointer" onClick={() => router.push('/teams/new')}>
          <div className="flex items-center gap-3 w-full">
            <div className="size-8 rounded-md bg-muted flex items-center justify-center">
              <Plus className="size-4" />
            </div>
            <span className="text-sm font-medium">Create New Team</span>
          </div>
        </DropdownMenuItem>
        <DropdownMenuItem className="cursor-pointer" onClick={() => router.push('/teams')}>
          <div className="flex items-center gap-3 w-full">
            <div className="size-8 rounded-md bg-muted flex items-center justify-center">
              <Building2 className="size-4" />
            </div>
            <span className="text-sm font-medium">All Teams</span>
          </div>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}