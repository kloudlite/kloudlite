import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const workspaceQueries = (executor = ExecuteQueryWithContext({})) => ({
  createWorkspace: executor(
    gql`
      mutation Core_createWorkspace($env: WorkspaceIn!) {
        core_createWorkspace(env: $env) {
          id
        }
      }
    `,
    {
      dataPath: 'core_createWorkspace',
    }
  ),
  listWorkspaces: executor(
    gql`
      query Core_listWorkspaces(
        $namespace: String!
        $search: SearchWorkspaces
        $pq: CursorPaginationIn
      ) {
        core_listWorkspaces(namespace: $namespace, search: $search, pq: $pq) {
          totalCount
          pageInfo {
            hasPreviousPage
            startCursor
            hasNextPage
            endCursor
          }
          edges {
            node {
              updateTime
              metadata {
                namespace
                name
                annotations
                labels
              }
              spec {
                targetNamespace
                projectName
              }
            }
          }
        }
      }
    `,
    {
      dataPath: 'core_listWorkspaces',
    }
  ),
});
