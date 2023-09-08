import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateConfigMutationVariables,
  ConsoleCreateEnvironmentMutation,
  ConsoleGetEnvironmentQuery,
  ConsoleGetEnvironmentQueryVariables,
  ConsoleListEnvironmentsQuery,
  ConsoleListEnvironmentsQueryVariables,
} from '~/root/src/generated/gql/server';

export type Environment = NN<ConsoleGetEnvironmentQuery['core_getEnvironment']>;

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
      transformer: (data: ConsoleGetEnvironmentQuery) => data,
      vars(_: ConsoleGetEnvironmentQueryVariables) {},
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
      transformer: (data: ConsoleCreateEnvironmentMutation) =>
        data.core_createEnvironment,
      vars(_: ConsoleCreateConfigMutationVariables) {},
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
            cursor
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
      transformer: (data: ConsoleListEnvironmentsQuery) =>
        data.core_listEnvironments,
      vars(_: ConsoleListEnvironmentsQueryVariables) {},
    }
  ),
});
