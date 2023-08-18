import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const appQueries = (executor = ExecuteQueryWithContext({})) => ({
  listApps: executor(
    gql`
      query Core_listApps(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $search: SearchApps
        $pagination: CursorPaginationIn
      ) {
        core_listApps(
          project: $project
          scope: $scope
          search: $search
          pq: $pagination
        ) {
          edges {
            node {
              spec {
                displayName
              }
              metadata {
                namespace
                name
                labels
                annotations
              }
            }
          }
        }
      }
    `,
    {
      dataPath: 'core_listApps',
    }
  ),
});
