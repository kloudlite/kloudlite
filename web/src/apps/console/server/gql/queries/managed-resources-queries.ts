import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleGetManagedResourceQuery,
  ConsoleGetManagedResourceQueryVariables,
  ConsoleListManagedResourcesQuery,
  ConsoleListManagedResourcesQueryVariables,
  ConsoleCreateManagedResourceMutation,
  ConsoleCreateManagedResourceMutationVariables,
  ConsoleUpdateManagedResourceMutation,
  ConsoleUpdateManagedResourceMutationVariables,
  ConsoleDeleteManagedResourceMutation,
  ConsoleDeleteManagedResourceMutationVariables,
} from '~/root/src/generated/gql/server';

export type IManagedResource = NN<
  ConsoleGetManagedResourceQuery['core_getManagedResource']
>;
export type IManagedResources = NN<
  ConsoleListManagedResourcesQuery['core_listManagedResources']
>;

export const managedResourceQueries = (executor: IExecutor) => ({
  getManagedResource: executor(
    gql`
      query Core_getManagedResource($envName: String!, $name: String!) {
        core_getManagedResource(envName: $envName, name: $name) {
          displayName
          enabled
          environmentName
          markedForDeletion
          metadata {
            name
            namespace
          }
          spec {
            resourceTemplate {
              apiVersion
              kind
              msvcRef {
                apiVersion
                kind
                name
                namespace
              }
              spec
            }
          }
          updateTime
        }
      }
    `,
    {
      transformer(data: ConsoleGetManagedResourceQuery) {
        return data.core_getManagedResource;
      },
      vars(_: ConsoleGetManagedResourceQueryVariables) {},
    }
  ),
  createManagedResource: executor(
    gql`
      mutation Core_createManagedResource(
        $envName: String!
        $mres: ManagedResourceIn!
      ) {
        core_createManagedResource(envName: $envName, mres: $mres) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateManagedResourceMutation) =>
        data.core_createManagedResource,
      vars(_: ConsoleCreateManagedResourceMutationVariables) {},
    }
  ),
  updateManagedResource: executor(
    gql`
      mutation Core_updateManagedResource(
        $envName: String!
        $mres: ManagedResourceIn!
      ) {
        core_updateManagedResource(envName: $envName, mres: $mres) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateManagedResourceMutation) =>
        data.core_updateManagedResource,
      vars(_: ConsoleUpdateManagedResourceMutationVariables) {},
    }
  ),
  listManagedResources: executor(
    gql`
      query Core_listManagedResources(
        $envName: String!
        $search: SearchManagedResources
        $pq: CursorPaginationIn
      ) {
        core_listManagedResources(envName: $envName, search: $search, pq: $pq) {
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
              spec {
                resourceTemplate {
                  apiVersion
                  kind
                  msvcRef {
                    apiVersion
                    kind
                    name
                    namespace
                    clusterName
                  }
                  spec
                }
              }
              status {
                checks
                checkList {
                  description
                  debug
                  name
                  title
                }
                isReady
                lastReadyGeneration
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
              syncedOutputSecretRef {
                metadata {
                  name
                  namespace
                }
              }
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
      transformer: (data: ConsoleListManagedResourcesQuery) =>
        data.core_listManagedResources,
      vars(_: ConsoleListManagedResourcesQueryVariables) {},
    }
  ),
  deleteManagedResource: executor(
    gql`
      mutation Core_deleteManagedResource(
        $envName: String!
        $mresName: String!
      ) {
        core_deleteManagedResource(envName: $envName, mresName: $mresName)
      }
    `,
    {
      transformer: (data: ConsoleDeleteManagedResourceMutation) =>
        data.core_deleteManagedResource,
      vars(_: ConsoleDeleteManagedResourceMutationVariables) {},
    }
  ),
});
