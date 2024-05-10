import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListRoutersQuery,
  ConsoleGetRouterQuery,
  ConsoleGetRouterQueryVariables,
  ConsoleCreateRouterMutation,
  ConsoleUpdateRouterMutation,
  ConsoleDeleteRouterMutation,
  ConsoleCreateRouterMutationVariables,
  ConsoleDeleteRouterMutationVariables,
  ConsoleUpdateRouterMutationVariables,
  ConsoleListRoutersQueryVariables,
} from '~/root/src/generated/gql/server';

export type IRouters = NN<ConsoleListRoutersQuery['core_listRouters']>;
export type IRouter = NN<ConsoleGetRouterQuery['core_getRouter']>;

export const routerQueries = (executor: IExecutor) => ({
  createRouter: executor(
    gql`
      mutation Core_createRouter(
        $envName: String!
        $router: RouterIn!
      ) {
        core_createRouter(
          envName: $envName
          router: $router
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateRouterMutation) =>
        data.core_createRouter,
      vars(_: ConsoleCreateRouterMutationVariables) {},
    }
  ),
  updateRouter: executor(
    gql`
      mutation Core_updateRouter(
        $envName: String!
        $router: RouterIn!
      ) {
        core_updateRouter(
          envName: $envName
          router: $router
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateRouterMutation) =>
        data.core_updateRouter,
      vars(_: ConsoleUpdateRouterMutationVariables) {},
    }
  ),
  deleteRouter: executor(
    gql`
      mutation Core_deleteRouter(
        $envName: String!
        $routerName: String!
      ) {
        core_deleteRouter(
          envName: $envName
          routerName: $routerName
        )
      }
    `,
    {
      transformer: (data: ConsoleDeleteRouterMutation) =>
        data.core_deleteRouter,
      vars(_: ConsoleDeleteRouterMutationVariables) {},
    }
  ),
  listRouters: executor(
    gql`
      query Core_listRouters(
        $envName: String!
        $search: SearchRouters
        $pq: CursorPaginationIn
      ) {
        core_listRouters(
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
                generation
                name
                namespace
              }
              recordVersion
              spec {
                backendProtocol
                basicAuth {
                  enabled
                  secretName
                  username
                }
                cors {
                  allowCredentials
                  enabled
                  origins
                }
                domains
                https {
                  clusterIssuer
                  enabled
                  forceRedirect
                }
                ingressClass
                maxBodySizeInMB
                rateLimit {
                  connections
                  enabled
                  rpm
                  rps
                }
                routes {
                  app
                  path
                  port
                  rewrite
                }
              }
              status {
                checks
                checkList {
                  description
                  debug
                  name
                  title
                }
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
      transformer: (data: ConsoleListRoutersQuery) => data.core_listRouters,
      vars(_: ConsoleListRoutersQueryVariables) {},
    }
  ),
  getRouter: executor(
    gql`
      query Core_getRouter(
        $envName: String!
        $name: String!
      ) {
        core_getRouter(
          envName: $envName
          name: $name
        ) {
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
            name
            namespace
          }
          spec {
            backendProtocol
            basicAuth {
              enabled
              secretName
              username
            }
            cors {
              allowCredentials
              enabled
              origins
            }
            domains
            https {
              clusterIssuer
              enabled
              forceRedirect
            }
            ingressClass
            maxBodySizeInMB
            rateLimit {
              connections
              enabled
              rpm
              rps
            }
            routes {
              app
              path
              port
              rewrite
            }
          }
          status {
            checks
            checkList {
              description
              debug
              name
              title
            }
            isReady
          }
          updateTime
        }
      }
    `,
    {
      transformer: (data: ConsoleGetRouterQuery) => data.core_getRouter,
      vars(_: ConsoleGetRouterQueryVariables) {},
    }
  ),
});
