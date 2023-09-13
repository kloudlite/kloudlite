import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateSecretMutation,
  ConsoleCreateSecretMutationVariables,
  ConsoleGetSecretQuery,
  ConsoleGetSecretQueryVariables,
  ConsoleListSecretsQuery,
  ConsoleListSecretsQueryVariables,
  ConsoleUpdateSecretMutation,
  ConsoleUpdateSecretMutationVariables,
} from '~/root/src/generated/gql/server';

export type ISecret = NN<ConsoleGetSecretQuery['core_getSecret']>;

export const secretQueries = (executor: IExecutor) => ({
  listSecrets: executor(
    gql`
      query Core_listSecrets(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $pq: CursorPaginationIn
        $search: SearchSecrets
      ) {
        core_listSecrets(
          project: $project
          scope: $scope
          pq: $pq
          search: $search
        ) {
          pageInfo {
            startCursor
            hasPreviousPage
            hasNextPage
            endCursor
          }
          totalCount
          edges {
            node {
              stringData
              updateTime
              displayName
              metadata {
                name
                namespace
                annotations
                labels
              }
            }
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleListSecretsQuery) => data.core_listSecrets,
      vars(_: ConsoleListSecretsQueryVariables) {},
    }
  ),
  createSecret: executor(
    gql`
      mutation createSecret($secret: SecretIn!) {
        core_createSecret(secret: $secret) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateSecretMutation) =>
        data.core_createSecret,
      vars(_: ConsoleCreateSecretMutationVariables) {},
    }
  ),

  getSecret: executor(
    gql`
      query Core_getSecret(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $name: String!
      ) {
        core_getSecret(project: $project, scope: $scope, name: $name) {
          stringData
          updateTime
          displayName
          metadata {
            name
            namespace
            annotations
            labels
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleGetSecretQuery) => data.core_getSecret,
      vars(_: ConsoleGetSecretQueryVariables) {},
    }
  ),
  updateSecret: executor(
    gql`
      mutation Core_updateSecret($secret: SecretIn!) {
        core_updateSecret(secret: $secret) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateSecretMutation) => data,
      vars(_: ConsoleUpdateSecretMutationVariables) {},
    }
  ),
});
