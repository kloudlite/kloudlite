import {
  Users,
  Database,
  Network,
  Layers,
  Monitor,
  Cpu,
  HardDrive,
  Zap,
  ChevronDown,
  ChevronRight,
  CodeXml,
  Combine,
  TrendingUp,
} from 'lucide-react';
import { AdminDashboardStats } from '@/types/dashboard';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { StatusDot } from '@/components/ui/status-dot';
import { CountDisplay } from '@/components/ui/count-display';
import { StatsCard } from '@/components/dashboard/stats/StatsCard';

interface AdminDashboardProps {
  stats: AdminDashboardStats;
}

export function AdminDashboard({ stats }: AdminDashboardProps) {
  const [expandedMachine, setExpandedMachine] = useState<string | null>(null);
  const initialMachinesCount = 3;
  const initialNodePoolsCount = 3;

  const toggleMachine = (machineId: string) => {
    if (expandedMachine === machineId) {
      setExpandedMachine(null);
    } else {
      setExpandedMachine(machineId);
    }
  };

  return (
    <div className="space-y-8">
      {/* Core Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatsCard
          title="Developers"
          value={stats.developers.total}
          icon={Users}
          variant="users"
          description={`${stats.developers.online} online • ${stats.developers.admins} admins`}
        />
        
        <StatsCard
          title="Workspaces"
          value={stats.workspaces.total}
          icon={CodeXml}
          variant="workspaces"
          description={`${stats.workspaces.running} running • ${stats.workspaces.stopped} stopped`}
        />
        
        <StatsCard
          title="Environments"
          value={56}
          icon={Layers}
          variant="environments"
          description="42 active • 14 stopped"
        />
        
        <StatsCard
          title="Shared Services"
          value={24}
          icon={Network}
          variant="services"
          description="18 healthy • 6 external"
        />
      </div>

      {/* Infrastructure Overview */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Work Machines */}
        <div className="bg-card border border-border rounded-lg transition-colors duration-200 hover:border-primary/20">
          <div className="p-6 border-b border-border">
            <div className="flex items-center justify-between">
              <div>
                <Button variant="link" className="p-0 h-auto font-bold text-xl text-foreground hover:text-primary">
                  Work Machines
                </Button>
                <p className="text-base text-muted-foreground">Development infrastructure</p>
              </div>
              <div className="h-8 w-8 bg-muted rounded-lg flex items-center justify-center">
                <Monitor className="h-4 w-4 text-muted-foreground" />
              </div>
            </div>
          </div>
          <div className="p-6 space-y-4">
            {stats.individualWorkMachines.slice(0, initialMachinesCount).map((machine) => {
              const isExpanded = expandedMachine === machine.id;
              const totalEnvironments = machine.runningEnvironments + 2; // Mock total
              const totalWorkspaces = machine.runningWorkspaces + 1; // Mock total
              
              return (
                <div key={machine.id} className={cn(
                  "overflow-hidden transition-all duration-200 ease-out border rounded-lg",
                  isExpanded ? "border-border" : "border-transparent"
                )}>
                  <div
                    className={cn(
                      "w-full flex justify-between p-4 rounded-lg transition-colors duration-200",
                      machine.status === 'on' ? "hover:bg-muted/30 cursor-pointer" : "cursor-default"
                    )}
                    onClick={() => machine.status === 'on' && toggleMachine(machine.id)}
                  >
                    <div className="flex items-center gap-3">
                      <StatusDot status={machine.status === 'on' ? 'online' : 'offline'} />
                      <div className="text-left">
                        <p className="text-sm font-medium text-foreground">{machine.developerName}</p>
                        <p className="text-xs text-muted-foreground">{machine.vmSize}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3">
                      <div className="text-right">
                        {machine.status === 'on' ? (
                          <div className="flex items-center gap-6">
                            <div className="flex items-center gap-2">
                              <Layers className="h-4 w-4 text-muted-foreground" />
                              <CountDisplay active={machine.runningEnvironments} total={totalEnvironments} />
                            </div>
                            <div className="flex items-center gap-2">
                              <CodeXml className="h-4 w-4 text-muted-foreground" />
                              <CountDisplay active={machine.runningWorkspaces} total={totalWorkspaces} />
                            </div>
                          </div>
                        ) : (
                          <p className="text-sm font-medium text-muted-foreground">Offline</p>
                        )}
                        <p className="text-xs text-muted-foreground mt-1">
                          {machine.cpuCores}C / {machine.memoryGB}GB / {machine.storageGB}GB
                        </p>
                      </div>
                    </div>
                  </div>
                  
                  {/* Expanded Metrics */}
                  <div className={cn(
                    "overflow-hidden transition-all duration-200 ease-out",
                    isExpanded && machine.status === 'on' ? "max-h-32 opacity-100 mt-0 pt-4 border-t border-border" : "max-h-0 opacity-0"
                  )}>
                    <div className="grid grid-cols-3 gap-3 px-4 pb-4">
                      <div className="text-center p-3 bg-muted rounded transition-colors duration-200 hover:bg-muted/60">
                        <div className="flex items-center justify-center mb-1">
                          <Cpu className="h-4 w-4 text-primary" />
                        </div>
                        <p className="text-xs text-muted-foreground mb-1">CPU</p>
                        <p className="text-sm font-bold text-foreground">{machine.cpuUtilization.toFixed(0)}%</p>
                      </div>
                      <div className="text-center p-3 bg-muted rounded transition-colors duration-200 hover:bg-muted/60">
                        <div className="flex items-center justify-center mb-1">
                          <HardDrive className="h-4 w-4 text-primary" />
                        </div>
                        <p className="text-xs text-muted-foreground mb-1">Memory</p>
                        <p className="text-sm font-bold text-foreground">{(machine.usedMemoryGB).toFixed(1)}GB</p>
                      </div>
                      <div className="text-center p-3 bg-muted rounded transition-colors duration-200 hover:bg-muted/60">
                        <div className="flex items-center justify-center mb-1">
                          <Zap className="h-4 w-4 text-primary" />
                        </div>
                        <p className="text-xs text-muted-foreground mb-1">Uptime</p>
                        <p className="text-sm font-bold text-foreground">
                          {machine.status === 'on' ? '3h 24m' : '0m'}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              );
            })}
            
          </div>
        </div>

        {/* Node Pools */}
        <div className="bg-card border border-border rounded-lg transition-colors duration-200 hover:border-primary/20">
          <div className="p-6 border-b border-border">
            <div className="flex items-center justify-between">
              <div>
                <Button variant="link" className="p-0 h-auto font-bold text-xl text-foreground hover:text-primary">
                  Node Pools
                </Button>
                <p className="text-base text-muted-foreground">Kubernetes clusters</p>
              </div>
              <div className="h-8 w-8 bg-muted rounded-lg flex items-center justify-center">
                <Combine className="h-4 w-4 text-muted-foreground" />
              </div>
            </div>
          </div>
          <div className="p-6 space-y-4">
            {stats.individualNodePools.slice(0, initialNodePoolsCount).map((pool) => (
              <div key={pool.id} className="flex items-center justify-between py-3 transition-colors duration-200 hover:bg-muted/20 rounded-lg px-2 -mx-2">
                <div className="flex items-center gap-3">
                  <StatusDot status={pool.status} />
                  <div>
                    <p className="text-sm font-medium text-foreground">{pool.name}</p>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <span>{pool.vmSize}</span>
                      {pool.autoscaleEnabled && (
                        <>
                          <span>•</span>
                          <div className="flex items-center gap-1 text-primary">
                            <TrendingUp className="h-3 w-3" />
                            <span>{pool.minNodes}-{pool.maxNodes} nodes</span>
                          </div>
                        </>
                      )}
                    </div>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium text-foreground">
                    {pool.nodeCount} nodes
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {pool.cpuPerNode}C / {pool.memoryPerNodeGB}GB
                  </p>
                </div>
              </div>
            ))}
            
          </div>
        </div>
      </div>
    </div>
  );
}