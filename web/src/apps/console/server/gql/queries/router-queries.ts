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
        $projectName: String!
        $envName: String!
        $router: RouterIn!
      ) {
        core_createRouter(
          projectName: $projectName
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
        $projectName: String!
        $envName: String!
        $router: RouterIn!
      ) {
        core_updateRouter(
          projectName: $projectName
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
        $projectName: String!
        $envName: String!
        $routerName: String!
      ) {
        core_deleteRouter(
          projectName: $projectName
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
      vars(_: ConsoleListRoutersQueryVariables) {},
    }
  ),
  getRouter: executor(
    gql`
      query Core_getRouter(
        $projectName: String!
        $envName: String!
        $name: String!
      ) {
        core_getRouter(
          projectName: $projectName
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
          projectName
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
    `,
    {
      transformer: (data: ConsoleGetRouterQuery) => data.core_getRouter,
      vars(_: ConsoleGetRouterQueryVariables) {},
    }
  ),
});
