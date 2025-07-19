import { Cpu, HardDrive, Layers, FolderOpen } from 'lucide-react';
import { cn } from '@/lib/utils';

interface MetricsTooltipProps {
  title: string;
  cpuUsed: number;
  cpuTotal: number;
  memoryUsed: number;
  memoryTotal: number;
  runningWorkspaces?: number;
  runningEnvironments?: number;
  className?: string;
}

function ProgressBar({ 
  value, 
  max, 
  label, 
  unit = '' 
}: { 
  value: number; 
  max: number; 
  label: string; 
  unit?: string; 
}) {
  const percentage = (value / max) * 100;
  const colorClass = percentage > 80 ? 'bg-destructive' : percentage > 60 ? 'bg-yellow-500' : 'bg-green-500';
  
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-xs">
        <span className="text-foreground/70">{label}</span>
        <span className="font-medium text-foreground">
          {value}{unit} / {max}{unit}
        </span>
      </div>
      <div className="w-full bg-muted rounded-full h-1.5">
        <div
          className={cn('h-1.5 rounded-full transition-all duration-300', colorClass)}
          style={{ width: `${Math.min(percentage, 100)}%` }}
        />
      </div>
      <div className="text-xs text-foreground/60 text-right">
        {percentage.toFixed(1)}% utilized
      </div>
    </div>
  );
}

export function MetricsTooltip({
  title,
  cpuUsed,
  cpuTotal,
  memoryUsed,
  memoryTotal,
  runningWorkspaces,
  runningEnvironments,
  className,
}: MetricsTooltipProps) {
  return (
    <div className={cn(
      'p-3 bg-background border rounded-md space-y-3 min-w-[240px]',
      className
    )}>
      <div className="font-medium text-sm border-b pb-2 text-foreground">{title}</div>
      
      {/* Resource Utilization */}
      <div className="space-y-3">
        <ProgressBar
          value={cpuUsed}
          max={cpuTotal}
          label="CPU Cores"
        />
        <ProgressBar
          value={memoryUsed}
          max={memoryTotal}
          label="Memory"
          unit="GB"
        />
      </div>
      
      {/* Running Workloads */}
      {(runningWorkspaces !== undefined || runningEnvironments !== undefined) && (
        <div className="pt-2 border-t">
          <div className="grid grid-cols-2 gap-3">
            {runningWorkspaces !== undefined && (
              <div className="flex items-center gap-1.5">
                <FolderOpen className="h-3 w-3 text-muted-foreground" />
                <span className="text-xs text-foreground/70">Workspaces</span>
                <span className="text-xs font-medium text-foreground">{runningWorkspaces}</span>
              </div>
            )}
            {runningEnvironments !== undefined && (
              <div className="flex items-center gap-1.5">
                <Layers className="h-3 w-3 text-muted-foreground" />
                <span className="text-xs text-foreground/70">Environments</span>
                <span className="text-xs font-medium text-foreground">{runningEnvironments}</span>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}