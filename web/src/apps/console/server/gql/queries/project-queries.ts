import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';

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
      transformer(data) {},
      vars(variables) {},
    }
  ),
  getProject: executor(
    gql`
      query Core_getProject($name: String!) {
        core_getProject(name: $name) {
          metadata {
            name
            annotations
            namespace
          }
          spec {
            targetNamespace
            displayName
          }
        }
      }
    `,
    {
      transformer(data) {},
      vars(variables) {},
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
      transformer(data) {},
      vars(variables) {},
    }
  ),
});
