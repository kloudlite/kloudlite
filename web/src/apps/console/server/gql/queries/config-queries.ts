import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';

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
      transformer(data) {},
      vars(variables) {},
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
      transformer(data) {},
      vars(variables) {},
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
      transformer(data) {},
      vars(variables) {},
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
      transformer(data) {},
      vars(variables) {},
    }
  ),
});
