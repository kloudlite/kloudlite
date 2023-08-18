import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const secretQueries = (executor = ExecuteQueryWithContext({})) => ({
  listSecrets: executor(
    gql`
      query Core_listConfigs(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $search: SearchConfigs
        $pq: CursorPaginationIn
      ) {
        core_listConfigs(
          project: $project
          scope: $scope
          search: $search
          pq: $pq
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
});
