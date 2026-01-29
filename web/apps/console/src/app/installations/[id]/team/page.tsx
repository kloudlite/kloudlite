import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getInstallationById,
  getInstallationMembers,
  getInstallationInvitations,
  getMemberRole,
} from '@/lib/console/supabase-storage-service'
import { TeamMembersTable } from '@/components/team-members-table'
import { TeamInvitationsTable } from '@/components/team-invitations-table'
import { InviteMemberButton } from '@/components/invite-member-button'

interface PageProps {
  params: Promise<{ id: string }>
}

export default async function TeamManagementPage({ params }: PageProps) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  const installation = await getInstallationById(id)

  if (!installation) {
    redirect('/installations')
  }

  // Check user's role
  const userRole = await getMemberRole(id, session.user.id)

  if (!userRole) {
    redirect('/installations')
  }

  const members = await getInstallationMembers(id)
  const invitations = await getInstallationInvitations(id)

  const canManage = userRole === 'owner' || userRole === 'admin'

  return (
    <div className="space-y-6">
      {/* Members Card */}
      <div className="border border-foreground/10 rounded-lg p-6 bg-background">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-foreground">Team Members</h2>
            <p className="text-muted-foreground mt-1 text-sm">Manage who has access to this installation</p>
          </div>
          {canManage && <InviteMemberButton installationId={id} />}
        </div>
        <TeamMembersTable
          members={members}
          currentUserId={session.user.id}
          userRole={userRole}
          installationId={id}
        />
      </div>

      {/* Pending Invitations */}
      {canManage && invitations.length > 0 && (
        <div className="border border-foreground/10 rounded-lg p-6 bg-background">
          <div className="mb-6">
            <h2 className="text-lg font-semibold text-foreground">Pending Invitations</h2>
            <p className="text-muted-foreground mt-1 text-sm">Invitations waiting to be accepted</p>
          </div>
          <TeamInvitationsTable invitations={invitations} installationId={id} />
        </div>
      )}
    </div>
  )
}
