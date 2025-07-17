import { notFound } from 'next/navigation'
import { getTeams } from '@/actions/teams'
import { DashboardLayout } from '@/components/dashboard/layout'

interface TeamLayoutProps {
  children: React.ReactNode
  params: Promise<{ teamname: string }>
}

export default async function TeamLayout({ children, params }: TeamLayoutProps) {
  // Await params before accessing properties
  const { teamname } = await params
  
  // Get all user's teams to validate access
  const teams = await getTeams()
  
  // Find the team by slug
  const team = teams.find(t => 
    t.slug === teamname || 
    t.name.toLowerCase().replace(/\s+/g, '-') === teamname
  )
  
  // If team not found or user doesn't have access, show 404
  if (!team) {
    notFound()
  }

  const teamSlug = team.slug || team.name.toLowerCase().replace(/\s+/g, '-')

  return (
    <DashboardLayout teamSlug={teamSlug} teamName={team.name}>
      {children}
    </DashboardLayout>
  )
}