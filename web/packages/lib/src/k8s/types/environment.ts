/**
 * Environment CRD type definitions
 * Based on: api/internal/controllers/environment/v1/types.go
 */

import type { K8sResource, K8sList, Condition } from './common';

export const EnvironmentGroup = 'environments.kloudlite.io';
export const EnvironmentVersion = 'v1';
export const EnvironmentPlural = 'environments';

// Environment spec types
export interface ResourceQuotas {
  'limits.cpu'?: string;
  'limits.memory'?: string;
  'requests.cpu'?: string;
  'requests.memory'?: string;
  persistentvolumeclaims?: string;
  'services.nodeports'?: string;
  'services.loadbalancers'?: string;
}

export interface LabelSelector {
  matchLabels?: Record<string, string>;
}

export interface NetworkPolicyPeer {
  namespaceSelector?: LabelSelector;
  podSelector?: LabelSelector;
}

export interface NetworkPolicyPort {
  protocol?: 'TCP' | 'UDP';
  port?: number;
}

export interface IngressRule {
  from?: NetworkPolicyPeer[];
  ports?: NetworkPolicyPort[];
}

export interface NetworkPolicies {
  enabled: boolean;
  allowedNamespaces?: string[];
  ingressRules?: IngressRule[];
}

export interface FromSnapshotRef {
  snapshotName: string;
  sourceNamespace: string;
}

export interface CompositionSpec {
  // This will be defined in composition_types.ts
  // Placeholder for now
  [key: string]: any;
}

export interface EnvironmentSpec {
  targetNamespace?: string;
  ownedBy?: string;
  visibility?: 'private' | 'shared' | 'open';
  sharedWith?: string[];
  workmachineName?: string;
  activated: boolean;
  resourceQuotas?: ResourceQuotas;
  networkPolicies?: NetworkPolicies;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  fromSnapshot?: FromSnapshotRef;
  nodeName?: string;
  compose?: CompositionSpec;
}

// Environment status types
export type EnvironmentState =
  | 'active'
  | 'inactive'
  | 'activating'
  | 'deactivating'
  | 'snapping'
  | 'deleting'
  | 'error';

export interface ResourceCount {
  deployments?: number;
  statefulsets?: number;
  services?: number;
  configmaps?: number;
  secrets?: number;
  pvcs?: number;
}

export type SnapshotRestorePhase =
  | 'Pending'
  | 'Pulling'
  | 'Restoring'
  | 'DataRestoring'
  | 'Completed'
  | 'Failed';

export interface SnapshotRestoreStatus {
  phase?: SnapshotRestorePhase;
  message?: string;
  sourceSnapshot?: string;
  imageRef?: string;
  startTime?: string;
  completionTime?: string;
  errorMessage?: string;
}

export interface LastRestoredSnapshotInfo {
  name: string;
  restoredAt: string;
  lineage?: string[];
}

export interface CompositionStatus {
  // This will be defined in composition_types.ts
  // Placeholder for now
  [key: string]: any;
}

export type EnvironmentConditionType =
  | 'Ready'
  | 'NamespaceCreated'
  | 'ResourceQuotaApplied'
  | 'NetworkPolicyApplied'
  | 'Forked';

export interface EnvironmentCondition extends Condition {
  type: EnvironmentConditionType;
}

export interface EnvironmentStatus {
  state?: EnvironmentState;
  message?: string;
  lastActivatedTime?: string;
  lastDeactivatedTime?: string;
  resourceCount?: ResourceCount;
  conditions?: EnvironmentCondition[];
  observedGeneration?: number;
  snapshotRestoreStatus?: SnapshotRestoreStatus;
  hash?: string;
  subdomain?: string;
  lastRestoredSnapshot?: LastRestoredSnapshotInfo;
  composeStatus?: CompositionStatus;
}

// Main Environment resource
export interface Environment extends K8sResource<EnvironmentSpec, EnvironmentStatus> {
  apiVersion: 'environments.kloudlite.io/v1';
  kind: 'Environment';
}

export interface EnvironmentList extends K8sList<Environment> {
  apiVersion: 'environments.kloudlite.io/v1';
  kind: 'EnvironmentList';
}

// Helper types
export type EnvironmentCreateInput = Omit<Environment, 'apiVersion' | 'kind' | 'status'> & {
  apiVersion?: string;
  kind?: string;
};

export type EnvironmentUpdateInput = Partial<EnvironmentSpec>;

