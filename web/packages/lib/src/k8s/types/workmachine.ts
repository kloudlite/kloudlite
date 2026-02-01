/**
 * WorkMachine CRD type definitions
 * Based on: api/internal/controllers/workmachine/v1/workmachine_types.go
 */

import type { K8sResource, K8sList, Condition, Toleration } from './common';

export const WorkMachineGroup = 'machines.kloudlite.io';
export const WorkMachineVersion = 'v1';
export const WorkMachinePlural = 'workmachines';

// WorkMachine spec types
export type CloudProvider = 'aws' | 'gcp' | 'azure';

export type MachineState =
  | 'running'
  | 'stopped'
  | 'starting'
  | 'stopping'
  | 'errored'
  | 'disabled';

export interface AutoStopConfig {
  enabled: boolean;
  idleMinutes: number;
}

export interface AutoShutdownConfig {
  enabled: boolean;
  idleThresholdMinutes: number;
  checkIntervalMinutes: number;
}

export interface MachineConfiguration {
  autoStop?: AutoStopConfig;
  timezone?: string;
}

export interface WorkMachineSpec {
  displayName: string;
  ownedBy: string;
  targetNamespace: string;
  state: MachineState;
  sshPublicKeys?: string[];
  machineType: string;
  volumeSize?: number;
  volumeType?: string;
  deleteVolumePostTermination?: boolean;
  autoShutdown?: AutoShutdownConfig;
}

// WorkMachine status types
export interface GPUInfo {
  hasGPU?: boolean;
  count?: number;
  product?: string;
  driverVersion?: string;
  runtimeConfigured?: boolean;
  driverInstallationStatus?: string;
  driverInstallationMessage?: string;
}

export interface MachineResources {
  cpu: string;
  memory: string;
  gpu?: string;
}

export interface MachineInfo {
  machineID?: string;
  state?: MachineState;
  storageVolumeSize?: number;
  publicIP?: string;
  privateIP?: string;
  region?: string;
  availabilityZone?: string;
  message?: string;
  hasGPU?: boolean;
  gpuModel?: string;
}

export interface ReconcilerStatus {
  isReady?: boolean;
  message?: Record<string, string>;
  checkList?: string[];
  lastReconcileTime?: string;
  resources?: {
    checkList?: string[];
    lastReconcileTime?: string;
  }[];
}

export interface WorkMachineStatus extends MachineInfo {
  status?: ReconcilerStatus;
  isReady?: boolean;
  message?: Record<string, string>;
  checkList?: string[];
  lastReconcileTime?: string;
  startedAt?: string;
  stoppedAt?: string;
  lastActivityAt?: string;
  accessURL?: string;
  allocatedResources?: MachineResources;
  gpu?: GPUInfo;
  sshPublicKey?: string;
  lastWorkspaceActivity?: string;
  activeWorkspaceCount?: number;
  isAutoStopped?: boolean;
  allIdleSince?: string;
  nodeLabels?: Record<string, string>;
  podTolerations?: Toleration[];
  currentMachineType?: string;
  machineTypeChanging?: boolean;
  machineTypeChangeMessage?: string;
}

// Main WorkMachine resource (cluster-scoped)
export interface WorkMachine extends K8sResource<WorkMachineSpec, WorkMachineStatus> {
  apiVersion: 'machines.kloudlite.io/v1';
  kind: 'WorkMachine';
}

export interface WorkMachineList extends K8sList<WorkMachine> {
  apiVersion: 'machines.kloudlite.io/v1';
  kind: 'WorkMachineList';
}

// Helper types
export type WorkMachineCreateInput = Omit<WorkMachine, 'apiVersion' | 'kind' | 'status'> & {
  apiVersion?: string;
  kind?: string;
};

export type WorkMachineUpdateInput = Partial<WorkMachineSpec>;

// ============================================================================
// MachineType - Predefined machine configurations
// ============================================================================

export const MachineTypeGroup = 'machines.kloudlite.io';
export const MachineTypeVersion = 'v1';
export const MachineTypePlural = 'machinetypes';

export type MachineCategory =
  | 'general'
  | 'compute-optimized'
  | 'memory-optimized'
  | 'gpu'
  | 'development';

export interface MachineTypeToleration {
  key?: string;
  operator?: 'Exists' | 'Equal';
  value?: string;
  effect?: 'NoSchedule' | 'PreferNoSchedule' | 'NoExecute';
}

export interface MachineTypeSpec {
  displayName: string;
  description?: string;
  category: MachineCategory;
  resources: MachineResources;
  active: boolean;
  isDefault: boolean;
  priority?: number;
  podLabels?: Record<string, string>;
  podAnnotations?: Record<string, string>;
  nodeSelector?: Record<string, string>;
  tolerations?: MachineTypeToleration[];
}

export type MachineTypeConditionType = 'Ready' | 'Validated';

export interface MachineTypeCondition extends Condition {
  type: MachineTypeConditionType;
}

export interface MachineTypeStatus {
  inUseCount?: number;
  lastUpdated?: string;
  conditions?: MachineTypeCondition[];
}

// Main MachineType resource (cluster-scoped)
export interface MachineType extends K8sResource<MachineTypeSpec, MachineTypeStatus> {
  apiVersion: 'machines.kloudlite.io/v1';
  kind: 'MachineType';
}

export interface MachineTypeList extends K8sList<MachineType> {
  apiVersion: 'machines.kloudlite.io/v1';
  kind: 'MachineTypeList';
}

// Helper types
export type MachineTypeCreateInput = Omit<MachineType, 'apiVersion' | 'kind' | 'status'> & {
  apiVersion?: string;
  kind?: string;
};

export type MachineTypeUpdateInput = Partial<MachineTypeSpec>;
