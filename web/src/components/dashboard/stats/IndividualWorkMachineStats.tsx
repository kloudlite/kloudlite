import {
  Server,
  CheckCircle,
  AlertCircle,
  Settings,
  MapPin,
  Layers,
  FolderOpen,
  Cpu,
  HardDrive,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { IndividualWorkMachine } from '@/types/dashboard';
import { cn } from '@/lib/utils';

interface IndividualWorkMachineStatsProps {
  machines: IndividualWorkMachine[];
}

function getStatusIcon(status: IndividualWorkMachine['status']) {
  switch (status) {
    case 'available':
      return CheckCircle;
    case 'busy':
      return AlertCircle;
    case 'maintenance':
      return Settings;
    default:
      return Server;
  }
}

function getStatusColor(status: IndividualWorkMachine['status']) {
  switch (status) {
    case 'available':
      return 'text-green-500 bg-green-500/10';
    case 'busy':
      return 'text-yellow-500 bg-yellow-500/10';
    case 'maintenance':
      return 'text-destructive bg-destructive/10';
    default:
      return 'text-muted-foreground bg-muted';
  }
}

function getStatusLabel(status: IndividualWorkMachine['status']) {
  switch (status) {
    case 'available':
      return 'Available';
    case 'busy':
      return 'Busy';
    case 'maintenance':
      return 'Maintenance';
    default:
      return 'Unknown';
  }
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

export function IndividualWorkMachineStats({ machines }: IndividualWorkMachineStatsProps) {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-bold">Individual Work Machine Metrics</h3>
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {machines.map((machine) => {
          const StatusIcon = getStatusIcon(machine.status);
          const statusColor = getStatusColor(machine.status);
          
          return (
            <Card key={machine.id}>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    <div className={cn('p-1.5 rounded-md', statusColor)}>
                      <StatusIcon className="h-4 w-4" />
                    </div>
                    {machine.name}
                  </span>
                  <span className="text-sm font-normal text-muted-foreground">
                    {getStatusLabel(machine.status)}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {/* Machine Info */}
                <div className="flex items-center justify-between text-sm">
                  <span className="flex items-center gap-1 text-muted-foreground">
                    <MapPin className="h-3 w-3" />
                    {machine.location}
                  </span>
                  <span className="font-mono text-xs bg-muted px-2 py-1 rounded">
                    {machine.id}
                  </span>
                </div>
                
                {/* Resource Utilization */}
                <div className="space-y-3">
                  <ProgressBar
                    value={machine.usedCpuCores}
                    max={machine.cpuCores}
                    label="CPU Cores"
                  />
                  <ProgressBar
                    value={machine.usedMemoryGB}
                    max={machine.memoryGB}
                    label="Memory (GB)"
                  />
                </div>
                
                {/* Running Workloads */}
                <div className="grid grid-cols-2 gap-4 pt-2 border-t">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <FolderOpen className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">Workspaces</span>
                    </div>
                    <div className="text-lg font-bold text-primary">
                      {machine.runningWorkspaces}
                    </div>
                  </div>
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <Layers className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">Environments</span>
                    </div>
                    <div className="text-lg font-bold text-green-500">
                      {machine.runningEnvironments}
                    </div>
                  </div>
                </div>
                
                {/* Utilization Summary */}
                <div className="grid grid-cols-2 gap-2 pt-2 border-t">
                  <div className="text-center">
                    <div className="text-sm font-medium">
                      {machine.cpuUtilization.toFixed(1)}%
                    </div>
                    <div className="text-xs text-muted-foreground">CPU Usage</div>
                  </div>
                  <div className="text-center">
                    <div className="text-sm font-medium">
                      {machine.memoryUtilization.toFixed(1)}%
                    </div>
                    <div className="text-xs text-muted-foreground">Memory Usage</div>
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}