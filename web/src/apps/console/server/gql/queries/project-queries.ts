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
      vars(_: ConsoleCreateProjectMutationVariables) {},
    }
  ),
  getProject: executor(
    gql`
      query Core_getProject($name: String!) {
        core_getProject(name: $name) {
          id
          displayName
          creationTime
          clusterName
          apiVersion
          kind
          metadata {
            namespace
            name
            labels
            deletionTimestamp
            generation
            creationTimestamp
            annotations
          }
          recordVersion
          spec {
            targetNamespace
            logo
            displayName
            clusterName
            accountName
          }
          status {
            resources {
              name
              kind
              apiVersion
              namespace
            }
            message {
              RawMessage
            }
            lastReconcileTime
            isReady
            checks
          }
          syncStatus {
            syncScheduledAt
            state
            recordVersion
            lastSyncedAt
            error
            action
          }
          updateTime
          accountName
        }
      }
    `,
    {
      transformer: (data: ConsoleGetProjectQuery) => data.core_getProject,
      vars(_: ConsoleGetProjectQueryVariables) {},
    }
  ),
  listProjects: executor(
    gql`
      query Core_listProjects(
        $clusterName: String
        $pagination: CursorPaginationIn
        $search: SearchProjects
      ) {
        core_listProjects(
          clusterName: $clusterName
          pq: $pagination
          search: $search
        ) {
          totalCount
          edges {
            node {
              id
              displayName
              creationTime
              clusterName
              apiVersion
              kind
              metadata {
                namespace
                name
                labels
                deletionTimestamp
                generation
                creationTimestamp
                annotations
              }
              recordVersion
              spec {
                targetNamespace
                logo
                displayName
                clusterName
                accountName
              }
              status {
                resources {
                  name
                  kind
                  apiVersion
                  namespace
                }
                message {
                  RawMessage
                }
                lastReconcileTime
                isReady
                checks
              }
              syncStatus {
                syncScheduledAt
                state
                recordVersion
                lastSyncedAt
                error
                action
              }
              updateTime
              accountName
            }
          }
          pageInfo {
            startCursor
            hasNextPage
            endCursor
            hasPreviousPage
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleListProjectsQuery) => data.core_listProjects,
      vars(_: ConsoleListProjectsQueryVariables) {},
    }
  ),
});
