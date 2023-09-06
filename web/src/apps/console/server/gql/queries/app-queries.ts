import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  ConsoleListAppsQuery,
  ConsoleListAppsQueryVariables,
  ConsoleCreateAppMutation,
  ConsoleCreateAppMutationVariables,
} from '~/root/src/generated/gql/server';

export const appQueries = (executor: IExecutor) => ({
  createApp: executor(
    gql`
      mutation Core_createApp($app: AppIn!) {
        core_createApp(app: $app) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateAppMutation) => data.core_createApp,
      vars(_: ConsoleCreateAppMutationVariables) {},
    }
  ),
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
          totalCount
          pageInfo {
            startCursor
            hasPreviousPage
            hasNextPage
            endCursor
          }

          edges {
            cursor
            node {
              spec {
                displayName
              }
              clusterName
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
      transformer: (data: ConsoleListAppsQuery) => data.core_listApps,
      vars(_: ConsoleListAppsQueryVariables) {},
    }
  ),
});
