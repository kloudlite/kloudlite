import { Layers, CodeXml } from 'lucide-react';
import { UserDashboardStats } from '@/types/dashboard';
import { StatsCard } from '@/components/dashboard/stats/StatsCard';
import { CountDisplay } from '@/components/ui/count-display';

interface UserDashboardProps {
  stats: UserDashboardStats;
}

export function UserDashboard({ stats }: UserDashboardProps) {
  const totalMyWorkspaces = stats.myWorkspaces.online + stats.myWorkspaces.offline;
  const totalMyEnvironments = stats.myEnvironments.online + stats.myEnvironments.offline;
  const totalTeamEnvironments = stats.totalEnvironments.online + stats.totalEnvironments.offline;

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <StatsCard
        title="My Workspaces"
        value={<CountDisplay active={stats.myWorkspaces.online} total={totalMyWorkspaces} />}
        icon={CodeXml}
        variant="workspaces"
        description={`${stats.myWorkspaces.online} online • ${stats.myWorkspaces.offline} offline`}
      />
      
      <StatsCard
        title="My Environments"
        value={<CountDisplay active={stats.myEnvironments.online} total={totalMyEnvironments} />}
        icon={Layers}
        variant="environments"
        description={`${stats.myEnvironments.online} online • ${stats.myEnvironments.offline} offline`}
      />
      
      <StatsCard
        title="Team Environments"
        value={<CountDisplay active={stats.totalEnvironments.online} total={totalTeamEnvironments} />}
        icon={Layers}
        variant="environments"
        description={`${stats.totalEnvironments.online} online • ${stats.totalEnvironments.offline} offline`}
      />
    </div>
  );
}