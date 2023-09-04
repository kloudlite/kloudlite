import { DeepReadOnly, FlatMapType } from '~/root/lib/types/common';
import {
  Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode as AvailabilityMode,
  Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider as CloudProvider,
} from '~/root/src/generated/gql/server';

type IparseNodes<T> = {
  edges: Array<{ cursor: string; node: T }>;
};

export const parseNodes = <T>(resources: IparseNodes<T>): T[] =>
  resources.edges.map(({ node }) => node);

type IparseName =
  | {
      metadata: {
        name: string;
      };
    }
  | undefined
  | null;

export const parseName = (resource: IparseName) => {
  if (!resource) {
    throw Error('resource not found');
  }
  return resource.metadata.name || '';
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
        namespace: string;
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

  return resource.spec.namespace;
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

export const validateAvailabilityMode = (v: string): AvailabilityMode => {
  switch (v as AvailabilityMode) {
    case 'HA':
    case 'dev':
      return v as AvailabilityMode;
    default:
      throw Error(`invalid availabilityMode ${v}`);
  }
};
