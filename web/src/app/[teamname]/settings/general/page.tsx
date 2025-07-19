import { getTeams } from '@/actions/teams'
import { TeamGeneralSettings } from '@/components/teams/settings/team-general-settings'

interface GeneralSettingsPageProps {
  params: Promise<{ teamname: string }>
}

export default async function GeneralSettingsPage({ params }: GeneralSettingsPageProps) {
  // Await params in Next.js 15
  const { teamname } = await params
  
  // Get team data
  const teams = await getTeams()
  const team = teams.find(t => 
    t.slug === teamname || 
    t.name.toLowerCase().replace(/\s+/g, '-') === teamname
  )
  
  if (!team) {
    return null // Layout will handle 404
  }

  // Check if user is owner (mock for now)
  const isOwner = team.userRole === 'owner'

  return <TeamGeneralSettings team={team} isOwner={isOwner} />
}