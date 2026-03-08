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
import { Building2 } from 'lucide-react'

export default async function OrganizationSettingsPage() {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const currentOrg = await getSelectedOrg(session.user.id, session.user.name, session.user.email)
  if (!currentOrg) redirect('/installations')
  const userRole = await getOrgMemberRole(currentOrg.id, session.user.id)

  if (!userRole) redirect('/installations')

  // getOrgMembers already enriches with userName/userEmail from PII DB
  const members = await getOrgMembers(currentOrg.id)

  // getOrgInvitations already enriches with inviterName from PII DB
  const allInvitations = await getOrgInvitations(currentOrg.id)
  const pendingInvitations = allInvitations.filter(
    (inv) => inv.status === 'pending'
  )

  return (
    <div className="space-y-8">
      {/* Org Info Section */}
      <div>
        <div className="flex items-center gap-3 mb-5">
          <Building2 className="h-5 w-5 text-muted-foreground" />
          <div>
            <h2 className="text-xl font-semibold">{currentOrg.name}</h2>
            <p className="text-muted-foreground text-sm">
              Slug: <span className="font-mono">{currentOrg.slug}</span>
            </p>
          </div>
        </div>
      </div>

      {/* Team Members Section */}
      <div>
        <div className="flex items-center justify-between mb-5">
          <div>
            <h2 className="text-xl font-semibold">Team Members</h2>
            <p className="text-muted-foreground mt-1 text-base">
              Manage who has access to this organization
            </p>
          </div>
          {(userRole === 'owner' || userRole === 'admin') && (
            <InviteMemberButton orgId={currentOrg.id} />
          )}
        </div>

        <TeamMembersTable
          members={members}
          currentUserId={session.user.id}
          userRole={userRole}
          orgId={currentOrg.id}
        />
      </div>

      {/* Pending Invitations Section */}
      {pendingInvitations.length > 0 && (
        <div>
          <div className="mb-5">
            <h2 className="text-xl font-semibold">Pending Invitations</h2>
            <p className="text-muted-foreground mt-1 text-base">
              Invitations waiting to be accepted
            </p>
          </div>

          <TeamInvitationsTable
            invitations={pendingInvitations}
            orgId={currentOrg.id}
          />
        </div>
      )}

      {/* Danger Zone — owner only */}
      {userRole === 'owner' && (
        <div>
          <div className="mb-5">
            <h2 className="text-xl font-semibold">Danger Zone</h2>
          </div>
          <DeleteOrganization orgId={currentOrg.id} orgName={currentOrg.name} />
        </div>
      )}
    </div>
  )
}
