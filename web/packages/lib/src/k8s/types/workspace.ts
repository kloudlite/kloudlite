/**
 * Workspace CRD type definitions
 * Based on: api/internal/controllers/workspace/v1/workspace_types.go
 */

import type { K8sResource, K8sList, Condition } from './common';

export const WorkspaceGroup = 'workspaces.kloudlite.io';
export const WorkspaceVersion = 'v1';
export const WorkspacePlural = 'workspaces';

// Workspace spec types
export interface EnvironmentConnectionSpec {
  environmentRef: {
    kind?: string;
    namespace?: string;
    name: string;
    uid?: string;
    apiVersion?: string;
  };
}

export interface ExposedPort {
  port: number;
}

export interface GitRepository {
  url: string;
  branch?: string;
}

export interface WorkspaceSettings {
  autoStop?: boolean;
  idleTimeout?: number;
  maxRuntime?: number;
  startupScript?: string;
  environmentVariables?: Record<string, string>;
  gitConfig?: GitConfig;
  vscodeExtensions?: string[];
  dotfilesRepo?: string;
}

export interface GitConfig {
  userName?: string;
  userEmail?: string;
  defaultBranch?: string;
}

export interface ResourceQuota {
  cpu?: string;
  memory?: string;
  storage?: string;
  gpus?: number;
}

export interface WorkspaceFromSnapshotRef {
  snapshotName: string;
}

export interface WorkspaceSpec {
  displayName: string;
  description?: string;
  ownedBy: string;
  visibility?: 'private' | 'shared' | 'open';
  sharedWith?: string[];
  workmachine: string;
  environmentConnection?: EnvironmentConnectionSpec;
  gitRepository?: GitRepository;
  settings?: WorkspaceSettings;
  tags?: string[];
  resourceQuota?: ResourceQuota;
  vscodeVersion?: string;
  status?: 'active' | 'suspended' | 'archived';
  fromSnapshot?: WorkspaceFromSnapshotRef;
  expose?: ExposedPort[];
}

// Workspace status types
export interface ConnectedEnvironmentInfo {
  name: string;
  targetNamespace: string;
  availableServices?: string[];
}

export interface WorkspaceResourceUsage {
  cpu?: string;
  memory?: string;
  storage?: string;
  lastUpdated?: string;
}

export interface WorkspaceLastRestoredSnapshotInfo {
  name: string;
  restoredAt: string;
}

export type WorkspaceSnapshotRestorePhase =
  | 'Pending'
  | 'Pulling'
  | 'Restoring'
  | 'Completed'
  | 'Failed';

export interface WorkspaceSnapshotRestoreStatus {
  phase?: WorkspaceSnapshotRestorePhase;
  message?: string;
  sourceSnapshot?: string;
  imageRef?: string;
  startTime?: string;
  completionTime?: string;
  errorMessage?: string;
}

export type WorkspacePhase =
  | 'Pending'
  | 'Creating'
  | 'Running'
  | 'Stopping'
  | 'Stopped'
  | 'Failed'
  | 'Terminating';

export interface WorkspaceStatus {
  phase?: WorkspacePhase;
  conditions?: Condition[];
  message?: string;
  lastActivityTime?: string;
  startTime?: string;
  stopTime?: string;
  accessUrl?: string; // deprecated
  accessUrls?: Record<string, string>;
  resourceUsage?: WorkspaceResourceUsage;
  totalRuntime?: number;
  podName?: string;
  podIP?: string;
  nodeName?: string;
  activeConnections?: number;
  idleState?: 'active' | 'idle';
  idleSince?: string;
  connectedEnvironment?: ConnectedEnvironmentInfo;
  snapshotRestoreStatus?: WorkspaceSnapshotRestoreStatus;
  hash?: string;
  subdomain?: string;
  exposedRoutes?: Record<string, string>;
  lastRestoredSnapshot?: WorkspaceLastRestoredSnapshotInfo;
}

// Main Workspace resource
export interface Workspace extends K8sResource<WorkspaceSpec, WorkspaceStatus> {
  apiVersion: 'workspaces.kloudlite.io/v1';
  kind: 'Workspace';
}

export interface WorkspaceList extends K8sList<Workspace> {
  apiVersion: 'workspaces.kloudlite.io/v1';
  kind: 'WorkspaceList';
}

// Helper type for creating workspaces
export type WorkspaceCreateInput = Omit<Workspace, 'apiVersion' | 'kind' | 'status'> & {
  apiVersion?: string;
  kind?: string;
};

// Helper type for updating workspaces
export type WorkspaceUpdateInput = Partial<WorkspaceSpec>;
