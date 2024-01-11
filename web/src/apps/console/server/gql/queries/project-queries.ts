import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateProjectMutation,
  ConsoleCreateProjectMutationVariables,
  ConsoleGetProjectQuery,
  ConsoleGetProjectQueryVariables,
  ConsoleListProjectsQuery,
  ConsoleListProjectsQueryVariables,
  ConsoleUpdateProjectMutation,
  ConsoleUpdateProjectMutationVariables,
} from '~/root/src/generated/gql/server';

export type IProjects = NN<ConsoleListProjectsQuery['core_listProjects']>;
export type IProject = NN<ConsoleGetProjectQuery['core_getProject']>;

export const projectQueries = (executor: IExecutor) => ({
  createProject: executor(
    gql`
      mutation Core_createProject($project: ProjectIn!) {
        core_createProject(project: $project) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateProjectMutation) =>
        data.core_createProject,
      vars(_: ConsoleCreateProjectMutationVariables) { },
    }
  ),
  updateProject: executor(
    gql`
      mutation Core_updateProject($project: ProjectIn!) {
        core_updateProject(project: $project) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateProjectMutation) =>
        data.core_updateProject,
      vars(_: ConsoleUpdateProjectMutationVariables) { },
    }
  ),
  getProject: executor(
    gql`
      query Core_getProject($name: String!) {
        core_getProject(name: $name) {
          clusterName
          displayName
          metadata {
            name
            namespace
          }
          spec {
            targetNamespace
          }
          accountName
          apiVersion
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          id
          kind
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          markedForDeletion
          recordVersion
          status {
            checks
            isReady
            lastReadyGeneration
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
          syncStatus {
            action
            error
            lastSyncedAt
            recordVersion
            state
            syncScheduledAt
          }
          updateTime
        }
      }
    `,
    {
      transformer: (data: ConsoleGetProjectQuery) => data.core_getProject,
      vars(_: ConsoleGetProjectQueryVariables) { },
    }
  ),
  listProjects: executor(
    gql`
      query Core_listProjects(
        $search: SearchProjects
        $pq: CursorPaginationIn
      ) {
        core_listProjects(search: $search, pq: $pq) {
          edges {
            cursor
            node {
              clusterName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              metadata {
                name
                namespace
              }
              spec {
                targetNamespace
              }
              status {
                checks
                isReady
                lastReadyGeneration
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
              syncStatus {
                action
                error
                lastSyncedAt
                recordVersion
                state
                syncScheduledAt
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
      transformer: (data: ConsoleListProjectsQuery) => data.core_listProjects,
      vars(_: ConsoleListProjectsQueryVariables) { },
    }
  ),
});
