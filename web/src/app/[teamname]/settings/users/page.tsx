import { getTeams } from '@/actions/teams'
import { TeamUserManagement } from '@/components/teams/settings/team-user-management'

interface UserManagementPageProps {
  params: Promise<{ teamname: string }>
}

export default async function UserManagementPage({ params }: UserManagementPageProps) {
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

  return <TeamUserManagement team={team} isOwner={isOwner} />
}