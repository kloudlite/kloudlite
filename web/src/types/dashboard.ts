export interface DashboardStats {
  workspaces: WorkspaceStats;
  environments: EnvironmentStats;
  services: ServiceStats;
  workMachines: WorkMachineStats;
}

// Simple user dashboard - only 4 cards
export interface UserDashboardStats {
  myWorkspaces: {
    online: number;
    offline: number;
  };
  myEnvironments: {
    online: number;
    offline: number;
  };
  totalEnvironments: {
    online: number;
    offline: number;
  };
}

// Admin dashboard - comprehensive team view
export interface AdminDashboardStats {
  workspaces: WorkspaceStats;
  developers: DeveloperStats;
  workMachines: WorkMachineStats;
  nodePools: NodePoolStats;
  totalResources: TotalResourceStats;
  individualWorkMachines: IndividualWorkMachine[];
  individualNodePools: IndividualNodePool[];
}

export interface WorkspaceStats {
  total: number;
  running: number;
  stopped: number;
  archived: number;
}

export interface EnvironmentStats {
  total: number;
  active: number;
  deploying: number;
  failed: number;
  stopped: number;
}

export interface ServiceStats {
  sharedServices: number;
  externalServices: number;
  healthy: number;
  unhealthy: number;
  recentlyAdded: number;
}


export interface Activity {
  id: string;
  type: ActivityType;
  title: string;
  description: string;
  user: {
    name: string;
    email: string;
    avatar?: string;
  };
  timestamp: Date;
  metadata?: Record<string, any>;
  status: ActivityStatus;
}

export type ActivityType =
  | 'environment.created'
  | 'environment.started'
  | 'environment.stopped'
  | 'environment.deleted'
  | 'environment.deployed'
  | 'workspace.created'
  | 'workspace.started'
  | 'workspace.stopped'
  | 'workspace.archived'
  | 'workspace.deleted'
  | 'user.joined'
  | 'user.role_changed'
  | 'user.removed'
  | 'service.shared.created'
  | 'service.shared.updated'
  | 'service.shared.deleted'
  | 'service.external.created'
  | 'service.external.updated'
  | 'service.external.deleted'
  | 'workmachine.status_changed'
  | 'workmachine.capacity_updated';

export type ActivityStatus = 'success' | 'pending' | 'failed' | 'warning';

export interface ActivityFilter {
  types?: ActivityType[];
  users?: string[];
  dateRange?: {
    start: Date;
    end: Date;
  };
  status?: ActivityStatus[];
}

// Dashboard mode toggle
export type DashboardMode = 'user' | 'admin';

// User role for checking admin permissions
export type UserRole = 'user' | 'admin' | 'owner';

export interface UserContext {
  role: UserRole;
  name: string;
  email: string;
}

// Additional interfaces for comprehensive admin dashboard
export interface DeveloperStats {
  total: number;
  active: number;
  admins: number;
  online: number;
}

export interface NodePoolStats {
  total: number;
  healthy: number;
  degraded: number;
  failed: number;
}

export interface TotalResourceStats {
  totalCpuCores: number;
  totalMemoryGB: number;
  usedCpuCores: number;
  usedMemoryGB: number;
  cpuUtilization: number;
  memoryUtilization: number;
}

export interface WorkMachineStats {
  total: number;
  on: number;
  off: number;
  totalCpuCores: number;
  totalMemoryGB: number;
  usedCpuCores: number;
  usedMemoryGB: number;
  runningEnvironments: number;
  runningWorkspaces: number;
  cpuUtilization: number;
  memoryUtilization: number;
}

export interface IndividualWorkMachine {
  id: string;
  developerName: string;
  status: 'on' | 'off';
  vmSize: string;
  cpuCores: number;
  memoryGB: number;
  storageGB: number;
  usedCpuCores: number;
  usedMemoryGB: number;
  cpuUtilization: number;
  memoryUtilization: number;
  runningWorkspaces: number;
  runningEnvironments: number;
  // Infrastructure machines don't have personal owners
  activeWorkloads?: number; // Number of active workloads
}


export interface IndividualNodePool {
  id: string;
  name: string;
  status: 'healthy' | 'degraded' | 'failed';
  nodeCount: number;
  vmSize: string;
  cpuPerNode: number;
  memoryPerNodeGB: number;
  autoscaleEnabled: boolean;
  minNodes: number;
  maxNodes: number;
  nodeSelector: Record<string, string>;
  region: string;
}