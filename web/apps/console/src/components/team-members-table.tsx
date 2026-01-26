'use client'

import { useState } from 'react'
import { Button } from '@kloudlite/ui'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import { MoreHorizontal, Shield, User, Eye } from 'lucide-react'
import type {
  InstallationMember,
  MemberRole,
} from '@/lib/console/supabase-storage-service'

interface TeamMembersTableProps {
  members: InstallationMember[]
  currentUserId: string
  userRole: MemberRole
  installationId: string
}

const roleIcons: Record<MemberRole, any> = {
  owner: Shield,
  admin: Shield,
  member: User,
  viewer: Eye,
}

const roleColors: Record<MemberRole, string> = {
  owner: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
  admin: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  member: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  viewer: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
}

export function TeamMembersTable({
  members,
  currentUserId,
  userRole,
  installationId,
}: TeamMembersTableProps) {
  const [removingMember, setRemovingMember] = useState<string | null>(null)

  const canManage = userRole === 'owner' || userRole === 'admin'

  const handleRemoveMember = async (memberId: string) => {
    if (!confirm('Are you sure you want to remove this member?')) return

    setRemovingMember(memberId)
    try {
      const response = await fetch(
        `/api/installations/${installationId}/team/members/${memberId}`,
        { method: 'DELETE' }
      )

      if (!response.ok) throw new Error('Failed to remove member')

      // Reload page to show updated list
      window.location.reload()
    } catch (error) {
      alert('Failed to remove member')
    } finally {
      setRemovingMember(null)
    }
  }

  const handleChangeRole = async (memberId: string, newRole: MemberRole) => {
    try {
      const response = await fetch(
        `/api/installations/${installationId}/team/members/${memberId}`,
        {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ role: newRole }),
        }
      )

      if (!response.ok) throw new Error('Failed to update role')

      window.location.reload()
    } catch (error) {
      alert('Failed to update member role')
    }
  }

  return (
    <div className="overflow-hidden border">
      <table className="min-w-full">
        <thead className="bg-muted/50 border-b">
          <tr>
            <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium uppercase">
              Member
            </th>
            <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium uppercase">
              Email
            </th>
            <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium uppercase">
              Role
            </th>
            <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium uppercase">
              Added
            </th>
            {canManage && (
              <th className="text-muted-foreground px-6 py-3 text-right text-sm font-medium uppercase">
                Actions
              </th>
            )}
          </tr>
        </thead>
        <tbody className="divide-y">
          {members.map((member) => {
            const RoleIcon = roleIcons[member.role]
            const isCurrentUser = member.userId === currentUserId
            const isOwner = member.role === 'owner'
            const canRemove = canManage && !isOwner && !isCurrentUser

            return (
              <tr key={member.id} className="hover:bg-muted/50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center gap-2">
                    <div className="text-base font-semibold">
                      {member.userName}
                      {isCurrentUser && (
                        <span className="text-muted-foreground ml-2 text-sm">(You)</span>
                      )}
                    </div>
                  </div>
                </td>
                <td className="px-6 py-4 text-base whitespace-nowrap">
                  {member.userEmail}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span
                    className={`inline-flex items-center gap-1.5 px-2 py-0.5 text-sm font-medium ${roleColors[member.role]}`}
                  >
                    <RoleIcon className="h-3.5 w-3.5" />
                    {member.role.charAt(0).toUpperCase() + member.role.slice(1)}
                  </span>
                </td>
                <td className="text-muted-foreground px-6 py-4 text-base whitespace-nowrap">
                  {new Date(member.addedAt).toLocaleDateString()}
                </td>
                {canManage && (
                  <td className="px-6 py-4 text-right whitespace-nowrap">
                    {canRemove ? (
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          {userRole === 'owner' && (
                            <>
                              <DropdownMenuItem
                                onClick={() => handleChangeRole(member.id, 'admin')}
                                disabled={member.role === 'admin'}
                              >
                                Change to Admin
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleChangeRole(member.id, 'member')}
                                disabled={member.role === 'member'}
                              >
                                Change to Member
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleChangeRole(member.id, 'viewer')}
                                disabled={member.role === 'viewer'}
                              >
                                Change to Viewer
                              </DropdownMenuItem>
                            </>
                          )}
                          <DropdownMenuItem
                            onClick={() => handleRemoveMember(member.id)}
                            className="text-destructive"
                            disabled={removingMember === member.id}
                          >
                            {removingMember === member.id ? 'Removing...' : 'Remove'}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    ) : (
                      <span className="text-muted-foreground text-sm">—</span>
                    )}
                  </td>
                )}
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
