import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const routerQueries = (executor = ExecuteQueryWithContext({})) => ({
  listRouters: executor(
    gql`
      query Core_listRouters(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $search: SearchRouters
        $pq: CursorPaginationIn
      ) {
        core_listRouters(
          project: $project
          scope: $scope
          search: $search
          pq: $pq
        ) {
          edges {
            node {
              metadata {
                name
                namespace
                annotations
                labels
              }
              spec {
                routes {
                  app
                  lambda
                  path
                }
              }
            }
          }
        }
      }
    `,
    {
      dataPath: 'core_listRouters',
    }
  ),
});
