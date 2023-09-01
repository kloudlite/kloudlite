import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { IGqlReturn } from '~/root/lib/types/common';

export interface IGQLMethodsConfig {
  updateConfig: (variables?: any) => IGqlReturn<any>;
  getConfig: (variables?: any) => IGqlReturn<any>;
  listConfigs: (variables?: any) => IGqlReturn<any>;
  createConfig: (variables?: any) => IGqlReturn<any>;
}

export const configQueries = (executor: IExecutor): IGQLMethodsConfig => ({
  updateConfig: executor(
    gql`
      mutation Mutation($config: ConfigIn!) {
        core_updateConfig(config: $config) {
          id
        }
      }
    `
  ),
  getConfig: executor(
    gql`
      query Core_getConfig(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $name: String!
      ) {
        core_getConfig(project: $project, scope: $scope, name: $name) {
          data
          updateTime
          displayName
          metadata {
            name
            namespace
            annotations
            labels
          }
        }
      }
    `,
    {
      dataPath: 'core_getConfig',
    }
  ),
  listConfigs: executor(
    gql`
      query Core_listConfigs(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $search: SearchConfigs
        $pagination: CursorPaginationIn
      ) {
        core_listConfigs(
          project: $project
          scope: $scope
          search: $search
          pq: $pagination
        ) {
          pageInfo {
            startCursor
            hasPreviousPage
            hasNextPage
            endCursor
          }
          totalCount
          edges {
            node {
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
        }
      }
    `,
    {
      dataPath: 'core_listConfigs',
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
    { dataPath: 'core_createConfig' }
  ),
});
