import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateWorkspaceMutation,
  ConsoleCreateWorkspaceMutationVariables,
  ConsoleGetWorkspaceQuery,
  ConsoleGetWorkspaceQueryVariables,
  ConsoleListWorkspacesQuery,
  ConsoleListWorkspacesQueryVariables,
  ConsoleUpdateWorkspaceMutation,
  ConsoleUpdateWorkspaceMutationVariables,
} from '~/root/src/generated/gql/server';

export type IWorkspace = NN<ConsoleGetWorkspaceQuery['core_getWorkspace']>;
export type IWorkspaces = NN<ConsoleListWorkspacesQuery['core_listWorkspaces']>;
export const workspaceQueries = (executor: IExecutor) => ({
  getWorkspace: executor(
    gql`
      query Core_getWorkspace($project: ProjectId!, $name: String!) {
        core_getWorkspace(project: $project, name: $name) {
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
  updateWorkspace: executor(
    gql`
      mutation Core_updateEnvironment($env: WorkspaceIn!) {
        core_updateWorkspace(env: $env) {
          id
        }
      }
    `,
    {
      transformer(data: ConsoleUpdateWorkspaceMutation) {
        return data.core_updateWorkspace;
      },
      vars(_: ConsoleUpdateWorkspaceMutationVariables) {},
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
              creationTime
              spec {
                targetNamespace
                projectName
              }
              createdBy {
                userEmail
                userId
                userName
              }
              lastUpdatedBy {
                userEmail
                userId
                userName
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
