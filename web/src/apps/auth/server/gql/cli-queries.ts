/* eslint-disable camelcase */
import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { vpnQueries } from './queries/device-queries';

export const cliQueries = (executor: IExecutor) => ({
  ...vpnQueries(executor),

  cli_createGlobalVPNDevice: executor(
    gql`
      mutation Infra_createGlobalVPNDevice($gvpnDevice: GlobalVPNDeviceIn!) {
        infra_createGlobalVPNDevice(gvpnDevice: $gvpnDevice) {
          accountName
          creationTime
          createdBy {
            userEmail
            userId
            userName
          }
          displayName
          globalVPNName
          id
          ipAddr
          lastUpdatedBy {
            userName
            userId
            userEmail
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
          privateKey
          publicKey
          recordVersion
          updateTime
          wireguardConfig {
            value
            encoding
          }
        }
      }
    `,
    {
      transformer(data: any) {
        return data.infra_createGlobalVPNDevice;
      },
      vars(_: any) {},
    }
  ),

  cli_getMresOutputKeyValues: executor(
    gql`
      query Core_getManagedResouceOutputKeyValues(
        $envName: String!
        $keyrefs: [ManagedResourceKeyRefIn]
      ) {
        core_getManagedResouceOutputKeyValues(
          envName: $envName
          keyrefs: $keyrefs
        ) {
          key
          mresName
          value
        }
      }
    `,
    {
      transformer: (data: any) => data.core_getManagedResouceOutputKeyValues,
      vars: (_: any) => {},
    }
  ),

  cli_getGlobalVpnDevice: executor(
    gql`
      query Infra_getGlobalVPNDevice($gvpn: String!, $deviceName: String!) {
        infra_getGlobalVPNDevice(gvpn: $gvpn, deviceName: $deviceName) {
          accountName
          creationTime
          createdBy {
            userEmail
            userId
            userName
          }
          displayName
          globalVPNName
          id
          ipAddr
          lastUpdatedBy {
            userName
            userId
            userEmail
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
          privateKey
          publicKey
          recordVersion
          updateTime
          wireguardConfig {
            value
            encoding
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.infra_getGlobalVPNDevice,
      vars: (_: any) => {},
    }
  ),

  cli_coreCheckNameAvailability: executor(
    gql`
      query Core_checkNameAvailability(
        $resType: ConsoleResType!
        $name: String!
      ) {
        core_checkNameAvailability(resType: $resType, name: $name) {
          result
          suggestedNames
        }
      }
    `,
    {
      transformer: (data: any) => data.core_checkNameAvailability,
      vars: (_: any) => {},
    }
  ),

  cli_getMresKeys: executor(
    gql`
      query Core_getManagedResouceOutputKeyValues(
        $envName: String!
        $name: String!
      ) {
        core_getManagedResouceOutputKeys(envName: $envName, name: $name)
      }
    `,
    {
      transformer: (data: any) => data.core_getManagedResouceOutputKeys,
      vars: (_: any) => {},
    }
  ),

  cli_listMreses: executor(
    gql`
      query Core_listManagedResources(
        $envName: String!
        $pq: CursorPaginationIn
      ) {
        core_listManagedResources(envName: $envName, pq: $pq) {
          edges {
            node {
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
      transformer: (data: any) => data.core_listManagedResources,
      vars: (_: any) => {},
    }
  ),

  cli_getMresConfigsValues: executor(
    gql`
      query Core_getManagedResouceOutputKeyValues(
        $keyrefs: [ManagedResourceKeyRefIn]
        $envName: String!
      ) {
        core_getManagedResouceOutputKeyValues(
          keyrefs: $keyrefs
          envName: $envName
        ) {
          key
          mresName
          value
        }
      }
    `,
    {
      transformer: (data: any) => data,
      vars: (_: any) => {},
    }
  ),

  cli_infraCheckNameAvailability: executor(
    gql`
      query Infra_checkNameAvailability(
        $resType: ResType!
        $name: String!
        $clusterName: String
      ) {
        infra_checkNameAvailability(
          resType: $resType
          name: $name
          clusterName: $clusterName
        ) {
          result
          suggestedNames
        }
      }
    `,
    {
      transformer: (data: any) => data.infra_checkNameAvailability,
      vars: (_: any) => {},
    }
  ),

  cli_getConfigSecretMap: executor(
    gql`
      query Core_getConfigValues(
        $envName: String!
        $configQueries: [ConfigKeyRefIn]
        $secretQueries: [SecretKeyRefIn!]
        $mresQueries: [ManagedResourceKeyRefIn]
      ) {
        configs: core_getConfigValues(
          envName: $envName
          queries: $configQueries
        ) {
          configName
          key
          value
        }
        secrets: core_getSecretValues(
          envName: $envName
          queries: $secretQueries
        ) {
          key
          secretName
          value
        }
        mreses: core_getManagedResouceOutputKeyValues(
          keyrefs: $mresQueries
          envName: $envName
        ) {
          key
          mresName
          value
        }
      }
    `,
    {
      transformer: (data: any) => {
        return {
          configs: data.configs,
          secrets: data.secrets,
          mreses: data.mreses,
        };
      },
      vars: (_: any) => {},
    }
  ),
  cli_interceptApp: executor(
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
      transformer: (data: any) => data.core_interceptApp,
      vars: (_: any) => {},
    }
  ),
  cli_getEnvironment: executor(
    gql`
      query Core_getEnvironment($name: String!) {
        core_getEnvironment(name: $name) {
          status {
            isReady
            message {
              RawMessage
            }
          }
          metadata {
            name
          }
          displayName
          clusterName
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
  cli_getSecret: executor(
    gql`
      query Core_getSecret($envName: String!, $name: String!) {
        core_getSecret(envName: $envName, name: $name) {
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
      query Core_getConfig($envName: String!, $name: String!) {
        core_getConfig(envName: $envName, name: $name) {
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
      query Core_listApps($envName: String!) {
        core_listApps(envName: $envName) {
          edges {
            cursor
            node {
              displayName
              environmentName
              markedForDeletion
              metadata {
                annotations
                name
                namespace
              }
              spec {
                displayName
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
                  name
                }
                intercept {
                  enabled
                  toDevice
                  portMappings {
                    appPort
                    devicePort
                  }
                }
                nodeSelector
                replicas
                serviceAccount
                services {
                  port
                }
              }
              status {
                checks
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
      transformer: (data: any) => data.core_listApps,
      vars: (_: any) => {},
    }
  ),
  cli_listConfigs: executor(
    gql`
      query Core_listConfigs($envName: String!) {
        core_listConfigs(envName: $envName) {
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
      query Core_listSecrets($envName: String!, $pq: CursorPaginationIn) {
        core_listSecrets(envName: $envName, pq: $pq) {
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

  cli_listEnvironments: executor(
    gql`
      query Core_listEnvironments($pq: CursorPaginationIn) {
        core_listEnvironments(pq: $pq) {
          edges {
            cursor
            node {
              displayName
              markedForDeletion
              clusterName
              metadata {
                name
                namespace
              }
              spec {
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
      transformer: (data: any) => data.infra_getCluster,
      vars: (_: any) => {},
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
      transformer: (data: any) => data.infra_listClusters,
      vars: (_: any) => {},
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
      transformer: (data: any) => data.accounts_listAccounts,
      vars: (_: any) => {},
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
      transformer: (data: any) => data.auth_me,
      vars: (_: any) => {},
    }
  ),

  cli_createRemoteLogin: executor(
    gql`
      mutation Auth_createRemoteLogin($secret: String) {
        auth_createRemoteLogin(secret: $secret)
      }
    `,
    {
      transformer: (data: any) => data.auth_createRemoteLogin,
      vars: (_: any) => {},
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
      transformer: (data: any) => data.auth_getRemoteLogin,
      vars: (_: any) => {},
    }
  ),
});
