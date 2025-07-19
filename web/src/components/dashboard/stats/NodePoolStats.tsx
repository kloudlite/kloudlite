import {
  Server,
  CheckCircle,
  AlertTriangle,
  XCircle,
  MapPin,
  Cpu,
  HardDrive,
  TrendingUp,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { StatsCard } from './StatsCard';
import { NodePoolStats as NodePoolStatsType, IndividualNodePool } from '@/types/dashboard';
import { cn } from '@/lib/utils';

interface NodePoolStatsProps {
  stats: NodePoolStatsType;
  individualPools: IndividualNodePool[];
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
      return 'text-green-500 bg-green-500/10';
    case 'degraded':
      return 'text-yellow-500 bg-yellow-500/10';
    case 'failed':
      return 'text-destructive bg-destructive/10';
    default:
      return 'text-muted-foreground bg-muted';
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

export function NodePoolStats({ stats, individualPools }: NodePoolStatsProps) {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-bold">Node Pools</h3>
      
      {/* Node Pool Status Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title="Total Pools"
          value={stats.total}
          icon={Server}
          description="All node pools"
        />
        <StatsCard
          title="Healthy"
          value={stats.healthy}
          icon={CheckCircle}
          variant="success"
          description="Operating normally"
        />
        <StatsCard
          title="Degraded"
          value={stats.degraded}
          icon={AlertTriangle}
          variant="warning"
          description="Performance issues"
        />
        <StatsCard
          title="Failed"
          value={stats.failed}
          icon={XCircle}
          variant="error"
          description="Not operational"
        />
      </div>

      {/* Individual Node Pool Details */}
      <div>
        <h4 className="text-lg font-semibold mb-4">Individual Node Pool Details</h4>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {individualPools.map((pool) => {
            const StatusIcon = getStatusIcon(pool.status);
            const statusColor = getStatusColor(pool.status);
            
            return (
              <Card key={pool.id}>
                <CardHeader>
                  <CardTitle className="flex items-center justify-between">
                    <span className="flex items-center gap-2">
                      <div className={cn('p-1.5 rounded-md', statusColor)}>
                        <StatusIcon className="h-4 w-4" />
                      </div>
                      {pool.name}
                    </span>
                    <span className="text-sm font-normal text-muted-foreground">
                      {pool.nodeCount} nodes
                    </span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Pool Info */}
                  <div className="space-y-2">
                    <span className="font-mono text-xs bg-muted px-2 py-1 rounded">
                      {pool.instanceType}
                    </span>
                    {pool.autoscaleEnabled && (
                      <div className="flex items-center gap-1 text-sm text-primary">
                        <TrendingUp className="h-4 w-4" />
                        <span className="font-medium">{pool.minNodes}-{pool.maxNodes} nodes</span>
                      </div>
                    )}
                  </div>
                  
                  {/* Resource Utilization */}
                  <div className="space-y-3">
                    <ProgressBar
                      value={pool.usedCpuCores}
                      max={pool.cpuCores}
                      label="CPU Cores"
                    />
                    <ProgressBar
                      value={pool.usedMemoryGB}
                      max={pool.memoryGB}
                      label="Memory (GB)"
                    />
                  </div>
                  
                  {/* Utilization Summary */}
                  <div className="grid grid-cols-2 gap-2 pt-2 border-t">
                    <div className="text-center">
                      <div className="text-sm font-medium">
                        {pool.cpuUtilization.toFixed(1)}%
                      </div>
                      <div className="text-xs text-muted-foreground">CPU Usage</div>
                    </div>
                    <div className="text-center">
                      <div className="text-sm font-medium">
                        {pool.memoryUtilization.toFixed(1)}%
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
    </div>
  );
}