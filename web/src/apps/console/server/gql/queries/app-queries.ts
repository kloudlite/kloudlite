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
  ConsoleRemoveDeviceInterceptsMutation,
  ConsoleRemoveDeviceInterceptsMutationVariables,
} from '~/root/src/generated/gql/server';

export type IApp = NN<ConsoleGetAppQuery['core_getApp']>;
export type IApps = NN<ConsoleListAppsQuery['core_listApps']>;

export const appQueries = (executor: IExecutor) => ({
  restartApp: executor(
    gql`
      query Query($envName: String!, $appName: String!) {
        core_restartApp(envName: $envName, appName: $appName)
      }
    `,
    {
      transformer: (data: ConsoleRestartAppQuery) => data.core_restartApp,
      vars: (_: ConsoleRestartAppQueryVariables) => {},
    }
  ),
  createApp: executor(
    gql`
      mutation Core_createApp($envName: String!, $app: AppIn!) {
        core_createApp(envName: $envName, app: $app) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateAppMutation) => data.core_createApp,
      vars(_: ConsoleCreateAppMutationVariables) {},
    }
  ),

  updateApp: executor(
    gql`
      mutation Core_updateApp($envName: String!, $app: AppIn!) {
        core_updateApp(envName: $envName, app: $app) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateAppMutation) => {
        return data.core_updateApp;
      },
      vars(_: ConsoleUpdateAppMutationVariables) {},
    }
  ),
  interceptApp: executor(
    gql`
      mutation Core_interceptApp(
        $portMappings: [Github__com___kloudlite___operator___apis___crds___v1__AppInterceptPortMappingsIn!]
        $intercept: Boolean!
        $deviceName: String!
        $appname: String!
        $envName: String!
      ) {
        core_interceptApp(
          portMappings: $portMappings
          intercept: $intercept
          deviceName: $deviceName
          appname: $appname
          envName: $envName
        )
      }
    `,
    {
      transformer: (data: ConsoleInterceptAppMutation) =>
        data.core_interceptApp,
      vars(_: ConsoleInterceptAppMutationVariables) {},
    }
  ),
  removeDeviceIntercepts: executor(
    gql`
      mutation Core_removeDeviceIntercepts(
        $envName: String!
        $deviceName: String!
      ) {
        core_removeDeviceIntercepts(envName: $envName, deviceName: $deviceName)
      }
    `,
    {
      transformer: (data: ConsoleRemoveDeviceInterceptsMutation) =>
        data.core_removeDeviceIntercepts,
      vars(_: ConsoleRemoveDeviceInterceptsMutationVariables) {},
    }
  ),
  deleteApp: executor(
    gql`
      mutation Core_deleteApp($envName: String!, $appName: String!) {
        core_deleteApp(envName: $envName, appName: $appName)
      }
    `,
    {
      transformer: (data: ConsoleDeleteAppMutation) => data.core_deleteApp,
      vars(_: ConsoleDeleteAppMutationVariables) {},
    }
  ),
  getApp: executor(
    gql`
      query Core_getApp($envName: String!, $name: String!) {
        core_getApp(envName: $envName, name: $name) {
          id
          recordVersion
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
          spec {
            router {
              domains
            }
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
              portMappings {
                devicePort
                appPort
              }
            }
            nodeSelector
            region
            replicas
            serviceAccount
            services {
              port
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
              debug
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
          ciBuildId
          updateTime
          build {
            id
            buildClusterName
            name
            source {
              branch
              provider
              repository
            }
            spec {
              buildOptions {
                buildArgs
                buildContexts
                contextDir
                dockerfileContent
                dockerfilePath
                targetPlatforms
              }
              registry {
                repo {
                  name
                  tags
                }
              }
              resource {
                cpu
                memoryInMb
              }
            }
          }
        }
      }
    `,
    {
      transformer(data: ConsoleGetAppQuery) {
        return data.core_getApp;
      },
      vars(_: ConsoleGetAppQueryVariables) {},
    }
  ),
  listApps: executor(
    gql`
      query Core_listApps(
        $envName: String!
        $pq: CursorPaginationIn
        $search: SearchApps
      ) {
        core_listApps(envName: $envName, pq: $pq, search: $search) {
          edges {
            cursor
            node {
              accountName
              apiVersion
              ciBuildId
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              enabled
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
                  portMappings {
                    appPort
                    devicePort
                  }
                  toDevice
                }
                nodeSelector
                region
                replicas
                router {
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
                serviceAccount
                services {
                  port
                  protocol
                }
                tolerations {
                  effect
                  key
                  operator
                  tolerationSeconds
                  value
                }
                topologySpreadConstraints {
                  labelSelector {
                    matchExpressions {
                      key
                      operator
                      values
                    }
                    matchLabels
                  }
                  matchLabelKeys
                  maxSkew
                  minDomains
                  nodeAffinityPolicy
                  nodeTaintsPolicy
                  topologyKey
                  whenUnsatisfiable
                }
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
              build {
                id
                buildClusterName
                creationTime
                credUser {
                  userEmail
                  userId
                  userName
                }
                errorMessages
                markedForDeletion
                name
                recordVersion
                source {
                  branch
                  provider
                  repository
                  webhookId
                }
                spec {
                  accountName
                  buildOptions {
                    buildArgs
                    buildContexts
                    contextDir
                    dockerfileContent
                    dockerfilePath
                    targetPlatforms
                  }
                  caches {
                    name
                    path
                  }
                  registry {
                    repo {
                      name
                      tags
                    }
                  }
                  resource {
                    cpu
                    memoryInMb
                  }
                }
                status
                updateTime
                latestBuildRun {
                  accountName
                  apiVersion
                  buildId
                  clusterName
                  creationTime
                  displayName
                  id
                  kind
                  markedForDeletion
                  recordVersion
                  updateTime
                }
              }
              serviceHost
            }
          }
          pageInfo {
            endCursor
            hasNextPage
            hasPrevPage
            startCursor
          }
          totalCount
        }
      }
    `,
    {
      transformer: (data: ConsoleListAppsQuery) => data.core_listApps,
      vars(_: ConsoleListAppsQueryVariables) {},
    }
  ),
});
