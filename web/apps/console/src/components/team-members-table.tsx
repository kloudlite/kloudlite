'use client'

import type React from 'react'
import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
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
import { MoreHorizontal, Shield } from 'lucide-react'
import { toast } from 'sonner'
import type { OrgMember, OrgRole } from '@/lib/console/storage'

interface TeamMembersTableProps {
  members: OrgMember[]
  currentUserId: string
  userRole: OrgRole
  orgId: string
}

const roleIcons: Record<OrgRole, React.ComponentType<{ className?: string }>> = {
  owner: Shield,
  admin: Shield,
}

const roleColors: Record<OrgRole, string> = {
  owner: 'bg-purple-500/10 text-purple-700 dark:text-purple-400 border border-purple-500/20',
  admin: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/20',
}

export function TeamMembersTable({
  members,
  currentUserId,
  userRole,
  orgId,
}: TeamMembersTableProps) {
  const router = useRouter()
  const [removingMember, setRemovingMember] = useState<string | null>(null)
  const [memberToRemove, setMemberToRemove] = useState<{ id: string; name: string } | null>(null)
  const [announcement, setAnnouncement] = useState('')

  const canManage = userRole === 'owner' || userRole === 'admin'

  const handleSetMemberToRemove = useCallback((memberId: string, memberName: string) => {
    setMemberToRemove({ id: memberId, name: memberName })
  }, [])

  const handleRemoveMember = async () => {
    if (!memberToRemove) return

    setRemovingMember(memberToRemove.id)
    try {
      const response = await fetch(
        `/api/orgs/${orgId}/members/${memberToRemove.id}`,
        { method: 'DELETE' },
      )

      if (!response.ok) throw new Error('Failed to remove member')
      setAnnouncement(`Removed ${memberToRemove.name} from team.`)
      // Refresh page data to show updated list
      router.refresh()
    } catch {
      toast.error('Failed to remove member')
      setAnnouncement('Failed to remove member.')
    } finally {
      setRemovingMember(null)
      setMemberToRemove(null)
    }
  }

  return (
    <div className="overflow-hidden">
      <p className="sr-only" role="status" aria-live="polite" aria-atomic="true">
        {announcement}
      </p>
      <div className="overflow-x-auto">
        <table className="min-w-full">
          <thead>
            <tr className="border-foreground/10 bg-muted/30 border-b">
              <th className="text-muted-foreground w-[25%] px-6 py-3.5 text-left text-xs font-semibold tracking-wide">
                Member
              </th>
              <th className="text-muted-foreground w-[30%] px-6 py-3.5 text-left text-xs font-semibold tracking-wide">
                Email
              </th>
              <th className="text-muted-foreground w-[20%] px-6 py-3.5 text-left text-xs font-semibold tracking-wide">
                Role
              </th>
              <th className="text-muted-foreground w-[15%] px-6 py-3.5 text-left text-xs font-semibold tracking-wide">
                Added
              </th>
              {canManage && (
                <th className="text-muted-foreground w-[10%] px-6 py-3.5 text-right text-xs font-semibold tracking-wide">
                  Actions
                </th>
              )}
            </tr>
          </thead>
          <tbody className="bg-background divide-foreground/5 divide-y">
            {members.map((member) => {
              const RoleIcon = roleIcons[member.role]
              const isCurrentUser = member.userId === currentUserId
              const isOwner = member.role === 'owner'
              const canRemove = canManage && !isOwner && !isCurrentUser

              return (
                <tr key={member.id} className="group hover:bg-muted/20 transition-colors">
                  <td className="px-6 py-3.5">
                    <div className="text-foreground text-sm font-medium">
                      {member.userName}
                      {isCurrentUser && (
                        <span className="text-muted-foreground ml-2 text-xs font-normal">
                          (You)
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="text-foreground px-6 py-3.5 text-sm">{member.userEmail}</td>
                  <td className="px-6 py-3.5">
                    <span
                      className={`inline-flex items-center gap-1.5 rounded-md px-2.5 py-1 text-[10px] font-semibold tracking-wider uppercase ${roleColors[member.role]}`}
                    >
                      <RoleIcon className="h-3 w-3" />
                      {member.role}
                    </span>
                  </td>
                  <td className="text-muted-foreground px-6 py-3.5 text-sm">
                    {new Date(member.createdAt).toLocaleDateString()}
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
                            <DropdownMenuItem
                              onClick={() =>
                                handleSetMemberToRemove(member.id, member.userName || '')
                              }
                              className="text-red-600 focus:bg-red-500/10 focus:text-red-600 dark:text-red-400 dark:focus:text-red-400"
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
      <AlertDialog
        open={!!memberToRemove}
        onOpenChange={(open) => !open && setMemberToRemove(null)}
      >
        <AlertDialogContent className="sm:max-w-md">
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Team Member</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove{' '}
              <span className="text-foreground font-semibold">{memberToRemove?.name}</span> from
              this organization? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRemoveMember}
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600 active:bg-red-800 dark:bg-red-600 dark:hover:bg-red-700 dark:active:bg-red-800"
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
