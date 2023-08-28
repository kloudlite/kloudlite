import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const workspaceQueries = (executor = ExecuteQueryWithContext({})) => ({
  getWorkspace: executor(
    gql`
      query Core_getWorkspace($project: ProjectId!, $name: String!) {
        core_getWorkspace(project: $project, name: $name) {
          spec {
            targetNamespace
            projectName
          }
          displayName
          metadata {
            namespace
            name
            annotations
            labels
          }
          updateTime
        }
      }
    `,
    { dataPath: 'core_getWorkspace' }
  ),
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
        $project: ProjectId!
        $search: SearchWorkspaces
        $pagination: CursorPaginationIn
      ) {
        core_listWorkspaces(
          project: $project
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
                name
                namespace
                labels
                annotations
              }
              displayName
              clusterName
              updateTime
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
