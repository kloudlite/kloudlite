'use client'

import { useEffect, useState } from 'react'
import { Button } from '@kloudlite/ui'
import { Mail, X } from 'lucide-react'
import type { InstallationInvitation } from '@/lib/console/supabase-storage-service'

export function PendingInvitationsBanner() {
  const [invitations, setInvitations] = useState<InstallationInvitation[]>([])
  const [loading, setLoading] = useState(true)
  const [dismissedIds, setDismissedIds] = useState<Set<string>>(new Set())

  useEffect(() => {
    fetchInvitations()
  }, [])

  const fetchInvitations = async () => {
    try {
      const response = await fetch('/api/invitations/my')
      if (response.ok) {
        const data = await response.json()
        setInvitations(data.invitations || [])
      }
    } finally {
      setLoading(false)
    }
  }

  const handleAccept = async (invitationId: string) => {
    try {
      const response = await fetch(`/api/invitations/${invitationId}/accept`, {
        method: 'POST',
      })

      if (response.ok) {
        window.location.href = '/installations'
      }
    } catch (error) {
      alert('Failed to accept invitation')
    }
  }

  const handleReject = async (invitationId: string) => {
    try {
      const response = await fetch(`/api/invitations/${invitationId}/reject`, {
        method: 'POST',
      })

      if (response.ok) {
        setInvitations((prev) => prev.filter((inv) => inv.id !== invitationId))
      }
    } catch (error) {
      alert('Failed to reject invitation')
    }
  }

  const visibleInvitations = invitations.filter((inv) => !dismissedIds.has(inv.id))

  if (loading || visibleInvitations.length === 0) return null

  return (
    <div className="bg-blue-50 dark:bg-blue-950 border-b border-blue-200 dark:border-blue-800">
      <div className="mx-auto max-w-7xl px-6 py-4">
        <div className="flex items-start gap-4">
          <Mail className="text-blue-600 dark:text-blue-400 mt-0.5 h-5 w-5 flex-shrink-0" />
          <div className="flex-1">
            <h3 className="text-blue-900 dark:text-blue-100 text-base font-semibold">
              You have {visibleInvitations.length} pending team{' '}
              {visibleInvitations.length === 1 ? 'invitation' : 'invitations'}
            </h3>
            <div className="mt-3 space-y-2">
              {visibleInvitations.map((invitation) => (
                <div
                  key={invitation.id}
                  className="bg-white dark:bg-gray-900 flex items-center justify-between gap-4 border border-foreground/10 p-3"
                >
                  <div className="flex-1">
                    <p className="text-base">
                      <strong>{invitation.inviterName}</strong> invited you to join{' '}
                      <strong>{invitation.installationName}</strong> as{' '}
                      <strong>{invitation.role}</strong>
                    </p>
                    <p className="text-muted-foreground mt-1 text-sm">
                      Expires {new Date(invitation.expiresAt).toLocaleDateString()}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button size="sm" onClick={() => handleAccept(invitation.id)}>
                      Accept
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => handleReject(invitation.id)}
                    >
                      Decline
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => setDismissedIds((prev) => new Set(prev).add(invitation.id))}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
