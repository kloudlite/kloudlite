import {
  ResourceFrom,
  ClusterIn,
  ProjectIn,
  SecretIn,
  CloudProviderSecretIn,
} from '.';
import {
  ConsoleGetClusterQueryVariables,
  ConsoleGetProjectQueryVariables,
  ConsoleGetProviderSecretQueryVariables,
  ConsoleGetSecretQueryVariables,
} from '../gql/server';

export type getIn<Resource> =
  | ResourceFrom<Resource, ClusterIn, ConsoleGetClusterQueryVariables>
  | ResourceFrom<
      Resource,
      CloudProviderSecretIn,
      ConsoleGetProviderSecretQueryVariables
    >
  | ResourceFrom<Resource, ProjectIn, ConsoleGetProjectQueryVariables>
  | ResourceFrom<Resource, SecretIn, ConsoleGetSecretQueryVariables>;
