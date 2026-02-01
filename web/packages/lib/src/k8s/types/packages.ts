/**
 * PackageRequest CRD type definitions
 * Based on: api/internal/controllers/packages/v1/packagerequest_types.go
 */

import type { K8sResource, K8sList } from './common';

export const PackageRequestGroup = 'packages.kloudlite.io';
export const PackageRequestVersion = 'v1';
export const PackageRequestPlural = 'packagerequests';

// Package types
export interface PackageSpec {
  name: string;
  channel?: string;
  nixpkgsCommit?: string;
}

export interface PackageRequestSpec {
  workspaceRef: string;
  packages?: PackageSpec[];
  profileName: string;
}

export type PackageRequestPhase = 'Pending' | 'Installing' | 'Ready' | 'Failed';

export interface PackageRequestStatus {
  observedGeneration?: number;
  phase?: PackageRequestPhase;
  message?: string;
  profileStorePath?: string;
  packagesPath?: string;
  specHash?: string;
  packageCount?: number;
  packages?: string[];
  failedPackage?: string;
  lastUpdated?: string;
}

// Main PackageRequest resource (namespaced)
export interface PackageRequest extends K8sResource<PackageRequestSpec, PackageRequestStatus> {
  apiVersion: 'packages.kloudlite.io/v1';
  kind: 'PackageRequest';
}

export interface PackageRequestList extends K8sList<PackageRequest> {
  apiVersion: 'packages.kloudlite.io/v1';
  kind: 'PackageRequestList';
}

// Helper types
export type PackageRequestCreateInput = Omit<PackageRequest, 'apiVersion' | 'kind' | 'status'> & {
  apiVersion?: string;
  kind?: string;
};

export type PackageRequestUpdateInput = Partial<PackageRequestSpec>;