// ============================================================================
// EnvironmentSnapshotRequest - Orchestrates creating a snapshot for an environment
// ============================================================================

export interface EnvironmentSnapshotRequestSpec {
  environmentName: string;
  environmentNamespace: string;
  snapshotName: string;
  description?: string;
  retentionDays?: number;
}

export type EnvironmentSnapshotRequestPhase =
  | 'Pending'
  | 'StoppingWorkloads'
  | 'WaitingForPods'
  | 'CreatingSnapshot'
  | 'UploadingSnapshot'
  | 'RestoringEnvironment'
  | 'Completed'
  | 'Failed';

export interface EnvironmentSnapshotRequestStatus {
  phase?: EnvironmentSnapshotRequestPhase;
  message?: string;
  previousEnvironmentState?: EnvironmentState;
  snapshotRequestName?: string;
  createdSnapshotName?: string;
  startTime?: string;
  completionTime?: string;
}

export interface EnvironmentSnapshotRequest
  extends K8sResource<EnvironmentSnapshotRequestSpec, EnvironmentSnapshotRequestStatus> {
  apiVersion: 'environments.kloudlite.io/v1';
  kind: 'EnvironmentSnapshotRequest';
}

export interface EnvironmentSnapshotRequestList extends K8sList<EnvironmentSnapshotRequest> {
  apiVersion: 'environments.kloudlite.io/v1';
  kind: 'EnvironmentSnapshotRequestList';
}

// ============================================================================
// EnvironmentSnapshotRestore - Orchestrates restoring an environment from a snapshot
// ============================================================================

export interface EnvironmentSnapshotRestoreSpec {
  environmentName: string;
  environmentNamespace: string;
  snapshotName: string;
  sourceNamespace: string;
  activateAfterRestore?: boolean;
}

export type EnvironmentSnapshotRestorePhase =
  | 'Pending'
  | 'StoppingWorkloads'
  | 'WaitingForPods'
  | 'Downloading'
  | 'RestoringData'
  | 'ApplyingArtifacts'
  | 'Activating'
  | 'Completed'
  | 'Failed';

export interface RestoredArtifactsInfo {
  configMapsRestored?: number;
  secretsRestored?: number;
}

export interface EnvironmentSnapshotRestoreStatus {
  phase?: EnvironmentSnapshotRestorePhase;
  message?: string;
  snapshotRestoreName?: string;
  startTime?: string;
  completionTime?: string;
  restoredArtifacts?: RestoredArtifactsInfo;
}

export interface EnvironmentSnapshotRestore
  extends K8sResource<EnvironmentSnapshotRestoreSpec, EnvironmentSnapshotRestoreStatus> {
  apiVersion: 'environments.kloudlite.io/v1';
  kind: 'EnvironmentSnapshotRestore';
}

export interface EnvironmentSnapshotRestoreList extends K8sList<EnvironmentSnapshotRestore> {
  apiVersion: 'environments.kloudlite.io/v1';
  kind: 'EnvironmentSnapshotRestoreList';
}

// ============================================================================
// EnvironmentForkRequest - Orchestrates forking an environment from a snapshot
// ============================================================================

export interface SourceSnapshotRef {
  snapshotName: string;
  sourceNamespace: string;
}

export interface EnvironmentSpecOverrides {
  visibility?: 'private' | 'shared' | 'open';
  ownedBy?: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  resourceQuotas?: ResourceQuotas;
}

export interface EnvironmentForkRequestSpec {
  newEnvironmentName: string;
  sourceSnapshot: SourceSnapshotRef;
  overrides?: EnvironmentSpecOverrides;
}

export type EnvironmentForkRequestPhase =
  | 'Pending'
  | 'Validating'
  | 'CreatingEnvironment'
  | 'WaitingForRestore'
  | 'Completed'
  | 'Failed';

export interface EnvironmentForkRequestStatus {
  phase?: EnvironmentForkRequestPhase;
  message?: string;
  createdEnvironment?: string;
  startTime?: string;
  completionTime?: string;
}

export interface EnvironmentForkRequest
  extends K8sResource<EnvironmentForkRequestSpec, EnvironmentForkRequestStatus> {
  apiVersion: 'environments.kloudlite.io/v1';
  kind: 'EnvironmentForkRequest';
}

export interface EnvironmentForkRequestList extends K8sList<EnvironmentForkRequest> {
  apiVersion: 'environments.kloudlite.io/v1';
  kind: 'EnvironmentForkRequestList';
}
