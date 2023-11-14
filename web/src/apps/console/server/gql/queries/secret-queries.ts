import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateSecretMutation,
  ConsoleCreateSecretMutationVariables,
  ConsoleDeleteSecretMutation,
  ConsoleDeleteSecretMutationVariables,
  ConsoleGetSecretQuery,
  ConsoleGetSecretQueryVariables,
  ConsoleListSecretsQuery,
  ConsoleListSecretsQueryVariables,
  ConsoleUpdateSecretMutation,
  ConsoleUpdateSecretMutationVariables,
} from '~/root/src/generated/gql/server';

export type ISecret = NN<ConsoleGetSecretQuery['core_getSecret']>;
export type ISecrets = NN<ConsoleListSecretsQuery['core_listSecrets']>;

export const secretQueries = (executor: IExecutor) => ({
  listSecrets: executor(
    gql`
      query Core_listSecrets(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $search: SearchSecrets
        $pq: CursorPaginationIn
      ) {
        core_listSecrets(
          project: $project
          scope: $scope
          search: $search
          pq: $pq
        ) {
          edges {
            cursor
            node {
              accountName
              apiVersion
              clusterName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              data
              displayName
              enabled
              id
              kind
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              metadata {
                annotations
                creationTimestamp
                deletionTimestamp
                generation
                labels
                name
                namespace
              }
              recordVersion
              status {
                checks
                isReady
                lastReconcileTime
                message {
                  RawMessage
                }
                resources {
                  apiVersion
                  kind
                  name
                  namespace
                }
              }
              stringData
              syncStatus {
                action
                error
                lastSyncedAt
                recordVersion
                state
                syncScheduledAt
              }
              type
              updateTime
            }
          }
          pageInfo {
            endCursor
            hasNextPage
            hasPreviousPage
            startCursor
          }
          totalCount
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
  deleteSecret: executor(
    gql`
      mutation Core_deleteSecret($namespace: String!, $name: String!) {
        core_deleteSecret(namespace: $namespace, name: $name)
      }
    `,
    {
      transformer: (data: ConsoleDeleteSecretMutation) => data,
      vars(_: ConsoleDeleteSecretMutationVariables) {},
    }
  ),
});
