import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateEnvironmentMutation,
  ConsoleCreateEnvironmentMutationVariables,
  ConsoleGetEnvironmentQuery,
  ConsoleGetEnvironmentQueryVariables,
  ConsoleListEnvironmentsQuery,
  ConsoleListEnvironmentsQueryVariables,
  ConsoleUpdateEnvironmentMutation,
  ConsoleUpdateEnvironmentMutationVariables,
} from '~/root/src/generated/gql/server';

export type IEnvironment = NN<
  ConsoleGetEnvironmentQuery['core_getEnvironment']
>;

export type IEnvironments = NN<
  ConsoleListEnvironmentsQuery['core_listEnvironments']
>;

export const environmentQueries = (executor: IExecutor) => ({
  getEnvironment: executor(
    gql`
      query Core_getEnvironment($project: ProjectId!, $name: String!) {
        core_getEnvironment(project: $project, name: $name) {
          accountName
          apiVersion
          clusterName
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
          id
          kind
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          markedForDeletion
          metadata {
            annotations
            creationTimestamp
            deletionTimestamp
            generation
            labels
            name
            namespace
          }
          projectName
          recordVersion
          spec {
            isEnvironment
            projectName
            targetNamespace
          }
          status {
            checks
            isReady
            lastReconcileTime
            message {
              RawMessage
            }
            resources {
              apiVersion
              kind
              name
              namespace
            }
          }
          updateTime
        }
      }
    `,
    {
      transformer: (data: ConsoleGetEnvironmentQuery) =>
        data.core_getEnvironment,
      vars(_: ConsoleGetEnvironmentQueryVariables) { },
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
      vars(_: ConsoleCreateEnvironmentMutationVariables) { },
    }
  ),
  updateEnvironment: executor(
    gql`
      mutation Core_updateEnvironment($env: WorkspaceIn!) {
        core_updateEnvironment(env: $env) {
          id
        }
      }
    `,
    {
      transformer(data: ConsoleUpdateEnvironmentMutation) {
        return data.core_updateEnvironment;
      },
      vars(_: ConsoleUpdateEnvironmentMutationVariables) { },
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
          edges {
            cursor
            node {
              accountName
              apiVersion
              clusterName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              id
              kind
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              metadata {
                annotations
                creationTimestamp
                deletionTimestamp
                generation
                labels
                name
                namespace
              }
              projectName
              recordVersion
              spec {
                isEnvironment
                projectName
                targetNamespace
              }
              status {
                checks
                isReady
                lastReconcileTime
                message {
                  RawMessage
                }
                resources {
                  apiVersion
                  kind
                  name
                  namespace
                }
              }
              updateTime
            }
          }
          pageInfo {
            endCursor
            hasNextPage
            hasPreviousPage
            startCursor
          }
          totalCount
        }
      }
    `,
    {
      transformer: (data: ConsoleListEnvironmentsQuery) =>
        data.core_listEnvironments,
      vars(_: ConsoleListEnvironmentsQueryVariables) { },
    }
  ),
});
