'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Crown, Shield, Users, MoreVertical, UserMinus } from 'lucide-react'
import type { TeamMember, TeamRole } from '@/lib/teams/types'
import { updateMemberRole, removeMember } from '@/actions/teams'
import { toast } from '@/components/ui/use-toast'
import { useRouter } from 'next/navigation'
import { formatDistanceToNow } from 'date-fns'

interface TeamMembersListProps {
  members: TeamMember[]
  currentUserRole: TeamRole
  teamId: string
}

const roleIcons = {
  owner: Crown,
  admin: Shield,
  member: Users,
}

const roleColors = {
  owner: 'text-amber-600 bg-amber-50 border-amber-200',
  admin: 'text-blue-600 bg-blue-50 border-blue-200',
  member: 'text-gray-600 bg-gray-50 border-gray-200',
}

export function TeamMembersList({ members, currentUserRole, teamId }: TeamMembersListProps) {
  const router = useRouter()
  const [loadingMember, setLoadingMember] = useState<string | null>(null)

  const canManageMembers = currentUserRole === 'owner' || currentUserRole === 'admin'
  const canChangeRoles = currentUserRole === 'owner'

  const handleRoleChange = async (memberId: string, newRole: TeamRole) => {
    setLoadingMember(memberId)
    try {
      const result = await updateMemberRole(teamId, memberId, newRole)
      if (result.success) {
        toast.success('Role updated successfully')
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to update role')
      }
    } catch (error) {
      toast.error('An unexpected error occurred')
    } finally {
      setLoadingMember(null)
    }
  }

  const handleRemoveMember = async (memberId: string) => {
    if (!confirm('Are you sure you want to remove this member?')) return

    setLoadingMember(memberId)
    try {
      const result = await removeMember(teamId, memberId)
      if (result.success) {
        toast.success('Member removed successfully')
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to remove member')
      }
    } catch (error) {
      toast.error('An unexpected error occurred')
    } finally {
      setLoadingMember(null)
    }
  }

  return (
    <div className="space-y-4">
      {members.map((member) => {
        const RoleIcon = roleIcons[member.role]
        const isCurrentUser = false // TODO: Check against current user ID
        
        return (
          <div
            key={member.id}
            className="flex items-center justify-between p-4 border border-border rounded-none bg-card hover:shadow-sm transition-shadow"
          >
            <div className="flex items-center gap-4">
              <Avatar className="h-10 w-10">
                <AvatarFallback className="rounded-none">
                  {member.user.name.slice(0, 2).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              
              <div>
                <div className="flex items-center gap-2">
                  <p className="font-medium">{member.user.name}</p>
                  {isCurrentUser && (
                    <span className="text-xs px-2 py-1 bg-muted rounded-none">You</span>
                  )}
                </div>
                <p className="text-sm text-muted-foreground">{member.user.email}</p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2 text-sm">
                <RoleIcon className="h-4 w-4" />
                <span className={`px-2 py-1 border rounded-none capitalize ${roleColors[member.role]}`}>
                  {member.role}
                </span>
              </div>
              
              <span className="text-sm text-muted-foreground">
                Joined {formatDistanceToNow(member.joinedAt, { addSuffix: true })}
              </span>

              {canManageMembers && !isCurrentUser && member.role !== 'owner' && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="rounded-none h-8 w-8 p-0"
                      disabled={loadingMember === member.id}
                    >
                      <MoreVertical className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="rounded-none">
                    {canChangeRoles && (
                      <>
                        <DropdownMenuItem
                          onClick={() => handleRoleChange(member.id, 'admin')}
                          disabled={member.role === 'admin'}
                        >
                          <Shield className="h-4 w-4 mr-2" />
                          Make Admin
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => handleRoleChange(member.id, 'member')}
                          disabled={member.role === 'member'}
                        >
                          <Users className="h-4 w-4 mr-2" />
                          Make Member
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                      </>
                    )}
                    <DropdownMenuItem
                      onClick={() => handleRemoveMember(member.id)}
                      className="text-destructive"
                    >
                      <UserMinus className="h-4 w-4 mr-2" />
                      Remove from Team
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
            </div>
          </div>
        )
      })}
    </div>
  )
}