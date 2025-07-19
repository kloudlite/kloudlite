import {
  FolderOpen,
  Layers,
  Share2,
  Globe,
  CheckCircle,
  Play,
  Square,
  Activity,
  Calendar,
  User,
} from 'lucide-react';
import { StatsCard } from './StatsCard';
import { StatsSection } from './StatsSection';
import { UserDashboardStats } from '@/types/dashboard';

interface UserDashboardProps {
  stats: UserDashboardStats;
  userName: string;
}

export function UserDashboard({ stats, userName }: UserDashboardProps) {
  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <div className="border-b pb-4">
        <h2 className="text-xl font-semibold">Welcome back, {userName}</h2>
        <p className="text-muted-foreground mt-1">
          Here's an overview of your workspaces, environments, and recent activity.
        </p>
      </div>

      {/* Personal Workspaces */}
      <StatsSection
        title="My Workspaces"
        description="Workspaces you've created or have been assigned to"
        icon={FolderOpen}
        variant="compact"
      >
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatsCard
            title="Created by Me"
            value={stats.personalWorkspaces.created}
            icon={User}
            description="Your workspaces"
          />
          <StatsCard
            title="Assigned to Me"
            value={stats.personalWorkspaces.assigned}
            icon={Share2}
            description="Shared with you"
          />
          <StatsCard
            title="Running"
            value={stats.personalWorkspaces.running}
            icon={Play}
            variant="success"
            description="Currently active"
          />
          <StatsCard
            title="Stopped"
            value={stats.personalWorkspaces.stopped}
            icon={Square}
            description="Not running"
          />
        </div>
      </StatsSection>

      {/* Environments & Services */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Accessible Environments */}
        <StatsSection
          title="My Environments"
          description="Environments you have access to"
          icon={Layers}
          variant="compact"
        >
          <div className="grid grid-cols-2 gap-4">
            <StatsCard
              title="Accessible"
              value={stats.accessibleEnvironments.accessible}
              icon={Layers}
              description="Total access"
            />
            <StatsCard
              title="Active"
              value={stats.accessibleEnvironments.active}
              icon={CheckCircle}
              variant="success"
              description="Running now"
            />
            <StatsCard
              title="Deploying"
              value={stats.accessibleEnvironments.deploying}
              icon={Activity}
              variant="warning"
              description="In progress"
            />
            <StatsCard
              title="Failed"
              value={stats.accessibleEnvironments.failed}
              icon={Square}
              variant="error"
              description="Need attention"
            />
          </div>
        </StatsSection>

        {/* Personal Services */}
        <StatsSection
          title="My Services"
          description="Services relevant to your work"
          icon={Share2}
          variant="compact"
        >
          <div className="grid grid-cols-2 gap-4">
            <StatsCard
              title="Personal"
              value={stats.personalServices.personalServices}
              icon={User}
              description="Your services"
            />
            <StatsCard
              title="Shared Access"
              value={stats.personalServices.sharedAccessible}
              icon={Share2}
              description="Team services"
            />
            <StatsCard
              title="External"
              value={stats.personalServices.externalAccessible}
              icon={Globe}
              description="External APIs"
            />
            <StatsCard
              title="Healthy"
              value={stats.personalServices.healthyServices}
              icon={CheckCircle}
              variant="success"
              description="Working well"
            />
          </div>
        </StatsSection>
      </div>

      {/* Recent Activity Summary */}
      <StatsSection
        title="Activity Summary"
        description="Your recent development activity"
        icon={Activity}
        variant="compact"
      >
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <StatsCard
            title="Today's Actions"
            value={stats.recentActivity.todayActions}
            icon={Activity}
            variant="success"
            description="Actions performed today"
          />
          <StatsCard
            title="This Week"
            value={stats.recentActivity.weekActions}
            icon={Calendar}
            description="Total weekly activity"
          />
          <StatsCard
            title="Last Activity"
            value={
              stats.recentActivity.lastActivity
                ? new Date(stats.recentActivity.lastActivity).toLocaleDateString()
                : 'No recent activity'
            }
            icon={Calendar}
            description="Most recent action"
          />
        </div>
      </StatsSection>
    </div>
  );
}