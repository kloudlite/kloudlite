import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  IotconsoleCreateIotAppMutation,
  IotconsoleCreateIotAppMutationVariables,
  IotconsoleGetIotAppQuery,
  IotconsoleGetIotAppQueryVariables,
  IotconsoleListIotAppsQuery,
  IotconsoleListIotAppsQueryVariables,
  IotconsoleUpdateIotAppMutation,
  IotconsoleUpdateIotAppMutationVariables,
  IotconsoleDeleteIotAppMutation,
  IotconsoleDeleteIotAppMutationVariables,
} from '~/root/src/generated/gql/server';

export type IApps = NN<IotconsoleListIotAppsQuery['iot_listApps']>;
export type IApp = NN<IotconsoleGetIotAppQuery['iot_getApp']>;

export const iotAppQueries = (executor: IExecutor) => ({
  deleteIotApp: executor(
    gql`
      mutation Iot_deleteApp(
        $projectName: String!
        $deviceBlueprintName: String!
        $name: String!
      ) {
        iot_deleteApp(
          projectName: $projectName
          deviceBlueprintName: $deviceBlueprintName
          name: $name
        )
      }
    `,
    {
      transformer: (data: IotconsoleDeleteIotAppMutation) => data.iot_deleteApp,
      vars(_: IotconsoleDeleteIotAppMutationVariables) {},
    }
  ),
  createIotApp: executor(
    gql`
      mutation Iot_createApp(
        $projectName: String!
        $deviceBlueprintName: String!
        $app: IOTAppIn!
      ) {
        iot_createApp(
          projectName: $projectName
          deviceBlueprintName: $deviceBlueprintName
          app: $app
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: IotconsoleCreateIotAppMutation) => data.iot_createApp,
      vars(_: IotconsoleCreateIotAppMutationVariables) {},
    }
  ),
  updateIotApp: executor(
    gql`
      mutation Iot_updateApp(
        $projectName: String!
        $deviceBlueprintName: String!
        $app: IOTAppIn!
      ) {
        iot_updateApp(
          projectName: $projectName
          deviceBlueprintName: $deviceBlueprintName
          app: $app
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: IotconsoleUpdateIotAppMutation) => data.iot_updateApp,
      vars(_: IotconsoleUpdateIotAppMutationVariables) {},
    }
  ),
  getIotApp: executor(
    gql`
      query Iot_getApp(
        $projectName: String!
        $deviceBlueprintName: String!
        $name: String!
      ) {
        iot_getApp(
          projectName: $projectName
          deviceBlueprintName: $deviceBlueprintName
          name: $name
        ) {
          accountName
          apiVersion
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          deviceBlueprintName
          displayName
          enabled
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
            labels
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
          updateTime
        }
      }
    `,
    {
      transformer: (data: IotconsoleGetIotAppQuery) => data.iot_getApp,
      vars(_: IotconsoleGetIotAppQueryVariables) {},
    }
  ),
  listIotApps: executor(
    gql`
      query Iot_listApps(
        $projectName: String!
        $deviceBlueprintName: String!
        $search: SearchIOTApps
        $pq: CursorPaginationIn
      ) {
        iot_listApps(
          projectName: $projectName
          deviceBlueprintName: $deviceBlueprintName
          search: $search
          pq: $pq
        ) {
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
              deviceBlueprintName
              displayName
              enabled
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
                labels
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
      transformer: (data: IotconsoleListIotAppsQuery) => {
        return data.iot_listApps;
      },
      vars(_: IotconsoleListIotAppsQueryVariables) {},
    }
  ),
});
