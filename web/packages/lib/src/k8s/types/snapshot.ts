/**
 * Snapshot CRD type definitions
 * Based on: api/internal/controllers/snapshot/v1/types.go
 */

import type { K8sResource, K8sList } from './common';

export const SnapshotGroup = 'snapshots.kloudlite.io';
export const SnapshotVersion = 'v1';
export const SnapshotPlural = 'snapshots';

// Snapshot types
export type ArtifactType =
  | 'kubernetes-manifests'
  | 'json'
  | 'yaml'
  | 'sql'
  | 'raw';

export interface ArtifactSpec {
  name: string;
  type?: ArtifactType;
  data?: string;
}

export interface RetentionPolicy {
  expiresAt?: string;
  keepForDays?: number;
}

export interface SnapshotSpec {
  owner: string;
  parentSnapshot?: string;
  description?: string;
  artifacts?: ArtifactSpec[];
  retentionPolicy?: RetentionPolicy;
}

export type SnapshotState = 'Ready' | 'Deleting' | 'Failed';

export interface SnapshotRegistryInfo {
  imageRef: string;
  digest?: string;
  pushedAt?: string;
  compressedSize?: number;
  parentImageRef?: string;
}

export interface ArtifactStatus {
  name: string;
  type: ArtifactType;
  sizeBytes?: number;
  stored: boolean;
}

export interface SnapshotStatus {
  state?: SnapshotState;
  message?: string;
  sizeBytes?: number;
  sizeHuman?: string;
  createdAt?: string;
  lineage?: string[];
  storageRefs?: string[];
  registry?: SnapshotRegistryInfo;
  artifacts?: ArtifactStatus[];
}

// Main Snapshot resource (namespaced)
export interface Snapshot extends K8sResource<SnapshotSpec, SnapshotStatus> {
  apiVersion: 'snapshots.kloudlite.io/v1';
  kind: 'Snapshot';
}

export interface SnapshotList extends K8sList<Snapshot> {
  apiVersion: 'snapshots.kloudlite.io/v1';
  kind: 'SnapshotList';
}

// Helper types
export type SnapshotCreateInput = Omit<Snapshot, 'apiVersion' | 'kind' | 'status'> & {
  apiVersion?: string;
  kind?: string;
};

export type SnapshotUpdateInput = Partial<SnapshotSpec>;
