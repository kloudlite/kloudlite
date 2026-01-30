'use client'

import { useState } from 'react'
import { Button } from '@kloudlite/ui'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@kloudlite/ui'
import { MoreHorizontal, Shield, User, Eye } from 'lucide-react'
import type {
  InstallationMember,
  MemberRole,
} from '@/lib/console/storage'

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
  owner: 'bg-purple-500/10 text-purple-700 dark:text-purple-400 border border-purple-500/20',
  admin: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/20',
  member: 'bg-green-500/10 text-green-700 dark:text-green-400 border border-green-500/20',
  viewer: 'bg-foreground/[0.06] text-foreground border border-foreground/10',
}

export function TeamMembersTable({
  members,
  currentUserId,
  userRole,
  installationId,
}: TeamMembersTableProps) {
  const [removingMember, setRemovingMember] = useState<string | null>(null)
  const [memberToRemove, setMemberToRemove] = useState<{ id: string; name: string } | null>(null)

  const canManage = userRole === 'owner' || userRole === 'admin'

  const handleRemoveMember = async () => {
    if (!memberToRemove) return

    setRemovingMember(memberToRemove.id)
    try {
      const response = await fetch(
        `/api/installations/${installationId}/team/members/${memberToRemove.id}`,
        { method: 'DELETE' }
      )

      if (!response.ok) throw new Error('Failed to remove member')

      // Reload page to show updated list
      window.location.reload()
    } catch (error) {
      alert('Failed to remove member')
    } finally {
      setRemovingMember(null)
      setMemberToRemove(null)
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
    <div className="overflow-hidden border border-foreground/10 rounded-lg">
      <div className="overflow-x-auto">
        <table className="min-w-full">
          <thead>
            <tr className="border-b border-foreground/10 bg-muted/30">
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide w-[25%]">
                Member
              </th>
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide w-[30%]">
                Email
              </th>
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide w-[20%]">
                Role
              </th>
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide w-[15%]">
                Added
              </th>
              {canManage && (
                <th className="text-muted-foreground px-6 py-3.5 text-right text-xs font-semibold tracking-wide w-[10%]">
                  Actions
                </th>
              )}
            </tr>
          </thead>
          <tbody className="bg-background divide-y divide-foreground/5">
            {members.map((member) => {
              const RoleIcon = roleIcons[member.role]
              const isCurrentUser = member.userId === currentUserId
              const isOwner = member.role === 'owner'
              const canRemove = canManage && !isOwner && !isCurrentUser

              return (
                <tr key={member.id} className="group hover:bg-muted/20 transition-colors">
                  <td className="px-6 py-3.5">
                    <div className="text-sm font-medium text-foreground">
                      {member.userName}
                      {isCurrentUser && (
                        <span className="text-muted-foreground ml-2 text-xs font-normal">(You)</span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-3.5 text-sm text-foreground">
                    {member.userEmail}
                  </td>
                  <td className="px-6 py-3.5">
                    <span
                      className={`inline-flex items-center gap-1.5 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-wider rounded-md ${roleColors[member.role]}`}
                    >
                      <RoleIcon className="h-3 w-3" />
                      {member.role}
                    </span>
                  </td>
                  <td className="text-muted-foreground px-6 py-3.5 text-sm">
                    {new Date(member.addedAt).toLocaleDateString()}
                  </td>
                  {canManage && (
                    <td className="px-6 py-3.5 text-right">
                      {canRemove ? (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon">
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
                              onClick={() => setMemberToRemove({ id: member.id, name: member.userName || '' })}
                              className="text-red-600 dark:text-red-400 focus:bg-red-500/10 focus:text-red-600 dark:focus:text-red-400"
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

      {/* Remove Member Confirmation Dialog */}
      <AlertDialog open={!!memberToRemove} onOpenChange={(open) => !open && setMemberToRemove(null)}>
        <AlertDialogContent className="sm:max-w-md">
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Team Member</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove <span className="font-semibold text-foreground">{memberToRemove?.name}</span> from this installation?
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRemoveMember}
              className="bg-red-600 hover:bg-red-700 active:bg-red-800 focus:ring-red-600 dark:bg-red-600 dark:hover:bg-red-700 dark:active:bg-red-800"
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
