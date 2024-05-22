import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListExternalAppsQuery,
  ConsoleListExternalAppsQueryVariables,
  ConsoleCreateExternalAppMutation,
  ConsoleDeleteExternalAppMutation,
  ConsoleDeleteExternalAppMutationVariables,
  ConsoleCreateExternalAppMutationVariables,
  ConsoleGetExternalAppQuery,
  ConsoleGetExternalAppQueryVariables,
  ConsoleUpdateExternalAppMutation,
  ConsoleUpdateExternalAppMutationVariables,
  ConsoleInterceptExternalAppMutation,
  ConsoleInterceptExternalAppMutationVariables,
} from '~/root/src/generated/gql/server';

export type IExternalApp = NN<
  ConsoleGetExternalAppQuery['core_getExternalApp']
>;
export type IExternalApps = NN<
  ConsoleListExternalAppsQuery['core_listExternalApps']
>;

export const externalAppQueries = (executor: IExecutor) => ({
  createExternalApp: executor(
    gql`
      mutation Core_createExternalApp(
        $envName: String!
        $externalApp: ExternalAppIn!
      ) {
        core_createExternalApp(envName: $envName, externalApp: $externalApp) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateExternalAppMutation) =>
        data.core_createExternalApp,
      vars(_: ConsoleCreateExternalAppMutationVariables) {},
    }
  ),

  updateExternalApp: executor(
    gql`
      mutation Core_updateExternalApp(
        $envName: String!
        $externalApp: ExternalAppIn!
      ) {
        core_updateExternalApp(envName: $envName, externalApp: $externalApp) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateExternalAppMutation) => {
        return data.core_updateExternalApp;
      },
      vars(_: ConsoleUpdateExternalAppMutationVariables) {},
    }
  ),
  interceptExternalApp: executor(
    gql`
      mutation Core_interceptExternalApp(
        $envName: String!
        $externalAppName: String!
        $deviceName: String!
        $intercept: Boolean!
        $portMappings: [Github__com___kloudlite___operator___apis___crds___v1__AppInterceptPortMappingsIn!]
      ) {
        core_interceptExternalApp(
          envName: $envName
          externalAppName: $externalAppName
          deviceName: $deviceName
          intercept: $intercept
          portMappings: $portMappings
        )
      }
    `,
    {
      transformer: (data: ConsoleInterceptExternalAppMutation) =>
        data.core_interceptExternalApp,
      vars(_: ConsoleInterceptExternalAppMutationVariables) {},
    }
  ),
  deleteExternalApp: executor(
    gql`
      mutation Core_deleteExternalApp(
        $envName: String!
        $externalAppName: String!
      ) {
        core_deleteExternalApp(
          envName: $envName
          externalAppName: $externalAppName
        )
      }
    `,
    {
      transformer: (data: ConsoleDeleteExternalAppMutation) =>
        data.core_deleteExternalApp,
      vars(_: ConsoleDeleteExternalAppMutationVariables) {},
    }
  ),
  getExternalApp: executor(
    gql`
      query Core_getExternalApp($envName: String!, $name: String!) {
        core_getExternalApp(envName: $envName, name: $name) {
          accountName
          apiVersion
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
          environmentName
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
          recordVersion
          spec {
            intercept {
              enabled
              portMappings {
                appPort
                devicePort
              }
              toDevice
            }
            record
            recordType
          }
          status {
            checkList {
              debug
              description
              hide
              name
              title
            }
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
      transformer(data: ConsoleGetExternalAppQuery) {
        return data.core_getExternalApp;
      },
      vars(_: ConsoleGetExternalAppQueryVariables) {},
    }
  ),
  listExternalApps: executor(
    gql`
      query Core_listExternalApps(
        $envName: String!
        $search: SearchExternalApps
        $pq: CursorPaginationIn
      ) {
        core_listExternalApps(envName: $envName, search: $search, pq: $pq) {
          edges {
            cursor
            node {
              accountName
              apiVersion
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              environmentName
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
              recordVersion
              spec {
                intercept {
                  enabled
                  portMappings {
                    appPort
                    devicePort
                  }
                  toDevice
                }
                record
                recordType
              }
              status {
                checkList {
                  debug
                  description
                  hide
                  name
                  title
                }
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
      transformer: (data: ConsoleListExternalAppsQuery) =>
        data.core_listExternalApps,
      vars(_: ConsoleListExternalAppsQueryVariables) {},
    }
  ),
});
