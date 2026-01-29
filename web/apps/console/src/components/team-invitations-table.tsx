'use client'

import { useState } from 'react'
import { Button } from '@kloudlite/ui'
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
import { X } from 'lucide-react'
import type { InstallationInvitation } from '@/lib/console/supabase-storage-service'

interface TeamInvitationsTableProps {
  invitations: InstallationInvitation[]
  installationId: string
}

export function TeamInvitationsTable({
  invitations,
  installationId,
}: TeamInvitationsTableProps) {
  const [cancelingId, setCancelingId] = useState<string | null>(null)
  const [invitationToCancel, setInvitationToCancel] = useState<{ id: string; email: string } | null>(null)

  const handleCancel = async () => {
    if (!invitationToCancel) return

    setCancelingId(invitationToCancel.id)
    try {
      const response = await fetch(
        `/api/installations/${installationId}/team/invitations/${invitationToCancel.id}`,
        { method: 'DELETE' }
      )

      if (!response.ok) throw new Error('Failed to cancel invitation')

      window.location.reload()
    } catch (error) {
      alert('Failed to cancel invitation')
    } finally {
      setCancelingId(null)
      setInvitationToCancel(null)
    }
  }

  if (invitations.length === 0) {
    return null
  }

  return (
    <div className="overflow-hidden border border-foreground/10 rounded-lg">
      <div className="overflow-x-auto">
        <table className="min-w-full">
          <thead>
            <tr className="border-b border-foreground/10 bg-muted/30">
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide w-[30%]">
                Email
              </th>
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide w-[15%]">
                Role
              </th>
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide w-[25%]">
                Invited By
              </th>
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide w-[15%]">
                Expires
              </th>
              <th className="text-muted-foreground px-6 py-3.5 text-right text-xs font-semibold tracking-wide w-[15%]">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-background divide-y divide-foreground/5">
            {invitations.map((invitation) => (
              <tr key={invitation.id} className="group hover:bg-muted/20 transition-colors">
                <td className="px-6 py-3.5 text-sm text-foreground">
                  {invitation.email}
                </td>
                <td className="px-6 py-3.5 text-sm text-foreground">
                  <span className="capitalize">{invitation.role}</span>
                </td>
                <td className="px-6 py-3.5 text-sm text-foreground">
                  {invitation.inviterName || 'Unknown'}
                </td>
                <td className="text-muted-foreground px-6 py-3.5 text-sm">
                  {new Date(invitation.expiresAt).toLocaleDateString()}
                </td>
                <td className="px-6 py-3.5 text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setInvitationToCancel({ id: invitation.id, email: invitation.email })}
                    disabled={cancelingId === invitation.id}
                    className="text-red-600 dark:text-red-400 hover:bg-red-500/10 hover:text-red-600 dark:hover:text-red-400"
                  >
                    <X className="h-4 w-4 mr-1" />
                    {cancelingId === invitation.id ? 'Canceling...' : 'Cancel'}
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Cancel Invitation Confirmation Dialog */}
      <AlertDialog open={!!invitationToCancel} onOpenChange={(open) => !open && setInvitationToCancel(null)}>
        <AlertDialogContent className="sm:max-w-md">
          <AlertDialogHeader>
            <AlertDialogTitle>Cancel Invitation</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to cancel the invitation for <span className="font-semibold text-foreground">{invitationToCancel?.email}</span>?
              They will not be able to join this installation.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Keep Invitation</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleCancel}
              className="bg-red-600 hover:bg-red-700 active:bg-red-800 focus:ring-red-600 dark:bg-red-600 dark:hover:bg-red-700 dark:active:bg-red-800"
            >
              Cancel Invitation
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
