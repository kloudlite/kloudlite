import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleImportManagedResourceMutation,
  ConsoleImportManagedResourceMutationVariables,
  ConsoleDeleteImportedManagedResourceMutation,
  ConsoleDeleteImportedManagedResourceMutationVariables,
} from '~/root/src/generated/gql/server';

// export type IManagedResource = NN<
//   ConsoleGetManagedResourceQuery['core_getManagedResource']
// >;
// export type IManagedResources = NN<
//   ConsoleListManagedResourcesQuery['core_listManagedResources']
// >;

export const importedManagedResourceQueries = (executor: IExecutor) => ({
  importManagedResource: executor(
    gql`
      mutation Core_importManagedResource(
        $envName: String!
        $mresName: String!
        $msvcName: String!
        $importName: String!
      ) {
        core_importManagedResource(
          envName: $envName
          mresName: $mresName
          msvcName: $msvcName
          importName: $importName
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleImportManagedResourceMutation) =>
        data.core_importManagedResource,
      vars(_: ConsoleImportManagedResourceMutationVariables) {},
    }
  ),
  deleteImportedManagedResource: executor(
    gql`
      mutation Core_deleteImportedManagedResource(
        $envName: String!
        $importName: String!
      ) {
        core_deleteImportedManagedResource(
          envName: $envName
          importName: $importName
        )
      }
    `,
    {
      transformer: (data: ConsoleDeleteImportedManagedResourceMutation) =>
        data.core_deleteImportedManagedResource,
      vars(_: ConsoleDeleteImportedManagedResourceMutationVariables) {},
    }
  ),
  //   listImportedManagedResources: executor(
  //     gql`
  //       query Core_listManagedResources(
  //         $search: SearchManagedResources
  //         $pq: CursorPaginationIn
  //       ) {
  //         core_listManagedResources(search: $search, pq: $pq) {
  //           edges {
  //             cursor
  //             node {
  //               accountName
  //               apiVersion
  //               clusterName
  //               createdBy {
  //                 userEmail
  //                 userId
  //                 userName
  //               }
  //               creationTime
  //               displayName
  //               enabled
  //               environmentName
  //               id
  //               isImported
  //               kind
  //               lastUpdatedBy {
  //                 userEmail
  //                 userId
  //                 userName
  //               }
  //               managedServiceName
  //               markedForDeletion
  //               metadata {
  //                 annotations
  //                 creationTimestamp
  //                 deletionTimestamp
  //                 generation
  //                 labels
  //                 name
  //                 namespace
  //               }
  //               mresRef
  //               recordVersion
  //               spec {
  //                 resourceNamePrefix
  //                 resourceTemplate {
  //                   apiVersion
  //                   kind
  //                   msvcRef {
  //                     apiVersion
  //                     clusterName
  //                     kind
  //                     name
  //                     namespace
  //                   }
  //                 }
  //               }
  //               status {
  //                 checkList {
  //                   debug
  //                   description
  //                   hide
  //                   name
  //                   title
  //                 }
  //                 checks
  //                 isReady
  //                 lastReadyGeneration
  //                 lastReconcileTime
  //                 message {
  //                   RawMessage
  //                 }
  //                 resources {
  //                   apiVersion
  //                   kind
  //                   name
  //                   namespace
  //                 }
  //               }
  //               syncedOutputSecretRef {
  //                 metadata {
  //                   name
  //                 }
  //                 apiVersion
  //                 data
  //                 immutable
  //                 kind
  //                 stringData
  //                 type
  //               }
  //               syncStatus {
  //                 action
  //                 error
  //                 lastSyncedAt
  //                 recordVersion
  //                 state
  //                 syncScheduledAt
  //               }
  //               updateTime
  //             }
  //           }
  //           pageInfo {
  //             endCursor
  //             hasNextPage
  //             hasPreviousPage
  //             startCursor
  //           }
  //           totalCount
  //         }
  //       }
  //     `,
  //     {
  //       transformer: (data: ConsoleListImportedManagedResourcesQuery) =>
  //         data.core_listManagedResources,
  //       vars(_: ConsoleListImportedManagedResourcesQueryVariables) {},
  //     }
  //   ),
});
