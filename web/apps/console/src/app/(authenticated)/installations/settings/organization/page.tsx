import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getOrgMembers,
  getOrgInvitations,
  getOrgMemberRole,
} from '@/lib/console/storage'
import { getSelectedOrg } from '@/lib/console/get-selected-org'
import { TeamMembersTable } from '@/components/team-members-table'
import { TeamInvitationsTable } from '@/components/team-invitations-table'
import { InviteMemberButton } from '@/components/invite-member-button'
import { DeleteOrganization } from '@/components/delete-organization'

export default async function OrganizationSettingsPage() {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)
  if (!currentOrg) redirect('/installations')
  const userRole = await getOrgMemberRole(currentOrg.id, session.user.id)

  if (!userRole) redirect('/installations')

  const members = await getOrgMembers(currentOrg.id)

  // getOrgInvitations already enriches with inviterName from PII DB
  const allInvitations = await getOrgInvitations(currentOrg.id)
  const pendingInvitations = allInvitations.filter(
    (inv) => inv.status === 'pending'
  )

  return (
    <div className="space-y-6">
      {/* Team Members Card */}
      <div className="border border-foreground/10 rounded-lg bg-background">
        <div className="border-b border-foreground/10 px-6 py-4 flex items-center justify-between">
          <div>
            <h3 className="font-medium text-foreground">Team Members</h3>
            <p className="text-muted-foreground mt-0.5 text-sm">
              Manage who has access to {currentOrg.name}
            </p>
          </div>
          {(userRole === 'owner' || userRole === 'admin') && (
            <InviteMemberButton orgId={currentOrg.id} />
          )}
        </div>
        <div className="p-0">
          <TeamMembersTable
            members={members}
            currentUserId={session.user.id}
            userRole={userRole}
            orgId={currentOrg.id}
          />
        </div>
      </div>

      {/* Pending Invitations Card */}
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
              orgId={currentOrg.id}
            />
          </div>
        </div>
      )}

      {/* Danger Zone — owner only */}
      {userRole === 'owner' && (
        <DeleteOrganization orgId={currentOrg.id} orgName={currentOrg.name} />
      )}
    </div>
  )
}
