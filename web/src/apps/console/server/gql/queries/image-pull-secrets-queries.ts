import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateImagePullSecretMutation,
  ConsoleCreateImagePullSecretMutationVariables,
  ConsoleUpdateImagePullSecretMutation,
  ConsoleUpdateImagePullSecretMutationVariables,
  ConsoleListImagePullSecretsQuery,
  ConsoleListImagePullSecretsQueryVariables,
  ConsoleDeleteImagePullSecretsMutation,
  ConsoleDeleteImagePullSecretsMutationVariables,
} from '~/root/src/generated/gql/server';

export type IImagePullSecrets = NN<
  ConsoleListImagePullSecretsQuery['core_listImagePullSecrets']
>;

export const imagePullSecretsQueries = (executor: IExecutor) => ({
  createImagePullSecret: executor(
    gql`
      mutation Core_createImagePullSecret($pullSecret: ImagePullSecretIn!) {
        core_createImagePullSecret(pullSecret: $pullSecret) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateImagePullSecretMutation) =>
        data.core_createImagePullSecret,
      vars(_: ConsoleCreateImagePullSecretMutationVariables) {},
    }
  ),
  updateImagePullSecret: executor(
    gql`
      mutation Core_updateImagePullSecret($pullSecret: ImagePullSecretIn!) {
        core_updateImagePullSecret(pullSecret: $pullSecret) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateImagePullSecretMutation) =>
        data.core_updateImagePullSecret,
      vars(_: ConsoleUpdateImagePullSecretMutationVariables) {},
    }
  ),
  deleteImagePullSecrets: executor(
    gql`
      mutation Core_deleteImagePullSecret($name: String!) {
        core_deleteImagePullSecret(name: $name)
      }
    `,
    {
      transformer: (data: ConsoleDeleteImagePullSecretsMutation) =>
        data.core_deleteImagePullSecret,
      vars(_: ConsoleDeleteImagePullSecretsMutationVariables) {},
    }
  ),
  listImagePullSecrets: executor(
    gql`
      query Core_listImagePullSecrets(
        $search: SearchImagePullSecrets
        $pq: CursorPaginationIn
      ) {
        core_listImagePullSecrets(search: $search, pq: $pq) {
          edges {
            cursor
            node {
              accountName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              dockerConfigJson
              format
              id
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
              registryPassword
              registryURL
              registryUsername
              syncStatus {
                action
                error
                lastSyncedAt
                recordVersion
                state
                syncScheduledAt
              }
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
      transformer: (data: ConsoleListImagePullSecretsQuery) =>
        data.core_listImagePullSecrets,
      vars(_: ConsoleListImagePullSecretsQueryVariables) {},
    }
  ),
});
