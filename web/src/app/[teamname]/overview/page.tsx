import { OverviewClient } from './OverviewClient';
import { getTeams } from '@/actions/teams';

interface OverviewPageProps {
  params: Promise<{ teamname: string }>;
}

export default async function OverviewPage({ params }: OverviewPageProps) {
  const { teamname } = await params;
  
  // Get team data for validation
  const teams = await getTeams();
  const team = teams.find(t => 
    t.slug === teamname || 
    t.name.toLowerCase().replace(/\s+/g, '-') === teamname
  );
  
  if (!team) {
    return null; // Layout will handle 404
  }

  return <OverviewClient />;
}