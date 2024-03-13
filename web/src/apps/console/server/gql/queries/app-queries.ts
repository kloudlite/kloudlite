import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListAppsQuery,
  ConsoleListAppsQueryVariables,
  ConsoleCreateAppMutation,
  ConsoleDeleteAppMutation,
  ConsoleDeleteAppMutationVariables,
  ConsoleCreateAppMutationVariables,
  ConsoleGetAppQuery,
  ConsoleGetAppQueryVariables,
  ConsoleUpdateAppMutation,
  ConsoleUpdateAppMutationVariables,
  ConsoleInterceptAppMutation,
  ConsoleInterceptAppMutationVariables,
  ConsoleRestartAppQuery,
  ConsoleRestartAppQueryVariables,
} from '~/root/src/generated/gql/server';

export type IApp = NN<ConsoleGetAppQuery['core_getApp']>;
export type IApps = NN<ConsoleListAppsQuery['core_listApps']>;

export const appQueries = (executor: IExecutor) => ({
  restartApp: executor(
    gql`
      query Query($projectName: String!, $envName: String!, $appName: String!) {
        core_restartApp(
          projectName: $projectName
          envName: $envName
          appName: $appName
        )
      }
    `,
    {
      transformer: (data: ConsoleRestartAppQuery) => data.core_restartApp,
      vars: (_: ConsoleRestartAppQueryVariables) => { },
    }
  ),
  createApp: executor(
    gql`
      mutation Core_createApp(
        $projectName: String!
        $envName: String!
        $app: AppIn!
      ) {
        core_createApp(
          projectName: $projectName
          envName: $envName
          app: $app
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateAppMutation) => data.core_createApp,
      vars(_: ConsoleCreateAppMutationVariables) { },
    }
  ),

  updateApp: executor(
    gql`
      mutation Core_updateApp(
        $projectName: String!
        $envName: String!
        $app: AppIn!
      ) {
        core_updateApp(
          projectName: $projectName
          envName: $envName
          app: $app
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateAppMutation) => {
        return data.core_updateApp;
      },
      vars(_: ConsoleUpdateAppMutationVariables) { },
    }
  ),
  interceptApp: executor(
    gql`
      mutation Core_interceptApp(
        $projectName: String!
        $envName: String!
        $appname: String!
        $deviceName: String!
        $intercept: Boolean!
      ) {
        core_interceptApp(
          projectName: $projectName
          envName: $envName
          appname: $appname
          deviceName: $deviceName
          intercept: $intercept
        )
      }
    `,
    {
      transformer: (data: ConsoleInterceptAppMutation) =>
        data.core_interceptApp,
      vars(_: ConsoleInterceptAppMutationVariables) { },
    }
  ),
  deleteApp: executor(
    gql`
      mutation Core_deleteApp(
        $projectName: String!
        $envName: String!
        $appName: String!
      ) {
        core_deleteApp(
          projectName: $projectName
          envName: $envName
          appName: $appName
        )
      }
    `,
    {
      transformer: (data: ConsoleDeleteAppMutation) => data.core_deleteApp,
      vars(_: ConsoleDeleteAppMutationVariables) { },
    }
  ),
  getApp: executor(
    gql`
      query Core_getApp(
        $projectName: String!
        $envName: String!
        $name: String!
      ) {
        core_getApp(projectName: $projectName, envName: $envName, name: $name) {
          id
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
            name
            namespace
          }
          projectName
          spec {
            containers {
              args
              command
              env {
                key
                optional
                refKey
                refName
                type
                value
              }
              envFrom {
                refName
                type
              }
              image
              imagePullPolicy
              livenessProbe {
                failureThreshold
                httpGet {
                  httpHeaders
                  path
                  port
                }
                initialDelay
                interval
                shell {
                  command
                }
                tcp {
                  port
                }
                type
              }
              name
              readinessProbe {
                failureThreshold
                initialDelay
                interval
                type
              }
              resourceCpu {
                max
                min
              }
              resourceMemory {
                max
                min
              }
              volumes {
                items {
                  fileName
                  key
                }
                mountPath
                refName
                type
              }
            }
            displayName
            freeze
            hpa {
              enabled
              maxReplicas
              minReplicas
              thresholdCpu
              thresholdMemory
            }
            intercept {
              enabled
              toDevice
            }
            nodeSelector
            region
            replicas
            serviceAccount
            services {
              name
              port
              targetPort
              type
            }
            tolerations {
              effect
              key
              operator
              tolerationSeconds
              value
            }
          }
          status {
            checkList {
              description
              title
              name
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
          updateTime
        }
      }
    `,
    {
      transformer(data: ConsoleGetAppQuery) {
        return data.core_getApp;
      },
      vars(_: ConsoleGetAppQueryVariables) { },
    }
  ),
  listApps: executor(
    gql`
      query Core_listApps(
        $projectName: String!
        $envName: String!
        $search: SearchApps
        $pq: CursorPaginationIn
      ) {
        core_listApps(
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
                generation
                name
                namespace
              }
              projectName
              recordVersion
              spec {
                containers {
                  args
                  command
                  env {
                    key
                    optional
                    refKey
                    refName
                    type
                    value
                  }
                  envFrom {
                    refName
                    type
                  }
                  image
                  imagePullPolicy
                  name
                  readinessProbe {
                    failureThreshold
                    initialDelay
                    interval
                    type
                  }
                  resourceCpu {
                    max
                    min
                  }
                  resourceMemory {
                    max
                    min
                  }
                }
                displayName
                freeze
                hpa {
                  enabled
                  maxReplicas
                  minReplicas
                  thresholdCpu
                  thresholdMemory
                }
                intercept {
                  enabled
                  toDevice
                }
                nodeSelector
                region
                replicas
                serviceAccount
                services {
                  name
                  port
                  targetPort
                  type
                }
                tolerations {
                  effect
                  key
                  operator
                  tolerationSeconds
                  value
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
                checkList {
                  description
                  title
                  name
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
      transformer: (data: ConsoleListAppsQuery) => data.core_listApps,
      vars(_: ConsoleListAppsQueryVariables) { },
    }
  ),
});
