'use client';

import {
  CheckCircle,
  AlertCircle,
  Settings,
  MapPin,
  Users,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { MetricsTooltip } from './MetricsTooltip';
import { IndividualWorkMachine } from '@/types/dashboard';
import { cn } from '@/lib/utils';

interface CleanWorkMachineTableProps {
  machines: IndividualWorkMachine[];
}

function getMachineSize(cpuCores: number, memoryGB: number): string {
  if (cpuCores <= 4 && memoryGB <= 16) return 'Small';
  if (cpuCores <= 8 && memoryGB <= 32) return 'Medium';
  if (cpuCores <= 16 && memoryGB <= 64) return 'Large';
  if (cpuCores <= 32 && memoryGB <= 128) return 'XLarge';
  return 'XXLarge';
}

function getStatusIcon(status: IndividualWorkMachine['status']) {
  switch (status) {
    case 'on':
      return CheckCircle;
    case 'off':
      return AlertCircle;
    default:
      return CheckCircle;
  }
}

function getStatusColor(status: IndividualWorkMachine['status']) {
  switch (status) {
    case 'on':
      return 'bg-green-500/10 text-green-500 border-green-500/20 hover:bg-green-500/20';
    case 'off':
      return 'bg-muted text-muted-foreground border-border hover:bg-muted/80';
    default:
      return 'bg-muted text-muted-foreground border-border hover:bg-muted/80';
  }
}

function getStatusLabel(status: IndividualWorkMachine['status']) {
  switch (status) {
    case 'on':
      return 'Online';
    case 'off':
      return 'Offline';
    default:
      return 'Unknown';
  }
}

export function CleanWorkMachineTable({ machines }: CleanWorkMachineTableProps) {
  return (
    <div className="space-y-6">
      <div className="border-b border-border pb-4">
        <h2 className="text-2xl font-bold tracking-tight">Work Machines</h2>
        <p className="text-base text-muted-foreground mt-1">
          Infrastructure machines hosting developer workspaces and environments
        </p>
      </div>
      
      <TooltipProvider>
        <div className="overflow-hidden rounded-lg border">
          <table className="w-full">
            <thead className="bg-muted/50">
              <tr className="border-b">
                <th className="text-left p-4 font-medium text-sm">Name</th>
                <th className="text-left p-4 font-medium text-sm">Size</th>
                <th className="text-left p-4 font-medium text-sm">Workloads</th>
                <th className="text-left p-4 font-medium text-sm">Location</th>
                <th className="text-left p-4 font-medium text-sm">Status</th>
              </tr>
            </thead>
            <tbody>
              {machines.map((machine, index) => {
                const StatusIcon = getStatusIcon(machine.status);
                const statusColor = getStatusColor(machine.status);
                const machineSize = getMachineSize(machine.cpuCores, machine.memoryGB);
                
                return (
                  <Tooltip key={machine.id}>
                    <TooltipTrigger asChild>
                      <tr className={cn(
                        "border-b hover:bg-muted/40 cursor-pointer transition-colors",
                        index % 2 === 0 ? "bg-background" : "bg-muted/20"
                      )}>
                        <td className="p-4">
                          <div className="font-medium text-sm">{machine.name}</div>
                          <div className="text-xs text-muted-foreground">{machine.id}</div>
                        </td>
                        <td className="p-4">
                          <div className="font-medium text-sm">{machineSize}</div>
                          <div className="text-xs text-muted-foreground">
                            {machine.cpuCores}C / {machine.memoryGB}GB
                          </div>
                        </td>
                        <td className="p-4">
                          {machine.status === 'off' ? (
                            <div>
                              <div className="text-sm text-muted-foreground">Offline</div>
                              <div className="text-xs text-muted-foreground">-</div>
                            </div>
                          ) : (
                            <div>
                              <div className="font-medium text-sm">
                                {machine.activeWorkloads || 0} active
                              </div>
                              <div className="text-xs text-muted-foreground">
                                {machine.runningWorkspaces} workspaces / {machine.runningEnvironments} environments
                              </div>
                            </div>
                          )}
                        </td>
                        <td className="p-4">
                          <div className="flex items-center gap-1">
                            <MapPin className="h-3 w-3 text-muted-foreground" />
                            <span className="text-sm">{machine.location}</span>
                          </div>
                        </td>
                        <td className="p-4">
                          <Badge variant="outline" className={cn('text-xs font-medium border', statusColor)}>
                            <StatusIcon className="h-3 w-3 mr-1.5" />
                            {getStatusLabel(machine.status)}
                          </Badge>
                        </td>
                      </tr>
                    </TooltipTrigger>
                    <TooltipContent side="right" className="p-0 border-0 bg-transparent shadow-none">
                      <MetricsTooltip
                        title={machine.name}
                        cpuUsed={machine.usedCpuCores}
                        cpuTotal={machine.cpuCores}
                        memoryUsed={machine.usedMemoryGB}
                        memoryTotal={machine.memoryGB}
                        runningWorkspaces={machine.runningWorkspaces}
                        runningEnvironments={machine.runningEnvironments}
                      />
                    </TooltipContent>
                  </Tooltip>
                );
              })}
            </tbody>
          </table>
        </div>
      </TooltipProvider>
    </div>
  );
}