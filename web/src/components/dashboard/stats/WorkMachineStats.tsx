import {
  Server,
  CheckCircle,
  AlertCircle,
  Settings,
  Cpu,
  HardDrive,
  Layers,
  FolderOpen,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { StatsCard } from './StatsCard';
import { WorkMachineStats as WorkMachineStatsType } from '@/types/dashboard';

interface WorkMachineStatsProps {
  stats: WorkMachineStatsType;
}

function ProgressBar({ value, max, label }: { value: number; max: number; label: string }) {
  const percentage = (value / max) * 100;
  
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-sm">
        <span className="text-muted-foreground">{label}</span>
        <span className="font-medium">
          {value} / {max}
        </span>
      </div>
      <div className="w-full bg-muted rounded-full h-2">
        <div
          className="bg-primary h-2 rounded-full transition-all duration-300"
          style={{ width: `${Math.min(percentage, 100)}%` }}
        />
      </div>
      <div className="text-xs text-muted-foreground text-right">
        {percentage.toFixed(1)}% utilized
      </div>
    </div>
  );
}

export function WorkMachineStats({ stats }: WorkMachineStatsProps) {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-bold">Work Machines</h3>
      
      {/* Machine Status Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title="Total Machines"
          value={stats.total}
          icon={Server}
          description="Available work machines"
        />
        <StatsCard
          title="Available"
          value={stats.available}
          icon={CheckCircle}
          variant="success"
          description="Ready for workloads"
        />
        <StatsCard
          title="Busy"
          value={stats.busy}
          icon={AlertCircle}
          variant="warning"
          description="Currently in use"
        />
        <StatsCard
          title="Maintenance"
          value={stats.maintenance}
          icon={Settings}
          variant="error"
          description="Under maintenance"
        />
      </div>

      {/* Resource Utilization */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* CPU and Memory Usage */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Cpu className="h-5 w-5" />
              Resource Utilization
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <ProgressBar
              value={stats.usedCpuCores}
              max={stats.totalCpuCores}
              label="CPU Cores"
            />
            <ProgressBar
              value={stats.usedMemoryGB}
              max={stats.totalMemoryGB}
              label="Memory (GB)"
            />
          </CardContent>
        </Card>

        {/* Running Workloads */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <HardDrive className="h-5 w-5" />
              Running Workloads
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Layers className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Environments</span>
                </div>
                <div className="text-2xl font-bold text-green-500">
                  {stats.runningEnvironments}
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <FolderOpen className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Workspaces</span>
                </div>
                <div className="text-2xl font-bold text-primary">
                  {stats.runningWorkspaces}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <StatsCard
          title="CPU Utilization"
          value={`${stats.cpuUtilization.toFixed(1)}%`}
          icon={Cpu}
          variant={stats.cpuUtilization > 80 ? 'error' : stats.cpuUtilization > 60 ? 'warning' : 'success'}
          description={`${stats.usedCpuCores}/${stats.totalCpuCores} cores used`}
        />
        <StatsCard
          title="Memory Utilization"
          value={`${stats.memoryUtilization.toFixed(1)}%`}
          icon={HardDrive}
          variant={stats.memoryUtilization > 80 ? 'error' : stats.memoryUtilization > 60 ? 'warning' : 'success'}
          description={`${stats.usedMemoryGB}/${stats.totalMemoryGB}GB used`}
        />
      </div>
    </div>
  );
}