import {
  Account,
  App,
  CloudProviderSecret,
  Cluster,
  Config,
  CursorPaginationIn,
  ManagedResource,
  ManagedService,
  NodePool,
  PageInfo,
  Project,
  Router,
  Scalars,
  SearchCluster,
  SearchProjects,
  SearchProviderSecret,
  SearchSecrets,
  Secret,
  Workspace,
} from '../gql/server';

export {
  type CloudProviderSecret,
  type CloudProviderSecretIn,
  type NodePool,
  type Cluster,
  type ClusterIn,
  type Project,
  type Workspace,
  type App,
  type ManagedService,
  type ManagedResource,
  type Router,
  type Secret,
  type Config,
  type Account,
  type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode as AvailabiltyMode,
  type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider as CloudProvider,
  type User,
} from '../gql/server';

export interface EdgeOut<T> {
  cursor: Scalars['String']['output'];
  node: T;
}

export interface PaginatedOut<T> {
  edges: Array<EdgeOut<T>>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
}

export type AnyKResource =
  | CloudProviderSecret
  | NodePool
  | Cluster
  | Project
  | Workspace
  | App
  | ManagedService
  | ManagedResource
  | Router
  | Secret
  | Config
  | Account;

type search<T, Resource, SearchType> = T extends Resource ? SearchType : never;

export type SearchIn<Resource> =
  | search<Resource, Cluster, SearchCluster>
  | search<Resource, CloudProviderSecret, SearchProviderSecret>
  | search<Resource, Project, SearchProjects>
  | search<Resource, Secret, SearchSecrets>;

export interface PaginatedArgs<T> {
  pagination?: CursorPaginationIn;
  search?: SearchIn<T>;
}
