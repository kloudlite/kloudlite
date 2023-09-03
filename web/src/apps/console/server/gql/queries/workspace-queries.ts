import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  ConsoleCreateWorkspaceMutation,
  ConsoleCreateWorkspaceMutationVariables,
  ConsoleGetWorkspaceQuery,
  ConsoleGetWorkspaceQueryVariables,
  ConsoleListWorkspacesQuery,
  ConsoleListWorkspacesQueryVariables,
} from '~/root/src/generated/gql/server';
import { NN } from '~/root/src/generated/r-types/utils';

export type Workspace = NN<ConsoleGetWorkspaceQuery['core_getWorkspace']>;

export const workspaceQueries = (executor: IExecutor) => ({
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
    {
      transformer: (data: ConsoleGetWorkspaceQuery) => data.core_getWorkspace,
      vars(_: ConsoleGetWorkspaceQueryVariables) {},
    }
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
      transformer: (data: ConsoleCreateWorkspaceMutation) =>
        data.core_createWorkspace,
      vars(_: ConsoleCreateWorkspaceMutationVariables) {},
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
      transformer: (data: ConsoleListWorkspacesQuery) =>
        data.core_listWorkspaces,
      vars(_: ConsoleListWorkspacesQueryVariables) {},
    }
  ),
});
