import {
  FolderOpen,
  Layers,
  Share2,
  Globe,
  CheckCircle,
  XCircle,
  AlertCircle,
  Play,
  Square,
  Archive,
  Plus,
} from 'lucide-react';
import { StatsCard } from './StatsCard';
import { DashboardStats } from '@/types/dashboard';

interface StatsDashboardProps {
  stats: DashboardStats;
}

export function StatsDashboard({ stats }: StatsDashboardProps) {
  return (
    <div className="space-y-6">
      {/* Workspaces Stats */}
      <div>
        <h3 className="text-lg font-semibold mb-4">Workspaces</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatsCard
            title="Total Workspaces"
            value={stats.workspaces.total}
            icon={FolderOpen}
            description="All workspaces in team"
          />
          <StatsCard
            title="Running"
            value={stats.workspaces.running}
            icon={Play}
            variant="success"
            description="Currently active"
          />
          <StatsCard
            title="Stopped"
            value={stats.workspaces.stopped}
            icon={Square}
            description="Not currently running"
          />
          <StatsCard
            title="Archived"
            value={stats.workspaces.archived}
            icon={Archive}
            variant="warning"
            description="Archived workspaces"
          />
        </div>
      </div>

      {/* Environments Stats */}
      <div>
        <h3 className="text-lg font-semibold mb-4">Environments</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <StatsCard
            title="Total Environments"
            value={stats.environments.total}
            icon={Layers}
            description="All environments"
          />
          <StatsCard
            title="Active"
            value={stats.environments.active}
            icon={CheckCircle}
            variant="success"
            description="Running successfully"
          />
          <StatsCard
            title="Deploying"
            value={stats.environments.deploying}
            icon={AlertCircle}
            variant="warning"
            description="Currently deploying"
          />
          <StatsCard
            title="Failed"
            value={stats.environments.failed}
            icon={XCircle}
            variant="error"
            description="Deployment failed"
          />
          <StatsCard
            title="Stopped"
            value={stats.environments.stopped}
            icon={Square}
            description="Manually stopped"
          />
        </div>
      </div>

      {/* Services Stats */}
      <div>
        <h3 className="text-lg font-semibold mb-4">Services</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <StatsCard
            title="Shared Services"
            value={stats.services.sharedServices}
            icon={Share2}
            description="Team shared services"
          />
          <StatsCard
            title="External Services"
            value={stats.services.externalServices}
            icon={Globe}
            description="External integrations"
          />
          <StatsCard
            title="Healthy"
            value={stats.services.healthy}
            icon={CheckCircle}
            variant="success"
            description="Operating normally"
          />
          <StatsCard
            title="Unhealthy"
            value={stats.services.unhealthy}
            icon={XCircle}
            variant="error"
            description="Experiencing issues"
          />
          <StatsCard
            title="Recently Added"
            value={stats.services.recentlyAdded}
            icon={Plus}
            variant="success"
            description="Added this week"
          />
        </div>
      </div>
    </div>
  );
}