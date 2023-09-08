import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListAppsQuery,
  ConsoleListAppsQueryVariables,
  ConsoleCreateAppMutation,
  ConsoleCreateAppMutationVariables,
  ConsoleGetAppQuery,
  ConsoleGetAppQueryVariables,
} from '~/root/src/generated/gql/server';

export type IApp = NN<ConsoleGetAppQuery['core_getApp']>;
export type IApps = NN<ConsoleListAppsQuery['core_listApps']>;

export const appQueries = (executor: IExecutor) => ({
  createApp: executor(
    gql`
      mutation Core_createApp($app: AppIn!) {
        core_createApp(app: $app) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateAppMutation) => data.core_createApp,
      vars(_: ConsoleCreateAppMutationVariables) {},
    }
  ),
  getApp: executor(
    gql`
      query Core_getApp(
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $name: String!
      ) {
        core_getApp(project: $project, scope: $scope, name: $name) {
          creationTime
          accountName
          displayName
          createdBy {
            userName
            userId
            userEmail
          }
          lastUpdatedBy {
            userName
            userId
            userEmail
          }
          markedForDeletion
          metadata {
            name
          }

          updateTime
          spec {
            tolerations {
              value
              tolerationSeconds
              operator
              key
              effect
            }
            services {
              type
              targetPort
              port
              name
            }
            serviceAccount
            replicas
            region
            nodeSelector
            intercept {
              enabled
              toDevice
            }
            hpa {
              maxReplicas
              enabled
              minReplicas
              thresholdCpu
              thresholdMemory
            }
            freeze
            displayName
            containers {
              args
              command
              env {
                refName
                refKey
                optional
                key
                type
                value
              }
              envFrom {
                type
                refName
              }
              image
              imagePullPolicy
              livenessProbe {
                type
                tcp {
                  port
                }
                shell {
                  command
                }
                interval
                initialDelay
                httpGet {
                  httpHeaders
                  path
                  port
                }
                failureThreshold
              }
              name
              readinessProbe {
                type
                tcp {
                  port
                }
                shell {
                  command
                }
                interval
                initialDelay
                httpGet {
                  httpHeaders
                  path
                  port
                }
                failureThreshold
              }
              resourceCpu {
                min
                max
              }
              resourceMemory {
                min
                max
              }
              volumes {
                type
                refName
                mountPath
                items {
                  fileName
                  key
                }
              }
            }
          }

          status {
            resources {
              namespace
              name
              kind
              apiVersion
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
        $project: ProjectId!
        $scope: WorkspaceOrEnvId!
        $search: SearchApps
        $pagination: CursorPaginationIn
      ) {
        core_listApps(
          project: $project
          scope: $scope
          search: $search
          pq: $pagination
        ) {
          totalCount
          pageInfo {
            startCursor
            hasPreviousPage
            hasNextPage
            endCursor
          }

          edges {
            cursor
            node {
              creationTime
              displayName
              createdBy {
                userName
                userId
                userEmail
              }
              lastUpdatedBy {
                userName
                userId
                userEmail
              }
              markedForDeletion
              metadata {
                name
              }

              updateTime
              spec {
                freeze
                displayName
              }

              status {
                resources {
                  namespace
                  name
                  kind
                  apiVersion
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
            }
          }
        }
      }
    `,
    {
      transformer: (data: ConsoleListAppsQuery) => data.core_listApps,
      vars(_: ConsoleListAppsQueryVariables) {},
    }
  ),
});
