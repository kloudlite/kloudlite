import { ReactNode } from 'react'
import { getTeams } from '@/actions/teams'
import { TeamSettingsLayout } from '@/components/teams/settings/team-settings-layout'

interface TeamSettingsLayoutProps {
  children: ReactNode
  params: Promise<{ teamname: string }>
}

export default async function TeamSettingsRootLayout({ 
  children, 
  params 
}: TeamSettingsLayoutProps) {
  // Await params in Next.js 15
  const { teamname } = await params
  
  // Get team data
  const teams = await getTeams()
  const team = teams.find(t => 
    t.slug === teamname || 
    t.name.toLowerCase().replace(/\s+/g, '-') === teamname
  )
  
  if (!team) {
    return null // Will be handled by parent layout
  }

  return (
    <TeamSettingsLayout
      teamname={teamname}
      teamDisplayName={team.name}
    >
      {children}
    </TeamSettingsLayout>
  )
}