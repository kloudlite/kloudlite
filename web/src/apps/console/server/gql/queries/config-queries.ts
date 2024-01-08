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
      mutation Core_updateConfig(
        $projectName: String!
        $envName: String!
        $config: ConfigIn!
      ) {
        core_updateConfig(
          projectName: $projectName
          envName: $envName
          config: $config
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateConfigMutation) => data,
      vars(_: ConsoleUpdateConfigMutationVariables) { },
    }
  ),
  getConfig: executor(
    gql`
      query Core_getConfig(
        $projectName: String!
        $envName: String!
        $name: String!
      ) {
        core_getConfig(
          projectName: $projectName
          envName: $envName
          name: $name
        ) {
          binaryData
          data
          displayName
          environmentName
          immutable
          metadata {
            annotations
            creationTimestamp
            deletionTimestamp
            generation
            labels
            name
            namespace
          }
          projectName
        }
      }
    `,
    {
      transformer: (data: ConsoleGetConfigQuery) => data.core_getConfig,
      vars(_: ConsoleGetConfigQueryVariables) { },
    }
  ),
  listConfigs: executor(
    gql`
      query Core_listConfigs(
        $projectName: String!
        $envName: String!
        $search: SearchConfigs
        $pq: CursorPaginationIn
      ) {
        core_listConfigs(
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
              data
              environmentName
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
              projectName
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
      vars(_: ConsoleListConfigsQueryVariables) { },
    }
  ),
  createConfig: executor(
    gql`
      mutation Core_createConfig(
        $projectName: String!
        $envName: String!
        $config: ConfigIn!
      ) {
        core_createConfig(
          projectName: $projectName
          envName: $envName
          config: $config
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateConfigMutation) =>
        data.core_createConfig,
      vars(_: ConsoleCreateConfigMutationVariables) { },
    }
  ),
});
