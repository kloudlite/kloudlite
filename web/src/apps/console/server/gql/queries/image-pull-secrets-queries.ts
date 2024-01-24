import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
    ConsoleCreateImagePullSecretMutation,
    ConsoleCreateImagePullSecretMutationVariables,
    ConsoleListImagePullSecretsQuery,
    ConsoleListImagePullSecretsQueryVariables,
} from '~/root/src/generated/gql/server';

export type IImagePullSecrets = NN<
    ConsoleListImagePullSecretsQuery['core_listImagePullSecrets']
>;

export const imagePullSecretsQueries = (executor: IExecutor) => ({
    createImagePullSecret: executor(
        gql`
      mutation Core_createImagePullSecret(
        $projectName: String!
        $envName: String!
        $imagePullSecretIn: ImagePullSecretIn!
      ) {
        core_createImagePullSecret(
          projectName: $projectName
          envName: $envName
          imagePullSecretIn: $imagePullSecretIn
        ) {
          id
        }
      }
    `,
        {
            transformer: (data: ConsoleCreateImagePullSecretMutation) =>
                data.core_createImagePullSecret,
            vars(_: ConsoleCreateImagePullSecretMutationVariables) { },
        }
    ),
    listImagePullSecrets: executor(
        gql`
      query Core_listImagePullSecrets(
        $projectName: String!
        $envName: String!
        $search: SearchImagePullSecrets
        $pq: CursorPaginationIn
      ) {
        core_listImagePullSecrets(
          projectName: $projectName
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
              dockerConfigJson
              environmentName
              format
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              metadata {
                generation
                name
                namespace
              }
              projectName
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
            vars(_: ConsoleListImagePullSecretsQueryVariables) { },
        }
    ),
});
