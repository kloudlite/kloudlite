import {
  ClusterIn,
  ResourceFrom,
  ProjectIn,
  SecretIn,
  CloudProviderSecretIn,
} from '.';

import {
  ConsoleUpdateConfigMutationVariables,
  ConsoleUpdateProviderSecretMutationVariables,
} from '../gql/server';

export type updateIn<Resource> =
  | ResourceFrom<Resource, ClusterIn, ConsoleUpdateConfigMutationVariables>
  | ResourceFrom<
      Resource,
      CloudProviderSecretIn,
      ConsoleUpdateProviderSecretMutationVariables
    >
  | ResourceFrom<
      Resource,
      ProjectIn,
      ConsoleUpdateProviderSecretMutationVariables
    >
  | ResourceFrom<Resource, SecretIn, ConsoleUpdateConfigMutationVariables>;
