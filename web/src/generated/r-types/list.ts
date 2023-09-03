import {
  CloudProviderSecretIn,
  ClusterIn,
  ProjectIn,
  ResourceFrom,
  SecretIn,
} from '.';
import {
  ConsoleListClustersQuery,
  ConsoleListClustersQueryVariables,
  ConsoleListProjectsQuery,
  ConsoleListProjectsQueryVariables,
  ConsoleListProviderSecretsQuery,
  ConsoleListProviderSecretsQueryVariables,
  ConsoleListSecretsQuery,
  ConsoleListSecretsQueryVariables,
} from '../gql/server';

export type listIn<Resource> =
  | ResourceFrom<Resource, ClusterIn, ConsoleListClustersQueryVariables>
  | ResourceFrom<
      Resource,
      CloudProviderSecretIn,
      ConsoleListProviderSecretsQueryVariables
    >
  | ResourceFrom<Resource, ProjectIn, ConsoleListProjectsQueryVariables>
  | ResourceFrom<Resource, SecretIn, ConsoleListSecretsQueryVariables>;

export type listOut<Resource> =
  | ResourceFrom<Resource, ClusterIn, ConsoleListClustersQuery>
  | ResourceFrom<
      Resource,
      CloudProviderSecretIn,
      ConsoleListProviderSecretsQuery
    >
  | ResourceFrom<Resource, ProjectIn, ConsoleListProjectsQuery>
  | ResourceFrom<Resource, SecretIn, ConsoleListSecretsQuery>;
