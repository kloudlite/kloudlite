import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateConfigMutation,
  ConsoleCreateConfigMutationVariables,
  ConsoleGetConfigQuery,
  ConsoleGetConfigQueryVariables,
  ConsoleListConfigsQuery,
  ConsoleListConfigsQueryVariables,
  ConsoleUpdateConfigMutation,
  ConsoleUpdateConfigMutationVariables,
} from '~/root/src/generated/gql/server';

export type IConfig = NN<ConsoleGetConfigQuery['core_getConfig']>;
export type IConfigs = NN<ConsoleListConfigsQuery['core_listConfigs']>;

export const configQueries = (executor: IExecutor) => ({
  updateConfig: executor(
    gql`
      mutation updateConfig($config: ConfigIn!) {
        core_updateConfig(config: $config) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateConfigMutation) => data,
      vars(_: ConsoleUpdateConfigMutationVariables) {},
    }
  ),
  getConfig: executor(
    gql`
      query Core_getConfig(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $name: String!
      ) {
        core_getConfig(project: $project, scope: $scope, name: $name) {
          metadata {
            namespace
            name
            annotations
            labels
          }
          displayName
          updateTime
          data
        }
      }
    `,
    {
      transformer: (data: ConsoleGetConfigQuery) => data.core_getConfig,
      vars(_: ConsoleGetConfigQueryVariables) {},
    }
  ),
  listConfigs: executor(
    gql`
      query Core_listConfigs(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $pq: CursorPaginationIn
        $search: SearchConfigs
      ) {
        core_listConfigs(
          project: $project
          scope: $scope
          pq: $pq
          search: $search
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
      transformer: (data: ConsoleListConfigsQuery) => data.core_listConfigs,
      vars(_: ConsoleListConfigsQueryVariables) {},
    }
  ),
  createConfig: executor(
    gql`
      mutation Core_createConfig($config: ConfigIn!) {
        core_createConfig(config: $config) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateConfigMutation) =>
        data.core_createConfig,
      vars(_: ConsoleCreateConfigMutationVariables) {},
    }
  ),
});
