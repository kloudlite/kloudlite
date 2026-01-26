'use client'

import { useState } from 'react'
import { Button } from '@kloudlite/ui'
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

  const handleCancel = async (invitationId: string) => {
    if (!confirm('Cancel this invitation?')) return

    setCancelingId(invitationId)
    try {
      const response = await fetch(
        `/api/installations/${installationId}/team/invitations/${invitationId}`,
        { method: 'DELETE' }
      )

      if (!response.ok) throw new Error('Failed to cancel invitation')

      window.location.reload()
    } catch (error) {
      alert('Failed to cancel invitation')
    } finally {
      setCancelingId(null)
    }
  }

  if (invitations.length === 0) {
    return null
  }

  return (
    <div className="overflow-hidden border">
      <table className="min-w-full">
        <thead className="bg-muted/50 border-b">
          <tr>
            <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium uppercase">
              Email
            </th>
            <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium uppercase">
              Role
            </th>
            <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium uppercase">
              Invited By
            </th>
            <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium uppercase">
              Expires
            </th>
            <th className="text-muted-foreground px-6 py-3 text-right text-sm font-medium uppercase">
              Actions
            </th>
          </tr>
        </thead>
        <tbody className="divide-y">
          {invitations.map((invitation) => (
            <tr key={invitation.id} className="hover:bg-muted/50">
              <td className="px-6 py-4 text-base whitespace-nowrap">
                {invitation.email}
              </td>
              <td className="px-6 py-4 text-base whitespace-nowrap">
                <span className="capitalize">{invitation.role}</span>
              </td>
              <td className="px-6 py-4 text-base whitespace-nowrap">
                {invitation.inviterName || 'Unknown'}
              </td>
              <td className="text-muted-foreground px-6 py-4 text-base whitespace-nowrap">
                {new Date(invitation.expiresAt).toLocaleDateString()}
              </td>
              <td className="px-6 py-4 text-right whitespace-nowrap">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleCancel(invitation.id)}
                  disabled={cancelingId === invitation.id}
                  className="text-destructive hover:text-destructive"
                >
                  <X className="h-4 w-4" />
                  {cancelingId === invitation.id ? 'Canceling...' : 'Cancel'}
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
