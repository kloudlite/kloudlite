'use client';

import {
  CheckCircle,
  AlertTriangle,
  XCircle,
  MapPin,
  Server,
  TrendingUp,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { MetricsTooltip } from './MetricsTooltip';
import { IndividualNodePool } from '@/types/dashboard';
import { cn } from '@/lib/utils';

interface CleanNodePoolGridProps {
  pools: IndividualNodePool[];
}

function getStatusIcon(status: IndividualNodePool['status']) {
  switch (status) {
    case 'healthy':
      return CheckCircle;
    case 'degraded':
      return AlertTriangle;
    case 'failed':
      return XCircle;
    default:
      return Server;
  }
}

function getStatusColor(status: IndividualNodePool['status']) {
  switch (status) {
    case 'healthy':
      return 'bg-green-500/10 text-green-500 border-green-500/20 hover:bg-green-500/20';
    case 'degraded':
      return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20 hover:bg-yellow-500/20';
    case 'failed':
      return 'bg-destructive/10 text-destructive border-destructive/20 hover:bg-destructive/20';
    default:
      return 'bg-muted text-muted-foreground border-border hover:bg-muted/80';
  }
}

function getStatusLabel(status: IndividualNodePool['status']) {
  switch (status) {
    case 'healthy':
      return 'Healthy';
    case 'degraded':
      return 'Degraded';
    case 'failed':
      return 'Failed';
    default:
      return 'Unknown';
  }
}

export function CleanNodePoolGrid({ pools }: CleanNodePoolGridProps) {
  return (
    <div className="space-y-6">
      <div className="border-b border-border pb-4">
        <h2 className="text-2xl font-bold tracking-tight">Node Pools</h2>
        <p className="text-base text-muted-foreground mt-1">
          Kubernetes node pools providing compute resources for workloads
        </p>
      </div>
      
      <TooltipProvider>
        <div className="overflow-hidden rounded-lg border">
          <table className="w-full">
            <thead className="bg-muted/50">
              <tr className="border-b">
                <th className="text-left p-4 font-medium text-sm">Pool Name</th>
                <th className="text-left p-4 font-medium text-sm">Nodes</th>
                <th className="text-left p-4 font-medium text-sm">Instance Type</th>
                <th className="text-left p-4 font-medium text-sm">Per Node</th>
                <th className="text-left p-4 font-medium text-sm">Utilization</th>
                <th className="text-left p-4 font-medium text-sm">Status</th>
              </tr>
            </thead>
            <tbody>
              {pools.map((pool, index) => {
                const StatusIcon = getStatusIcon(pool.status);
                const statusColor = getStatusColor(pool.status);
                
                return (
                  <Tooltip key={pool.id}>
                    <TooltipTrigger asChild>
                      <tr className={cn(
                        "border-b hover:bg-muted/40 cursor-pointer transition-colors",
                        index % 2 === 0 ? "bg-background" : "bg-muted/20"
                      )}>
                        <td className="p-4">
                          <div className="font-medium text-sm">{pool.name}</div>
                          <div className="text-xs text-muted-foreground">{pool.id}</div>
                        </td>
                        <td className="p-4">
                          <div className="font-medium text-sm">{pool.nodeCount} nodes</div>
                          <div className="text-xs text-muted-foreground">
                            {pool.cpuCores}C / {pool.memoryGB}GB total
                          </div>
                        </td>
                        <td className="p-4">
                          <div className="font-medium text-sm">{pool.instanceType}</div>
                          {pool.autoscaleEnabled && (
                            <div className="flex items-center gap-1 text-xs text-primary mt-1">
                              <TrendingUp className="h-3 w-3" />
                              <span>{pool.minNodes}-{pool.maxNodes} nodes</span>
                            </div>
                          )}
                        </td>
                        <td className="p-4">
                          <div className="text-xs text-muted-foreground">
                            {(pool.cpuCores / pool.nodeCount)}C / {(pool.memoryGB / pool.nodeCount)}GB per node
                          </div>
                        </td>
                        <td className="p-4">
                          <div className="flex items-center gap-3">
                            <div className="text-center">
                              <div className={`text-sm font-medium ${pool.cpuUtilization > 80 ? 'text-destructive' : pool.cpuUtilization > 60 ? 'text-yellow-500' : 'text-green-500'}`}>
                                {pool.cpuUtilization.toFixed(1)}%
                              </div>
                              <div className="text-xs text-foreground/70">CPU</div>
                            </div>
                            <div className="text-center">
                              <div className={`text-sm font-medium ${pool.memoryUtilization > 80 ? 'text-destructive' : pool.memoryUtilization > 60 ? 'text-yellow-500' : 'text-green-500'}`}>
                                {pool.memoryUtilization.toFixed(1)}%
                              </div>
                              <div className="text-xs text-foreground/70">Memory</div>
                            </div>
                          </div>
                        </td>
                        <td className="p-4">
                          <Badge variant="outline" className={cn('text-xs font-medium border', statusColor)}>
                            <StatusIcon className="h-3 w-3 mr-1.5" />
                            {getStatusLabel(pool.status)}
                          </Badge>
                        </td>
                      </tr>
                    </TooltipTrigger>
                    <TooltipContent side="right" className="p-0 border-0 bg-transparent shadow-none">
                      <MetricsTooltip
                        title={pool.name}
                        cpuUsed={pool.usedCpuCores}
                        cpuTotal={pool.cpuCores}
                        memoryUsed={pool.usedMemoryGB}
                        memoryTotal={pool.memoryGB}
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