import { getTeams, getTeamInvitations } from '@/actions/teams'
import { TeamsPageContent } from '@/components/teams/teams-page-content'

export default async function TeamsPage() {
  // Fetch user's teams and invitations
  const [teams, invitations] = await Promise.all([
    getTeams(),
    getTeamInvitations()
  ])

  const pendingInvitations = invitations.filter(inv => inv.status === 'pending')

  return <TeamsPageContent teams={teams} pendingInvitations={pendingInvitations} />
}