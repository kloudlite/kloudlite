/* eslint-disable no-prototype-builtins */
import { Params } from '@remix-run/react';
import { dayjs } from '~/components/molecule/dayjs';
import { FlatMapType, NonNullableString } from '~/root/lib/types/common';
import {
  Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider as CloudProvider,
  Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecAvailabilityMode as AvailabilityMode,
  ProjectId,
  Github__Com___Kloudlite___Api___Pkg___Types__SyncStatusAction as SyncStatusAction,
  Github__Com___Kloudlite___Api___Pkg___Types__SyncStatusState as SyncStatusState,
  WorkspaceOrEnvId,
} from '~/root/src/generated/gql/server';

type IparseNodes<T> = {
  edges: Array<{ node: T }>;
};

interface IParamsCtx {
  params: Params<string>;
}

const getScopeQuery = (ctx: IParamsCtx): WorkspaceOrEnvId => {
  const { scope, workspace } = ctx.params;
  if (!workspace || !scope) {
    throw Error('scope and workspace is required, which is not provided');
  }
  return {
    value: workspace,
    type: scope === 'workspace' ? 'workspaceName' : 'environmentName',
  };
};

export const getProjectQuery = (ctx: IParamsCtx): ProjectId => {
  const { project } = ctx.params;
  if (!project) {
    throw Error(
      'project is required to render this page, which is not provide'
    );
  }
  return {
    type: 'name',
    value: project,
  };
};

export const getScopeAndProjectQuery = (
  ctx: IParamsCtx
): {
  project: ProjectId;
  scope: WorkspaceOrEnvId;
} => {
  return {
    project: getProjectQuery(ctx),
    scope: getScopeQuery(ctx),
  };
};

export const parseNodes = <T>(resources: IparseNodes<T> | undefined): T[] =>
  resources?.edges.map(({ node }) => node) || [];

type IparseName =
  | {
    metadata?: {
      name: string;
    };
  }
  | undefined
  | null;

export const parseName = (resource: IparseName, ensure = false) => {
  if (ensure) {
    if (!resource) {
      throw Error('resource not found');
    }

    if (!resource.metadata) {
      throw Error('metadata not found');
    }

    if (!resource.metadata.name) {
      throw Error('name not found');
    }
  }

  return resource?.metadata?.name || '';
};

type IparseNamespace =
  | {
    metadata: {
      namespace: string;
    };
  }
  | undefined
  | null;

export const parseNamespace = (resource: IparseNamespace) =>
  resource?.metadata.namespace || '';

type IparseTargetNs =
  | {
    spec?: {
      targetNamespace: string;
    };
  }
  | undefined
  | null;

export const parseTargetNs = (resource: IparseTargetNs) => {
  if (!resource) {
    throw Error('resource not found');
  }

  if (!resource.spec) {
    throw Error('spec not found');
  }

  return resource.spec.targetNamespace;
};

type parseFromAnnResource = {
  metadata?: {
    annotations?: FlatMapType<string>;
  };
} | null;

export const parseFromAnn = (resource: parseFromAnnResource, key: string) =>
  resource?.metadata?.annotations?.[key] || '';

export const validateClusterCloudProvider = (v: string): CloudProvider => {
  switch (v as CloudProvider) {
    case 'do':
    case 'aws':
    case 'azure':
    case 'gcp':
      return v as CloudProvider;
    default:
      throw Error(`invalid cloud provider type ${v}`);
  }
};

export const validateCloudProvider = (v: string): CloudProvider => {
  switch (v as CloudProvider) {
    case 'do':
    case 'aws':
    case 'azure':
    case 'gcp':
      return v as CloudProvider;
    default:
      throw Error(`invalid cloud provider type ${v}`);
  }
};

export const validateAvailabilityMode = (v: string): AvailabilityMode => {
  switch (v as AvailabilityMode) {
    case 'HA':
    case 'dev':
      return v as AvailabilityMode;
    default:
      throw Error(`invalid availabilityMode ${v}`);
  }
};

type Nodes = { edges: { node: any }[] };
export type ExtractNodeType<T> = T extends Nodes
  ? T['edges'][number]['node']
  : T;

export type IListOrGrid = 'list' | 'grid' | NonNullableString;
export type wsOrEnv = 'environment' | 'workspace' | NonNullableString;

export const parseUpdateTime = (resource: { updateTime: string }) => {
  return dayjs(resource.updateTime).fromNow();
};

export const parseCreationTime = (resource: { creationTime: string }) => {
  return dayjs(resource.creationTime).fromNow();
};

export const parseUpdateOrCreatedBy = (resource: {
  lastUpdatedBy: {
    userName: string;
  };
  createdBy: {
    userName: string;
  };
}) => {
  return resource.lastUpdatedBy.userName || resource.createdBy.userName;
};

export const parseUpdateOrCreatedOn = (resource: {
  updateTime: string;
  creationTime: string;
}) => {
  return dayjs(resource.updateTime || resource.creationTime).fromNow();
};

export function filterExtraFields(obj: any, schema: any): any {
  const result: any = {};

  for (const key in schema) {
    if (obj.hasOwnProperty(key)) {
      result[key] = obj[key];
    }
  }

  return result;
}

export interface Status {
  lastReconcileTime?: any;
  isReady: boolean;
  checks?: any;
  message?: { RawMessage?: any };
}

export interface SyncStatus {
  syncScheduledAt?: any;
  state: SyncStatusState;
  recordVersion: number;
  lastSyncedAt?: any;
  error?: string;
  action: SyncStatusAction;
}

interface IStatusProps {
  status?: Status;
  syncStatus: SyncStatus;
}

type IStatus = 'running' | 'error' | 'unknown' | 'syncing' | 'warning';

export const parseStatus = ({
  status,
  syncStatus,
}: IStatusProps): {
  status: IStatus;
} => {
  if (syncStatus.state === 'ERRORED_AT_AGENT') {
    return {
      status: 'error',
    };
  }

  if (
    syncStatus.state === 'APPLIED_AT_AGENT' ||
    syncStatus.state === 'IN_QUEUE'
  ) {
    return {
      status: 'syncing',
    };
  }

  if (status?.isReady && syncStatus.state === 'RECEIVED_UPDATE_FROM_AGENT') {
    return {
      status: 'running',
    };
  }

  return {
    status: 'unknown',
  };
};

export const ensureResource = <T>(v: T | undefined | null): T => {
  if (!v) {
    throw Error('Resource is not provided');
  }

  return v;
};
