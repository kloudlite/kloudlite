import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';

export const environmentQueries = (executor: IExecutor) => ({
  getEnvironment: executor(
    gql`
      query Core_getEnvironment($project: ProjectId!, $name: String!) {
        core_getEnvironment(project: $project, name: $name) {
          spec {
            targetNamespace
            projectName
          }
          updateTime
          displayName
          metadata {
            namespace
            name
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
  createEnvironment: executor(
    gql`
      mutation Core_createEnvironment($env: WorkspaceIn!) {
        core_createEnvironment(env: $env) {
          id
        }
      }
    `,
    {
      transformer(data) {},
      vars(variables) {},
    }
  ),
  listEnvironments: executor(
    gql`
      query Core_listEnvironments(
        $project: ProjectId!
        $search: SearchWorkspaces
        $pagination: CursorPaginationIn
      ) {
        core_listEnvironments(
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
      transformer(data) {},
      vars(variables) {},
    }
  ),
});
