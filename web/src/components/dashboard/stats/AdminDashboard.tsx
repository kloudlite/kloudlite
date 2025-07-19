import {
  Users,
  FolderOpen,
  Layers,
  Share2,
  Server,
  Cpu,
  HardDrive,
  Activity,
  AlertTriangle,
  CheckCircle,
  TrendingUp,
  Calendar,
  DollarSign,
  Wifi,
} from 'lucide-react';
import { StatsCard } from './StatsCard';
import { StatsSection } from './StatsSection';
import { ProgressCard } from './ProgressCard';
import { AdminDashboardStats } from '@/types/dashboard';

interface AdminDashboardProps {
  stats: AdminDashboardStats;
  teamName: string;
}

export function AdminDashboard({ stats, teamName }: AdminDashboardProps) {
  return (
    <div className="space-y-8">
      {/* Team Overview */}
      <StatsSection
        title="Team Overview"
        description={`Comprehensive overview of ${teamName} resources and activities`}
        icon={Users}
      >
        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
          <StatsCard
            title="Workspaces"
            value={stats.teamOverview.totalWorkspaces}
            icon={FolderOpen}
            description="Total team workspaces"
          />
          <StatsCard
            title="Environments"
            value={stats.teamOverview.totalEnvironments}
            icon={Layers}
            description="All environments"
          />
          <StatsCard
            title="Services"
            value={stats.teamOverview.totalServices}
            icon={Share2}
            description="Shared & external"
          />
          <StatsCard
            title="Active Users"
            value={stats.teamOverview.activeUsers}
            icon={Activity}
            variant="success"
            description="Recently active"
          />
          <StatsCard
            title="Team Members"
            value={stats.teamOverview.teamMembers}
            icon={Users}
            description="Total members"
          />
        </div>
      </StatsSection>

      {/* Resource Utilization */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Work Machine Resources */}
        <ProgressCard
          title="Compute Resources"
          icon={Cpu}
          items={[
            {
              label: "CPU Cores",
              value: stats.resourceUtilization.workMachines.usedCpuCores,
              max: stats.resourceUtilization.workMachines.totalCpuCores,
            },
            {
              label: "Memory (GB)",
              value: stats.resourceUtilization.workMachines.usedMemoryGB,
              max: stats.resourceUtilization.workMachines.totalMemoryGB,
            },
          ]}
        />

        {/* Storage and Network */}
        <ProgressCard
          title="Storage & Network"
          icon={HardDrive}
          items={[
            {
              label: "Storage (GB)",
              value: stats.resourceUtilization.storageUsed,
              max: stats.resourceUtilization.storageTotal,
            },
            {
              label: "Network (GB/month)",
              value: stats.resourceUtilization.networkTraffic,
              max: 1000, // Example limit
              color: 'default',
            },
          ]}
        />
      </div>

      {/* Work Machine Status */}
      <StatsSection
        title="Work Machine Fleet"
        description="Status and utilization of team work machines"
        icon={Server}
      >
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatsCard
            title="Total Machines"
            value={stats.resourceUtilization.workMachines.total}
            icon={Server}
            description="Available machines"
          />
          <StatsCard
            title="Available"
            value={stats.resourceUtilization.workMachines.available}
            icon={CheckCircle}
            variant="success"
            description="Ready for use"
          />
          <StatsCard
            title="Busy"
            value={stats.resourceUtilization.workMachines.busy}
            icon={Activity}
            variant="warning"
            description="Currently in use"
          />
          <StatsCard
            title="Maintenance"
            value={stats.resourceUtilization.workMachines.maintenance}
            icon={AlertTriangle}
            variant="error"
            description="Under maintenance"
          />
        </div>
      </StatsSection>

      {/* Team Activity & System Health */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Team Activity */}
        <StatsSection
          title="Team Activity"
          description="Deployment and workspace activity metrics"
          icon={TrendingUp}
          variant="compact"
        >
          <div className="grid grid-cols-2 gap-4">
            <StatsCard
              title="Deployments Today"
              value={stats.teamActivity.deploymentsToday}
              icon={Activity}
              variant="success"
              description="Today's deployments"
            />
            <StatsCard
              title="This Week"
              value={stats.teamActivity.deploymentsWeek}
              icon={Calendar}
              description="Weekly deployments"
            />
            <StatsCard
              title="Active Workspaces"
              value={stats.teamActivity.activeWorkspaces}
              icon={FolderOpen}
              variant="success"
              description="Currently running"
            />
            <StatsCard
              title="Active Environments"
              value={stats.teamActivity.activeEnvironments}
              icon={Layers}
              variant="success"
              description="Currently deployed"
            />
          </div>
        </StatsSection>

        {/* System Health */}
        <StatsSection
          title="System Health"
          description="Overall platform health and monitoring"
          icon={CheckCircle}
          variant="compact"
        >
          <div className="grid grid-cols-2 gap-4">
            <StatsCard
              title="Healthy Services"
              value={stats.systemHealth.healthyServices}
              icon={CheckCircle}
              variant="success"
              description="Operating normally"
            />
            <StatsCard
              title="Unhealthy Services"
              value={stats.systemHealth.unhealthyServices}
              icon={AlertTriangle}
              variant="error"
              description="Need attention"
            />
            <StatsCard
              title="Critical Alerts"
              value={stats.systemHealth.criticalAlerts}
              icon={AlertTriangle}
              variant={stats.systemHealth.criticalAlerts > 0 ? 'error' : 'success'}
              description="Require immediate action"
            />
            <StatsCard
              title="System Uptime"
              value={`${stats.systemHealth.systemUptime.toFixed(1)}%`}
              icon={Activity}
              variant={stats.systemHealth.systemUptime > 99 ? 'success' : 'warning'}
              description="Platform availability"
            />
          </div>
        </StatsSection>
      </div>

      {/* Cost Overview */}
      <StatsSection
        title="Cost & Usage"
        description="Monthly cost breakdown and usage metrics"
        icon={DollarSign}
        variant="compact"
      >
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <StatsCard
            title="This Month's Cost"
            value={`$${stats.resourceUtilization.costThisMonth.toFixed(2)}`}
            icon={DollarSign}
            description="Current month spending"
          />
          <StatsCard
            title="Running Workloads"
            value={
              stats.resourceUtilization.workMachines.runningEnvironments +
              stats.resourceUtilization.workMachines.runningWorkspaces
            }
            icon={Activity}
            variant="success"
            description="Active workloads"
          />
          <StatsCard
            title="CPU Utilization"
            value={`${stats.resourceUtilization.workMachines.cpuUtilization.toFixed(1)}%`}
            icon={Cpu}
            variant={
              stats.resourceUtilization.workMachines.cpuUtilization > 80
                ? 'error'
                : stats.resourceUtilization.workMachines.cpuUtilization > 60
                ? 'warning'
                : 'success'
            }
            description="Average CPU usage"
          />
          <StatsCard
            title="Memory Utilization"
            value={`${stats.resourceUtilization.workMachines.memoryUtilization.toFixed(1)}%`}
            icon={HardDrive}
            variant={
              stats.resourceUtilization.workMachines.memoryUtilization > 80
                ? 'error'
                : stats.resourceUtilization.workMachines.memoryUtilization > 60
                ? 'warning'
                : 'success'
            }
            description="Average memory usage"
          />
        </div>
      </StatsSection>
    </div>
  );
}