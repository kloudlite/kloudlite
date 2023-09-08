import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  ConsoleListRoutersQuery,
  ConsoleListRoutersQueryVariables,
} from '~/root/src/generated/gql/server';

export const routerQueries = (executor: IExecutor) => ({
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
            cursor
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
      transformer: (data: ConsoleListRoutersQuery) => data.core_listRouters,
      vars(_: ConsoleListRoutersQueryVariables) {},
    }
  ),
});
