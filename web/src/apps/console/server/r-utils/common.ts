/* eslint-disable no-prototype-builtins */
import { Params } from '@remix-run/react';
import { dayjs } from '~/components/molecule/dayjs';
import { FlatMapType, NonNullableString } from '~/root/lib/types/common';
import {
  Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode as AvailabilityMode,
  Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider as CloudProvider,
  Kloudlite_Io__Pkg__Types_SyncStatusAction as SyncStatusAction,
  Kloudlite_Io__Pkg__Types_SyncStatusState as SyncStatusState,
  Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigProvisionMode as ProvisionMode,
  ProjectId,
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
      metadata: {
        name: string;
      };
    }
  | undefined
  | null;

export const parseName = (resource: IparseName) => {
  // if (!resource) {
  //   throw Error('resource not found');
  // }
  return resource?.metadata.name || '';
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

type parseFromAnnResource =
  | {
      metadata: {
        annotations?: FlatMapType<string>;
      };
    }
  | undefined
  | null;

export const parseFromAnn = (resource: parseFromAnnResource, key: string) =>
  resource?.metadata?.annotations?.[key] || '';

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

export const validateProvisionMode = (v: string): ProvisionMode => {
  switch (v as ProvisionMode) {
    case 'on_demand':
    case 'reserved':
    case 'spot':
      return v as ProvisionMode;
    default:
      throw Error(`invalid provision mode type ${v}`);
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

export type listOrGrid = 'list' | 'grid' | NonNullableString;
export type wsOrEnv = 'environment' | 'workspace' | NonNullableString;

interface IStatus {
  syncStatus: {
    syncScheduledAt?: any;
    state: SyncStatusState;
    recordVersion: number;
    lastSyncedAt?: any;
    error?: string;
    action: SyncStatusAction;
  };
  status?: {
    lastReconcileTime?: any;
    isReady: boolean;
    checks?: any;
    resources?: Array<{
      namespace: string;
      name: string;
      kind?: string;
      apiVersion?: string;
    }>;
    message?: { RawMessage?: any };
  };
}

export const parseStatus = (_: IStatus) => {
  return 'status';
};

export const parseUpdateTime = (resource: { updateTime: string }) => {
  return dayjs(resource.updateTime).fromNow();
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
