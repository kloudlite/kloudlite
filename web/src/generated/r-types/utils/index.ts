import {
  Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode as AvailabilityMode,
  Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider as CloudProvider,
} from '../../gql/server';

interface IparseNodes<T> {
  edges: Array<{ cursor: string; node: T }>;
}

export const parseNodes = <T>(resources: IparseNodes<T>): T[] =>
  resources.edges.map(({ node }) => node);

interface IparseName {
  metadata: {
    name: string;
  };
}

export const parseName = (resource: IparseName) => resource.metadata.name;

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
