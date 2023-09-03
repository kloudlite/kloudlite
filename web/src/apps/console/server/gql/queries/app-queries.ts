import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  ConsoleListAppsQuery,
  ConsoleListAppsQueryVariables,
} from '~/root/src/generated/gql/server';

export const appQueries = (executor: IExecutor) => ({
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
      transformer: (data: ConsoleListAppsQuery) => data.core_listApps,
      vars(_: ConsoleListAppsQueryVariables) {},
    }
  ),
});
