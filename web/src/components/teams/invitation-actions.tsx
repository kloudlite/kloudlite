'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { acceptInvitation, declineInvitation } from '@/actions/teams'
import { useRouter } from 'next/navigation'
import { toast } from '@/components/ui/use-toast'
import { Loader2 } from 'lucide-react'

interface InvitationActionsProps {
  invitationId: string
}

export function InvitationActions({ invitationId }: InvitationActionsProps) {
  const [isAccepting, setIsAccepting] = useState(false)
  const [isDeclining, setIsDeclining] = useState(false)
  const router = useRouter()

  const handleAccept = async () => {
    setIsAccepting(true)
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
      setIsAccepting(false)
    }
  }

  const handleDecline = async () => {
    setIsDeclining(true)
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
      setIsDeclining(false)
    }
  }

  return (
    <div className="flex gap-2">
      <Button 
        size="sm" 
        variant="destructive-outline"
        onClick={handleDecline}
        disabled={isAccepting || isDeclining}
        className="flex-1"
      >
        {isDeclining && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
        Decline
      </Button>
      <Button 
        size="sm"
        onClick={handleAccept}
        disabled={isAccepting || isDeclining}
        className="flex-1"
      >
        {isAccepting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
        Accept
      </Button>
    </div>
  )
}