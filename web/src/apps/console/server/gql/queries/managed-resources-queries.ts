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
  // ConsoleImportManagedResourceMutation,
  // ConsoleImportManagedResourceMutationVariables,
  // ConsoleDeleteImportedManagedResourceMutation,
  // ConsoleDeleteImportedManagedResourceMutationVariables,
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
      query Core_getManagedResource(
        $name: String!
        $msvcName: String
        $envName: String
      ) {
        core_getManagedResource(
          name: $name
          msvcName: $msvcName
          envName: $envName
        ) {
          accountName
          apiVersion
          clusterName
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
          enabled
          environmentName
          id
          isImported
          kind
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          managedServiceName
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
          mresRef
          recordVersion
          spec {
            resourceNamePrefix
            resourceTemplate {
              apiVersion
              kind
              msvcRef {
                apiVersion
                clusterName
                kind
                name
                namespace
              }
            }
          }
          status {
            checkList {
              debug
              description
              hide
              name
              title
            }
            checks
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
            }
            apiVersion
            data
            immutable
            kind
            stringData
            type
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
        $msvcName: String!
        $mres: ManagedResourceIn!
      ) {
        core_createManagedResource(msvcName: $msvcName, mres: $mres) {
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
        $msvcName: String!
        $mres: ManagedResourceIn!
      ) {
        core_updateManagedResource(msvcName: $msvcName, mres: $mres) {
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
        $search: SearchManagedResources
        $pq: CursorPaginationIn
      ) {
        core_listManagedResources(search: $search, pq: $pq) {
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
              displayName
              enabled
              environmentName
              id
              isImported
              kind
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              managedServiceName
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
              mresRef
              recordVersion
              spec {
                resourceNamePrefix
                resourceTemplate {
                  apiVersion
                  kind
                  msvcRef {
                    apiVersion
                    clusterName
                    kind
                    name
                    namespace
                  }
                }
              }
              status {
                checkList {
                  debug
                  description
                  hide
                  name
                  title
                }
                checks
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
                }
                apiVersion
                data
                immutable
                kind
                stringData
                type
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
        $msvcName: String!
        $mresName: String!
      ) {
        core_deleteManagedResource(msvcName: $msvcName, mresName: $mresName)
      }
    `,
    {
      transformer: (data: ConsoleDeleteManagedResourceMutation) =>
        data.core_deleteManagedResource,
      vars(_: ConsoleDeleteManagedResourceMutationVariables) {},
    }
  ),
  // importManagedResource: executor(
  //   gql`
  //     mutation Core_importManagedResource(
  //       $envName: String!
  //       $mresName: String!
  //       $msvcName: String!
  //       $importName: String!
  //     ) {
  //       core_importManagedResource(
  //         envName: $envName
  //         mresName: $mresName
  //         msvcName: $msvcName
  //         importName: $importName
  //       ) {
  //         id
  //       }
  //     }
  //   `,
  //   {
  //     transformer: (data: ConsoleImportManagedResourceMutation) =>
  //       data.core_importManagedResource,
  //     vars(_: ConsoleImportManagedResourceMutationVariables) {},
  //   }
  // ),
  // deleteImportedManagedResource: executor(
  //   gql`
  //     mutation Core_deleteImportedManagedResource(
  //       $envName: String!
  //       $importName: String!
  //     ) {
  //       core_deleteImportedManagedResource(
  //         envName: $envName
  //         importName: $importName
  //       )
  //     }
  //   `,
  //   {
  //     transformer: (data: ConsoleDeleteImportedManagedResourceMutation) =>
  //       data.core_deleteImportedManagedResource,
  //     vars(_: ConsoleDeleteImportedManagedResourceMutationVariables) {},
  //   }
  // ),
});
