import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getInstallationById,
  getInstallationMembers,
  getInstallationInvitations,
  getMemberRole,
} from '@/lib/console/supabase-storage-service'
import { InstallationsHeader } from '@/components/installations-header'
import { GridContainer } from '@/components/grid-container'
import { TeamMembersTable } from '@/components/team-members-table'
import { TeamInvitationsTable } from '@/components/team-invitations-table'
import { InviteMemberButton } from '@/components/invite-member-button'
import { Button } from '@kloudlite/ui'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'

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
    <div className="bg-background min-h-screen">
      <InstallationsHeader user={session.user} />

      <main className="mx-auto max-w-6xl px-6 py-16">
        <GridContainer className="border-t">
          {/* Header */}
          <div className="border-b px-8 py-10">
            <Button asChild variant="ghost" size="sm" className="mb-6">
              <Link href={`/installations/${id}`}>
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Installation
              </Link>
            </Button>
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-3xl font-bold tracking-tight">Team Members</h1>
                <p className="text-muted-foreground mt-2 text-base">
                  Manage team access to {installation.name}
                </p>
              </div>
              {canManage && <InviteMemberButton installationId={id} />}
            </div>
          </div>

          {/* Members Table */}
          <div className="border-b px-8 py-10">
            <h2 className="mb-6 text-xl font-semibold">Members</h2>
            <TeamMembersTable
              members={members}
              currentUserId={session.user.id}
              userRole={userRole}
              installationId={id}
            />
          </div>

          {/* Pending Invitations */}
          {canManage && invitations.length > 0 && (
            <div className="px-8 py-10">
              <h2 className="mb-6 text-xl font-semibold">Pending Invitations</h2>
              <TeamInvitationsTable invitations={invitations} installationId={id} />
            </div>
          )}
        </GridContainer>
      </main>
    </div>
  )
}
