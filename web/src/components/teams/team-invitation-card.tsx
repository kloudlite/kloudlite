'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Mail, Calendar, User, Shield, Crown, Users } from 'lucide-react'
import type { TeamInvitation, TeamRole } from '@/lib/teams/types'
import { formatDistanceToNow } from 'date-fns'
import { acceptInvitation, declineInvitation } from '@/actions/teams'
import { useRouter } from 'next/navigation'
import { toast } from '@/components/ui/use-toast'

interface TeamInvitationCardProps {
  invitation: TeamInvitation
}

const roleIcons = {
  owner: Crown,
  admin: Shield,
  member: Users,
}

export function TeamInvitationCard({ invitation }: TeamInvitationCardProps) {
  const [isLoading, setIsLoading] = useState(false)
  const router = useRouter()
  const RoleIcon = roleIcons[invitation.role]

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

  const isExpired = new Date(invitation.expiresAt) < new Date()

  return (
    <div className={`
      relative border rounded-none transition-all
      ${isExpired ? 'border-border bg-muted/20 opacity-75' : 'border-warning bg-warning/5 border-warning/50'}
    `}>
      {/* Status Badge */}
      {isExpired && (
        <div className="absolute top-4 right-4 px-2 py-1 text-xs font-medium bg-destructive/10 text-destructive rounded-none">
          Expired
        </div>
      )}

      <div className="p-6 space-y-5">
        {/* Team Info */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <div className="p-2 bg-warning/10 text-warning rounded-none">
              <Mail className="h-5 w-5" />
            </div>
            <h3 className="text-xl font-semibold">{invitation.team.name}</h3>
          </div>
          {invitation.team.description && (
            <p className="text-muted-foreground line-clamp-2">
              {invitation.team.description}
            </p>
          )}
        </div>

        {/* Invitation Details */}
        <div className="bg-background p-4 rounded-none border border-border space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center text-sm">
              <User className="h-4 w-4 mr-2 text-muted-foreground" />
              <span className="text-muted-foreground">Invited by</span>
            </div>
            <span className="font-medium">{invitation.inviter.name}</span>
          </div>
          <div className="flex items-center justify-between">
            <div className="flex items-center text-sm">
              <RoleIcon className="h-4 w-4 mr-2 text-muted-foreground" />
              <span className="text-muted-foreground">Your role</span>
            </div>
            <div className={`
              flex items-center gap-1 px-2 py-0.5 text-xs font-medium rounded-none
              ${invitation.role === 'owner' ? 'bg-primary/10 text-primary' : 
                invitation.role === 'admin' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400' : 
                'bg-muted'}
            `}>
              <RoleIcon className="h-3 w-3" />
              <span className="capitalize">{invitation.role}</span>
            </div>
          </div>
          <div className="flex items-center justify-between">
            <div className="flex items-center text-sm">
              <Calendar className="h-4 w-4 mr-2 text-muted-foreground" />
              <span className="text-muted-foreground">Received</span>
            </div>
            <span className="text-sm">{formatDistanceToNow(invitation.createdAt, { addSuffix: true })}</span>
          </div>
        </div>

        {/* Actions */}
        {!isExpired && (
          <div className="flex gap-3 pt-2">
            <Button
              onClick={handleAccept}
              disabled={isLoading}
              className="flex-1 rounded-none h-11 font-medium"
            >
              Accept Invitation
            </Button>
            <Button
              onClick={handleDecline}
              disabled={isLoading}
              variant="outline"
              className="flex-1 rounded-none h-11 font-medium"
            >
              Decline
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}