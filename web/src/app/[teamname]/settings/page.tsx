import { getTeams } from '@/actions/teams'
import { TeamSettings } from '@/components/teams/team-settings'

interface TeamSettingsPageProps {
  params: { teamname: string }
}

export default async function TeamSettingsPage({ params }: TeamSettingsPageProps) {
  // Get team data
  const teams = await getTeams()
  const team = teams.find(t => 
    t.slug === params.teamname || 
    t.name.toLowerCase().replace(/\s+/g, '-') === params.teamname
  )
  
  if (!team) {
    return null // Layout will handle 404
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-semibold">Team Settings</h1>
        <p className="text-muted-foreground mt-1">
          Manage your team settings, members, and preferences
        </p>
      </div>
      
      <TeamSettings team={team} />
    </div>
  )
}