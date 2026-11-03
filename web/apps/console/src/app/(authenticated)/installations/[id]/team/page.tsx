import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { cachedInstallationAccess, cachedInstallationById } from '@/lib/console/cached-queries'
import { getOrgMembers, getOrgMemberRole, getOrgInvitations } from '@/lib/console/storage'
import { TeamMembersTable } from '@/components/team-members-table'
import { InviteMemberButton } from '@/components/invite-member-button'
import { TeamInvitationsTable } from '@/components/team-invitations-table'

interface PageProps {
  params: Promise<{ id: string }>
}

export default async function TeamManagementPage({ params }: PageProps) {
  const { id } = await params
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  let orgId: string
  try {
    const context = await cachedInstallationAccess(id)
    orgId = context.orgId
  } catch { redirect('/installations') }

  const installation = await cachedInstallationById(id)
  if (!installation) redirect('/installations')

  const userRole = await getOrgMemberRole(orgId, session.user.id)
  if (!userRole) redirect('/installations')
  const members = await getOrgMembers(orgId)

  const allInvitations = await getOrgInvitations(orgId)
  const pendingInvitations = allInvitations.filter(
    (inv) => inv.status === 'pending'
  )

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-2xl font-semibold">Team</h1>
          <p className="text-muted-foreground mt-1 text-sm">
            Organization members who can access this installation
          </p>
        </div>
        {(userRole === 'owner' || userRole === 'admin') && (
          <InviteMemberButton orgId={orgId} />
        )}
      </div>

      <TeamMembersTable
        orgId={orgId}
        members={members}
        currentUserId={session.user.id}
        userRole={userRole}
      />

      {pendingInvitations.length > 0 && (
        <div className="border border-foreground/10 rounded-lg bg-background">
          <div className="border-b border-foreground/10 px-6 py-4">
            <h3 className="font-medium text-foreground">Pending Invitations</h3>
            <p className="text-muted-foreground mt-0.5 text-sm">
              Invitations waiting to be accepted
            </p>
          </div>
          <div className="p-0">
            <TeamInvitationsTable
              invitations={pendingInvitations}
              orgId={orgId}
            />
          </div>
        </div>
      )}
    </div>
  )
}
