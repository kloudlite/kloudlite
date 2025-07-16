'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { acceptInvitation, declineInvitation } from '@/actions/teams'
import { useRouter } from 'next/navigation'
import { toast } from '@/components/ui/use-toast'

interface InvitationActionsProps {
  invitationId: string
}

export function InvitationActions({ invitationId }: InvitationActionsProps) {
  const [isLoading, setIsLoading] = useState(false)
  const router = useRouter()

  const handleAccept = async () => {
    setIsLoading(true)
    try {
      const result = await acceptInvitation(invitationId)
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
      const result = await declineInvitation(invitationId)
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
    <div className="flex gap-2">
      <Button 
        size="sm" 
        variant="destructive-outline"
        onClick={handleDecline}
        disabled={isLoading}
        className="flex-1"
      >
        Decline
      </Button>
      <Button 
        size="sm"
        onClick={handleAccept}
        disabled={isLoading}
        className="flex-1"
      >
        Accept
      </Button>
    </div>
  )
}