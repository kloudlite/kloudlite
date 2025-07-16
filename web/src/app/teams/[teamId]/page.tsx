import { notFound } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { ArrowLeft, Settings, UserPlus, Users } from 'lucide-react'
import { getTeam, getTeamMembers } from '@/actions/teams'
import { TeamMembersList } from '@/components/teams/team-members-list'
import { TeamSettings } from '@/components/teams/team-settings'
import { InviteMemberDialog } from '@/components/teams/invite-member-dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface TeamDetailsPageProps {
  params: {
    teamId: string
  }
}

export default async function TeamDetailsPage({ params }: TeamDetailsPageProps) {
  const [team, members] = await Promise.all([
    getTeam(params.teamId),
    getTeamMembers(params.teamId),
  ])

  if (!team) {
    notFound()
  }

  const isOwnerOrAdmin = team.userRole === 'owner' || team.userRole === 'admin'
  const isOwner = team.userRole === 'owner'

  return (
    <div className="container mx-auto px-4 py-8 max-w-7xl">
      {/* Back Button */}
      <Button variant="ghost" size="sm" asChild className="mb-6 rounded-none">
        <Link href="/teams">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Teams
        </Link>
      </Button>

      {/* Team Header */}
      <div className="mb-8">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-3xl font-semibold">{team.name}</h1>
            {team.description && (
              <p className="text-muted-foreground mt-2 max-w-2xl">
                {team.description}
              </p>
            )}
            <div className="flex items-center gap-4 mt-4 text-sm text-muted-foreground">
              <span className="flex items-center">
                <Users className="h-4 w-4 mr-1" />
                {team.memberCount} members
              </span>
              <span className="capitalize px-2 py-1 bg-muted rounded-none">
                {team.visibility}
              </span>
            </div>
          </div>
          {isOwnerOrAdmin && (
            <InviteMemberDialog teamId={team.id}>
              <Button className="rounded-none">
                <UserPlus className="h-4 w-4 mr-2" />
                Invite Member
              </Button>
            </InviteMemberDialog>
          )}
        </div>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="members" className="space-y-6">
        <TabsList className="rounded-none">
          <TabsTrigger value="members" className="rounded-none">
            <Users className="h-4 w-4 mr-2" />
            Members
          </TabsTrigger>
          {isOwnerOrAdmin && (
            <TabsTrigger value="settings" className="rounded-none">
              <Settings className="h-4 w-4 mr-2" />
              Settings
            </TabsTrigger>
          )}
        </TabsList>

        <TabsContent value="members" className="space-y-6">
          <TeamMembersList
            members={members}
            currentUserRole={team.userRole}
            teamId={team.id}
          />
        </TabsContent>

        {isOwnerOrAdmin && (
          <TabsContent value="settings" className="space-y-6">
            <TeamSettings
              team={team}
              isOwner={isOwner}
            />
          </TabsContent>
        )}
      </Tabs>
    </div>
  )
}