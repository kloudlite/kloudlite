'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import type { TeamInvitation } from '@/lib/teams/types'
import { acceptInvitation, declineInvitation } from '@/actions/teams'
import { useRouter } from 'next/navigation'
import { toast } from '@/components/ui/use-toast'

interface CompactInvitationCardProps {
  invitation: TeamInvitation
}

export function CompactInvitationCard({ invitation }: CompactInvitationCardProps) {
  const [isLoading, setIsLoading] = useState(false)
  const router = useRouter()

  const handleAccept = async () => {
    setIsLoading(true)
    try {
      const result = await acceptInvitation(invitation.id)
      if (result.success) {
        toast.success('Invitation accepted successfully')
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to accept invitation')
      }
    } catch (error) {
      toast.error('An unexpected error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  const handleDecline = async () => {
    setIsLoading(true)
    try {
      const result = await declineInvitation(invitation.id)
      if (result.success) {
        toast.success('Invitation declined')
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to decline invitation')
      }
    } catch (error) {
      toast.error('An unexpected error occurred')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="border border-border rounded-none p-4">
      <h4 className="font-medium mb-1">{invitation.team.name}</h4>
      <p className="text-sm text-muted-foreground mb-3">
        Invited by {invitation.inviter.name} as {invitation.role}
      </p>
      <div className="flex gap-2">
        <Button 
          size="sm" 
          className="rounded-none flex-1 h-8"
          onClick={handleAccept}
          disabled={isLoading}
        >
          Accept
        </Button>
        <Button 
          size="sm" 
          variant="outline" 
          className="rounded-none flex-1 h-8"
          onClick={handleDecline}
          disabled={isLoading}
        >
          Decline
        </Button>
      </div>
    </div>
  )
}