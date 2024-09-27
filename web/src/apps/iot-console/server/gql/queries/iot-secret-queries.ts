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

export const iotSecretQueries = (executor: IExecutor) => ({
  listSecrets: executor(
    gql`
      query Core_listSecrets(
        $envName: String!
        $search: SearchSecrets
        $pq: CursorPaginationIn
      ) {
        core_listSecrets(
          envName: $envName
          search: $search
          pq: $pq
        ) {
          edges {
            cursor
            node {
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              stringData
              environmentName
              isReadyOnly
              immutable
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
              type
              updateTime
            }
          }
          pageInfo {
            endCursor
            hasNextPage
            hasPrevPage
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
      mutation Core_createSecret(
        $envName: String!
        $secret: SecretIn!
      ) {
        core_createSecret(
          envName: $envName
          secret: $secret
        ) {
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
        $envName: String!
        $name: String!
      ) {
        core_getSecret(
          envName: $envName
          name: $name
        ) {
          data
          displayName
          environmentName
          immutable
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
          stringData
          type
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
      mutation Core_updateSecret(
        $envName: String!
        $secret: SecretIn!
      ) {
        core_updateSecret(
          envName: $envName
          secret: $secret
        ) {
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
      mutation Core_deleteSecret(
        $envName: String!
        $secretName: String!
      ) {
        core_deleteSecret(
          envName: $envName
          secretName: $secretName
        )
      }
    `,
    {
      transformer: (data: ConsoleDeleteSecretMutation) => data,
      vars(_: ConsoleDeleteSecretMutationVariables) {},
    }
  ),
});
