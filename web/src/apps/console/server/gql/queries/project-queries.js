import gql from 'graphql-tag';
import { ExecuteQueryWithContext } from '~/root/lib/server/helpers/execute-query-with-context';

export const projectQueries = (executor = ExecuteQueryWithContext({})) => ({
  createProject: executor(
    gql`
      mutation Core_createProject($project: ProjectIn!) {
        core_createProject(project: $project) {
          id
        }
      }
    `,
    {
      dataPath: 'core_createProject',
    }
  ),
  getProject: executor(
    gql`
      query Core_getProject($name: String!) {
        core_getProject(name: $name) {
          metadata {
            name
            annotations
          }
        }
      }
    `,
    {
      dataPath: 'core_getProject',
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
      dataPath: 'core_listProjects',
    }
  ),
});
