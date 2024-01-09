/* eslint-disable camelcase */
import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  AuthCli_CreateRemoteLoginMutationVariables,
  AuthCli_CreateRemoteLoginMutation,
  AuthCli_GetRemoteLoginQueryVariables,
  AuthCli_GetRemoteLoginQuery,
  AuthCli_GetCurrentUserQuery,
  AuthCli_GetCurrentUserQueryVariables,
  AuthCli_ListAccountsQuery,
  AuthCli_ListAccountsQueryVariables,
  AuthCli_ListClustersQuery,
  AuthCli_ListClustersQueryVariables,
  AuthCli_GetKubeConfigQuery,
  AuthCli_GetKubeConfigQueryVariables,
} from '~/root/src/generated/gql/server';

export const cliQueries = (executor: IExecutor) => ({
  cli_getEnvironment: executor(
    gql`
      query Core_getEnvironment($projectName: String!, $name: String!) {
        core_getEnvironment(projectName: $projectName, name: $name) {
          spec {
            targetNamespace
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.core_getEnvironment,
      vars: (_: any) => {},
    }
  ),
  cli_updateDeviceNs: executor(
    gql`
      mutation Infra_updateVPNDeviceNs(
        $clusterName: String!
        $deviceName: String!
        $namespace: String!
      ) {
        infra_updateVPNDeviceNs(
          clusterName: $clusterName
          deviceName: $deviceName
          namespace: $namespace
        )
      }
    `,
    {
      transformer: (data: any) => data.infra_updateVPNDeviceNs,
      vars: (_: any) => {},
    }
  ),
  cli_updateDevicePort: executor(
    gql`
      mutation Mutation(
        $clusterName: String!
        $deviceName: String!
        $ports: [PortIn!]!
      ) {
        infra_updateVPNDevicePorts(
          clusterName: $clusterName
          deviceName: $deviceName
          ports: $ports
        )
      }
    `,
    {
      transformer: (data: any) => data.infra_updateVPNDevicePorts,
      vars: (_: any) => {},
    }
  ),
  cli_getSecret: executor(
    gql`
      query Core_getSecret(
        $projectName: String!
        $envName: String!
        $name: String!
      ) {
        core_getSecret(
          projectName: $projectName
          envName: $envName
          name: $name
        ) {
          displayName
          metadata {
            name
            namespace
          }
          stringData
        }
      }
    `,
    {
      transformer: (data: any) => data.core_getSecret,
      vars: (_: any) => {},
    }
  ),
  cli_getConfig: executor(
    gql`
      query Core_getConfig(
        $projectName: String!
        $envName: String!
        $name: String!
      ) {
        core_getConfig(
          projectName: $projectName
          envName: $envName
          name: $name
        ) {
          data
          displayName
          metadata {
            name
            namespace
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.core_getConfig,
      vars: (_: any) => {},
    }
  ),

  cli_listApps: executor(
    gql`
      query Core_listApps($projectName: String!, $envName: String!) {
        core_listApps(projectName: $projectName, envName: $envName) {
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
              projectName
              spec {
                displayName
                freeze
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
      transformer: (data: any) => data.core_listApps,
      vars: (_: any) => {},
    }
  ),
  cli_listConfigs: executor(
    gql`
      query Core_listConfigs($projectName: String!, $envName: String!) {
        core_listConfigs(projectName: $projectName, envName: $envName) {
          totalCount
          edges {
            node {
              data
              displayName
              metadata {
                name
                namespace
              }
            }
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.core_listConfigs,
      vars: (_: any) => {},
    }
  ),
  cli_listSecrets: executor(
    gql`
      query Core_listSecrets(
        $projectName: String!
        $envName: String!
        $pq: CursorPaginationIn
      ) {
        core_listSecrets(
          projectName: $projectName
          envName: $envName
          pq: $pq
        ) {
          edges {
            cursor
            node {
              displayName
              markedForDeletion
              metadata {
                name
                namespace
              }
              stringData
            }
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.core_listSecrets,
      vars: (_: any) => {},
    }
  ),
  cli_updateDevice: executor(
    gql`
      mutation Mutation($clusterName: String!, $vpnDevice: VPNDeviceIn!) {
        infra_updateVPNDevice(
          clusterName: $clusterName
          vpnDevice: $vpnDevice
        ) {
          metadata {
            name
          }
          spec {
            deviceNamespace
            cnameRecords {
              target
              host
            }
            ports {
              targetPort
              port
            }
          }
          status {
            message {
              RawMessage
            }
            isReady
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.infra_updateVPNDevice,
      vars: (_: any) => {},
    }
  ),
  cli_listDevices: executor(
    gql`
      query Infra_listVPNDevices(
        $pq: CursorPaginationIn
        $clusterName: String
      ) {
        infra_listVPNDevices(pq: $pq, clusterName: $clusterName) {
          edges {
            node {
              displayName
              markedForDeletion
              metadata {
                name
                namespace
              }
              spec {
                cnameRecords {
                  host
                  target
                }
                deviceNamespace
                ports {
                  port
                  targetPort
                }
              }
              status {
                isReady
                message {
                  RawMessage
                }
              }
            }
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.infra_listVPNDevices,
      vars: (_: any) => {},
    }
  ),

  cli_getDevice: executor(
    gql`
      query Infra_getVPNDevice($clusterName: String!, $name: String!) {
        infra_getVPNDevice(clusterName: $clusterName, name: $name) {
          displayName
          markedForDeletion
          metadata {
            name
            namespace
          }
          spec {
            cnameRecords {
              host
              target
            }
            deviceNamespace
            nodeSelector
            ports {
              port
              targetPort
            }
          }
          status {
            isReady
            message {
              RawMessage
            }
          }
          wireguardConfig {
            encoding
            value
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.infra_getVPNDevice,
      vars: (_: any) => {},
    }
  ),

  cli_listEnvironments: executor(
    gql`
      query Core_listEnvironments(
        $projectName: String!
        $pq: CursorPaginationIn
      ) {
        core_listEnvironments(projectName: $projectName, pq: $pq) {
          edges {
            cursor
            node {
              displayName
              markedForDeletion
              metadata {
                name
                namespace
              }
              spec {
                projectName
                targetNamespace
              }
              status {
                isReady
                message {
                  RawMessage
                }
              }
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
      transformer: (data: any) => data.core_listEnvironments,
      vars: (_: any) => {},
    }
  ),

  cli_listProjects: executor(
    gql`
      query Core_listProjects($clusterName: String, $pq: CursorPaginationIn) {
        core_listProjects(clusterName: $clusterName, pq: $pq) {
          edges {
            node {
              displayName
              markedForDeletion
              metadata {
                name
                namespace
              }
              status {
                isReady
                message {
                  RawMessage
                }
              }
            }
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.core_listProjects,
      vars: (_: any) => {},
    }
  ),

  cli_getKubeConfig: executor(
    gql`
      query Infra_getCluster($name: String!) {
        infra_getCluster(name: $name) {
          adminKubeconfig {
            encoding
            value
          }
          status {
            isReady
          }
        }
      }
    `,
    {
      transformer: (data: AuthCli_GetKubeConfigQuery) => data.infra_getCluster,
      vars(_: AuthCli_GetKubeConfigQueryVariables) {},
    }
  ),
  cli_listClusters: executor(
    gql`
      query Node($pagination: CursorPaginationIn) {
        infra_listClusters(pagination: $pagination) {
          edges {
            node {
              displayName
              metadata {
                name
              }
              status {
                isReady
              }
            }
          }
        }
      }
    `,
    {
      transformer(data: AuthCli_ListClustersQuery) {
        return data.infra_listClusters;
      },
      vars(_: AuthCli_ListClustersQueryVariables) {},
    }
  ),
  cli_listAccounts: executor(
    gql`
      query Accounts_listAccounts {
        accounts_listAccounts {
          metadata {
            name
          }
          displayName
        }
      }
    `,
    {
      transformer(data: AuthCli_ListAccountsQuery) {
        return data.accounts_listAccounts;
      },
      vars(_: AuthCli_ListAccountsQueryVariables) {},
    }
  ),
  cli_getCurrentUser: executor(
    gql`
      query Auth_me {
        auth_me {
          id
          email
          name
        }
      }
    `,
    {
      transformer(data: AuthCli_GetCurrentUserQuery) {
        return data.auth_me;
      },
      vars(_: AuthCli_GetCurrentUserQueryVariables) {},
    }
  ),

  cli_createRemoteLogin: executor(
    gql`
      mutation Auth_createRemoteLogin($secret: String) {
        auth_createRemoteLogin(secret: $secret)
      }
    `,
    {
      transformer: (data: AuthCli_CreateRemoteLoginMutation) =>
        data.auth_createRemoteLogin,
      vars(_: AuthCli_CreateRemoteLoginMutationVariables) {},
    }
  ),

  cli_getRemoteLogin: executor(
    gql`
      query Auth_getRemoteLogin($loginId: String!, $secret: String!) {
        auth_getRemoteLogin(loginId: $loginId, secret: $secret) {
          authHeader
          status
        }
      }
    `,
    {
      transformer: (data: AuthCli_GetRemoteLoginQuery) =>
        data.auth_getRemoteLogin,
      vars(_: AuthCli_GetRemoteLoginQueryVariables) {},
    }
  ),
});
