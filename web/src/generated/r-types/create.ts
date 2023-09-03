import {
  ResourceFrom,
  ClusterIn,
  ProjectIn,
  SecretIn,
  CloudProviderSecretIn,
} from '.';
import {
  ConsoleCreateClusterMutationVariables,
  ConsoleCreateProjectMutationVariables,
  ConsoleCreateProviderSecretMutationVariables,
  ConsoleCreateSecretMutationVariables,
} from '../gql/server';

export type createIn<Resource> =
  | ResourceFrom<Resource, ClusterIn, ConsoleCreateClusterMutationVariables>
  | ResourceFrom<
      Resource,
      CloudProviderSecretIn,
      ConsoleCreateProviderSecretMutationVariables
    >
  | ResourceFrom<Resource, ProjectIn, ConsoleCreateProjectMutationVariables>
  | ResourceFrom<Resource, SecretIn, ConsoleCreateSecretMutationVariables>;
