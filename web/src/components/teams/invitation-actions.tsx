'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { acceptInvitation, declineInvitation } from '@/actions/teams'
import { useRouter } from 'next/navigation'
import { toast } from '@/components/ui/use-toast'
import { Loader2, Check, X } from 'lucide-react'

interface InvitationActionsProps {
  invitationId: string
  showLoadingSpinner?: boolean
}

export function InvitationActions({ invitationId, showLoadingSpinner = false }: InvitationActionsProps) {
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
    <div className="flex items-center gap-1 justify-end">
      <Button 
        size="icon-sm-static" 
        variant="ghost-destructive"
        onClick={handleDecline}
        disabled={isAccepting || isDeclining}
        title="Decline invitation"
      >
        {showLoadingSpinner && isDeclining ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <X className="h-4 w-4" />
        )}
      </Button>
      <Button 
        size="icon-sm-static"
        variant="ghost-success"
        onClick={handleAccept}
        disabled={isAccepting || isDeclining}
        title="Accept invitation"
      >
        {showLoadingSpinner && isAccepting ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <Check className="h-4 w-4" />
        )}
      </Button>
    </div>
  )
}