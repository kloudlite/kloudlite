import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateManagedResourceMutation,
  ConsoleCreateManagedResourceMutationVariables,
  ConsoleGetManagedResourceQuery,
  ConsoleGetManagedResourceQueryVariables,
  ConsoleListManagedResourceQuery,
  ConsoleListManagedResourceQueryVariables,
} from '~/root/src/generated/gql/server';

export type IManagedResources = NN<
  ConsoleListManagedResourceQuery['core_listManagedResources']
>;
export const managedResourceQueries = (executor: IExecutor) => ({
  getManagedResource: executor(
    gql`
      query Core_getManagedResource(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $name: String!
      ) {
        core_getManagedResource(project: $project, scope: $scope, name: $name) {
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
          spec {
            resourceTemplate {
              apiVersion
              kind
              msvcRef {
                name
                kind
                apiVersion
              }
              spec
            }
          }
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
  listManagedResource: executor(
    gql`
      query Core_listManagedResources(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $search: SearchManagedResources
        $pq: CursorPaginationIn
      ) {
        core_listManagedResources(
          project: $project
          scope: $scope
          search: $search
          pq: $pq
        ) {
          pageInfo {
            endCursor
            hasNextPage
            hasPreviousPage
            startCursor
          }
          totalCount
          edges {
            cursor
            node {
              updateTime
              metadata {
                name
              }
              lastUpdatedBy {
                userEmail
                userName
              }
              kind
              displayName
              creationTime
              createdBy {
                userEmail
                userName
              }
            }
          }
        }
      }
    `,
    {
      transformer(data: ConsoleListManagedResourceQuery) {
        return data.core_listManagedResources;
      },
      vars(_: ConsoleListManagedResourceQueryVariables) {},
    }
  ),
  createManagedResource: executor(
    gql`
      mutation Core_createManagedResource($mres: ManagedResourceIn!) {
        core_createManagedResource(mres: $mres) {
          id
        }
      }
    `,
    {
      transformer(data: ConsoleCreateManagedResourceMutation) {
        return data.core_createManagedResource;
      },
      vars(_: ConsoleCreateManagedResourceMutationVariables) {},
    }
  ),
});
