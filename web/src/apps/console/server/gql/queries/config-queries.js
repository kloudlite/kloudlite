import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const configQueries = (executor = ExecuteQueryWithContext({})) => ({
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
