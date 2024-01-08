import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  ConsoleListRoutersQuery,
  ConsoleListRoutersQueryVariables,
} from '~/root/src/generated/gql/server';

export const routerQueries = (executor: IExecutor) => ({
  listRouters: executor(
    gql`
      query Core_listRouters(
        $projectName: String!
        $envName: String!
        $search: SearchRouters
        $pq: CursorPaginationIn
      ) {
        core_listRouters(
          projectName: $projectName
          envName: $envName
          search: $search
          pq: $pq
        ) {
          edges {
            cursor
            node {
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              enabled
              environmentName
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
              spec {
                routes {
                  app
                  lambda
                  path
                  port
                  rewrite
                }
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
      transformer: (data: ConsoleListRoutersQuery) => data.core_listRouters,
      vars(_: ConsoleListRoutersQueryVariables) { },
    }
  ),
});
