import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { cachedInstallationAccess, cachedInstallationById } from '@/lib/console/cached-queries'
import { getOrgMembers, getOrgMemberRole } from '@/lib/console/storage'
import { TeamManagementClient } from '@/components/console/team-client'
import { InviteMemberButton } from '@/components/invite-member-button'

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
  const members = await getOrgMembers(orgId)

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

      <TeamManagementClient
        orgId={orgId}
        members={members.map((m) => ({ id: m.id, userId: m.userId, role: m.role, email: m.userEmail, name: m.userName }))}
        currentUserId={session.user.id}
        userRole={userRole || 'member'}
      />
    </div>
  )
}
