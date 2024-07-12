import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleImportManagedResourceMutation,
  ConsoleImportManagedResourceMutationVariables,
  ConsoleDeleteImportedManagedResourceMutation,
  ConsoleDeleteImportedManagedResourceMutationVariables,
  ConsoleListImportedManagedResourcesQuery,
  ConsoleListImportedManagedResourcesQueryVariables,
} from '~/root/src/generated/gql/server';

export type IImportedManagedResources = NN<
  ConsoleListImportedManagedResourcesQuery['core_listImportedManagedResources']
>;

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
  listImportedManagedResources: executor(
    gql`
      query Core_listImportedManagedResources(
        $envName: String!
        $search: SearchImportedManagedResources
        $pq: CursorPaginationIn
      ) {
        core_listImportedManagedResources(
          envName: $envName
          search: $search
          pq: $pq
        ) {
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
              environmentName
              id
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              managedResourceRef {
                id
                name
                namespace
              }
              markedForDeletion
              name
              recordVersion
              secretRef {
                name
                namespace
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
              managedResource {
                accountName
                apiVersion
                clusterName
                creationTime
                displayName
                enabled
                environmentName
                id
                isImported
                kind
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
                      kind
                      name
                      namespace
                    }
                    spec
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
                  apiVersion
                  data
                  immutable
                  kind
                  stringData
                  type
                }
                updateTime
              }
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
      transformer: (data: ConsoleListImportedManagedResourcesQuery) =>
        data.core_listImportedManagedResources,
      vars(_: ConsoleListImportedManagedResourcesQueryVariables) {},
    }
  ),
});
