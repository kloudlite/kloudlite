import {
  FolderOpen,
  Users,
  Server,
  Layers,
  Cpu,
  HardDrive,
} from 'lucide-react';
import { StatsCard } from './stats/StatsCard';
import { CleanWorkMachineTable } from './stats/CleanWorkMachineTable';
import { CleanNodePoolGrid } from './stats/CleanNodePoolGrid';
import { AdminDashboardStats } from '@/types/dashboard';

interface CleanAdminDashboardProps {
  stats: AdminDashboardStats;
}

export function CleanAdminDashboard({ stats }: CleanAdminDashboardProps) {
  return (
    <div className="space-y-8">
      {/* Overview Statistics */}
      <div>
        <h3 className="text-lg font-semibold mb-4">Team Overview</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatsCard
            title="Total Workspaces"
            value={stats.workspaces.total}
            icon={FolderOpen}
            description={`${stats.workspaces.running} running, ${stats.workspaces.stopped} stopped`}
          />
          <StatsCard
            title="Team Users"
            value={stats.users.total}
            icon={Users}
            description={`${stats.users.online} online, ${stats.users.admins} admins`}
          />
          <StatsCard
            title="Work Machines"
            value={stats.workMachines.total}
            icon={Server}
            description={`${stats.workMachines.available} available, ${stats.workMachines.busy} busy`}
          />
          <StatsCard
            title="Node Pools"
            value={stats.nodePools.total}
            icon={Layers}
            description={`${stats.nodePools.healthy} healthy, ${stats.nodePools.degraded} degraded`}
          />
        </div>
      </div>

      {/* Total CPU & RAM Usage - Simplified */}
      <div>
        <h3 className="text-lg font-semibold mb-4">Resource Overview</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <StatsCard
            title="Total CPU Utilization"
            value={`${stats.totalResources.cpuUtilization.toFixed(1)}%`}
            icon={Cpu}
            variant={stats.totalResources.cpuUtilization > 80 ? 'error' : stats.totalResources.cpuUtilization > 60 ? 'warning' : 'success'}
            description={`${stats.totalResources.usedCpuCores}/${stats.totalResources.totalCpuCores} cores used`}
          />
          <StatsCard
            title="Total Memory Utilization"
            value={`${stats.totalResources.memoryUtilization.toFixed(1)}%`}
            icon={HardDrive}
            variant={stats.totalResources.memoryUtilization > 80 ? 'error' : stats.totalResources.memoryUtilization > 60 ? 'warning' : 'success'}
            description={`${stats.totalResources.usedMemoryGB}/${stats.totalResources.totalMemoryGB}GB used`}
          />
        </div>
      </div>

      {/* Clean Work Machine Overview */}
      <CleanWorkMachineTable machines={stats.individualWorkMachines} />

      {/* Clean Node Pool Overview */}
      <CleanNodePoolGrid pools={stats.individualNodePools} />
    </div>
  );
}