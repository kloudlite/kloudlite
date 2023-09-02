import { AnyKResource, AvailabiltyMode, CloudProvider, PaginatedOut } from '..';

export const parseNodes = <T>(resources: PaginatedOut<T>): T[] =>
  resources.edges.map(({ node }) => node);

export const parseName = (resource: AnyKResource) => resource.metadata.name;

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

export const validateAvailabilityMode = (v: string): AvailabiltyMode => {
  switch (v as AvailabiltyMode) {
    case 'HA':
    case 'dev':
      return v as AvailabiltyMode;
    default:
      throw Error(`invalid availabilityMode ${v}`);
  }
};
