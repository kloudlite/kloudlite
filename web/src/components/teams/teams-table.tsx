'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { Crown, Shield, UserCheck, Settings, MoreVertical, Plus, Users, ArrowUpDown } from 'lucide-react'
import type { Team, TeamRole } from '@/lib/teams/types'
import { formatDistanceToNow } from 'date-fns'

const roleIcons = {
  owner: Crown,
  admin: Shield,
  member: UserCheck,
}

interface TeamsTableProps {
  teams: (Team & { userRole: TeamRole })[]
  initialDisplayCount?: number
}

export function TeamsTable({ teams, initialDisplayCount = 5 }: TeamsTableProps) {
  const [displayCount, setDisplayCount] = useState(initialDisplayCount)
  const hasMore = displayCount < teams.length
  const displayedTeams = teams.slice(0, displayCount)

  const handleLoadMore = () => {
    setDisplayCount(prev => Math.min(prev + 5, teams.length))
  }

  return (
    <>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b text-sm text-muted-foreground">
              <th className="text-left font-medium px-6 py-3">Team Name</th>
              <th className="text-left font-medium px-6 py-3">Role</th>
              <th className="text-left font-medium px-6 py-3">Members</th>
              <th className="text-left font-medium px-6 py-3">Last Accessed</th>
              <th className="text-left font-medium px-6 py-3">Member Since</th>
              <th className="px-6 py-3"></th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {displayedTeams.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-12 text-center">
                  <div className="max-w-sm mx-auto">
                    <Users className="h-10 w-10 mx-auto mb-3 text-muted-foreground/50" />
                    <h3 className="font-medium mb-1">No teams found</h3>
                    <p className="text-sm text-muted-foreground mb-4">
                      {teams.length === 0 
                        ? "Get started by creating your first team"
                        : "Try adjusting your search to find teams"
                      }
                    </p>
                    {teams.length === 0 && (
                      <Button size="sm" asChild>
                        <Link href="/teams/new">
                          <Plus className="h-4 w-4 mr-2" />
                          Create Team
                        </Link>
                      </Button>
                    )}
                  </div>
                </td>
              </tr>
            ) : (
              displayedTeams.map((team) => {
                const RoleIcon = roleIcons[team.userRole]
                return (
                  <tr key={team.id} className="hover:bg-muted/50 transition-colors group">
                    <td className="px-6 py-4">
                      <div>
                        <Link href={`/${team.slug || team.name.toLowerCase().replace(/\s+/g, '-')}`} className="font-medium hover:text-primary transition-colors">
                          {team.name}
                        </Link>
                        <p className="text-sm text-muted-foreground mt-0.5 max-w-md truncate">
                          {team.description}
                        </p>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-1.5">
                        <RoleIcon className="h-3.5 w-3.5 text-muted-foreground" />
                        <span className="text-sm capitalize">{team.userRole}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-sm">{team.memberCount}</span>
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-sm text-muted-foreground">
                        {team.lastActivity ? formatDistanceToNow(team.lastActivity, { addSuffix: true }) : 'Never'}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-sm text-muted-foreground">
                        {team.userRole === 'owner' 
                          ? formatDistanceToNow(team.createdAt, { addSuffix: true })
                          : formatDistanceToNow(team.joinedAt || team.createdAt, { addSuffix: true })
                        }
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <Button variant="ghost" size="icon" asChild>
                          <Link href={`/${team.slug || team.name.toLowerCase().replace(/\s+/g, '-')}/settings`}>
                            <Settings className="h-4 w-4" />
                          </Link>
                        </Button>
                        <Button variant="ghost" size="icon">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>
      
      {hasMore && (
        <div className="border-t px-6 py-4 text-center">
          <Button variant="ghost" onClick={handleLoadMore}>
            Load More ({teams.length - displayCount} remaining)
          </Button>
        </div>
      )}
    </>
  )
}