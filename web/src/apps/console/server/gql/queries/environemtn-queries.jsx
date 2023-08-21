import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const environmentQueries = (executor = ExecuteQueryWithContext({})) => ({
  getEnvironment: executor(
    gql`
      query Core_getEnvironment($project: ProjectId!, $name: String!) {
        core_getEnvironment(project: $project, name: $name) {
          spec {
            targetNamespace
            projectName
          }
          updateTime
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
      dataPath: 'core_getEnvironment',
    }
  ),
  createEnvironment: executor(gql`
    mutation Core_createEnvironment($env: EnvironmentIn!) {
      core_createEnvironment(env: $env) {
        id
      }
    }
  `),
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
    { dataPath: 'core_listEnvironments' }
  ),
});
