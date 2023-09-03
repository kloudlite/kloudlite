export {
  type CloudProviderSecretIn,
  type NodePoolIn,
  type ClusterIn,
  type ProjectIn,
  type WorkspaceIn,
  type AppIn,
  type ManagedServiceIn,
  type ManagedResourceIn,
  type RouterIn,
  type SecretIn,
  type ConfigIn,
  type AccountIn,
} from '../gql/server';

export type ResourceFrom<T, Resource, SearchType> = T extends Resource
  ? SearchType
  : never;
