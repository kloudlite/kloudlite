/* eslint-disable camelcase */
import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { AuthCli_ListAppsQuery } from '~/root/src/generated/gql/server';

export const cliQueries = (executor: IExecutor) => ({
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
        $msvcName: String!
        $keyrefs: [ManagedResourceKeyRefIn]
      ) {
        core_getManagedResouceOutputKeyValues(
          msvcName: $msvcName
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
        $name: String!
        $envName: String
      ) {
        core_getManagedResouceOutputKeys(name: $name, envName: $envName)
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
        $pq: CursorPaginationIn
        $search: SearchManagedResources
      ) {
        core_listManagedResources(pq: $pq, search: $search) {
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
        $envName: String
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
        $configQueries: [ConfigKeyRefIn!]
        $secretQueries: [SecretKeyRefIn!]
        $mresQueries: [SecretKeyRefIn!]
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
        mreses: core_getSecretValues(envName: $envName, queries: $mresQueries) {
          key
          secretName
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

  cli_intercepExternalApp: executor(
    gql`
      mutation Core_interceptExternalApp(
        $envName: String!
        $appName: String!
        $deviceName: String!
        $intercept: Boolean!
        $portMappings: [Github__com___kloudlite___operator___apis___crds___v1__AppInterceptPortMappingsIn!]
      ) {
        core_interceptExternalApp(
          envName: $envName
          externalAppName: $appName
          deviceName: $deviceName
          intercept: $intercept
          portMappings: $portMappings
        )
      }
    `,
    {
      transformer: (data: any) => data.core_interceptExternalApp,
      vars: (_: any) => {},
    }
  ),
  cli_interceptApp: executor(
    gql`
      mutation Core_interceptApp(
        $portMappings: [Github__com___kloudlite___operator___apis___crds___v1__AppInterceptPortMappingsIn!]
        $intercept: Boolean!
        $deviceName: String!
        $appName: String!
        $envName: String!
      ) {
        core_interceptApp(
          portMappings: $portMappings
          intercept: $intercept
          deviceName: $deviceName
          appname: $appName
          envName: $envName
        )
      }
    `,
    {
      transformer: (data: any) => data.core_interceptApp,
      vars: (_: any) => {},
    }
  ),
  cli_removeDeviceIntercepts: executor(
    gql`
      mutation Core_removeDeviceIntercepts(
        $envName: String!
        $deviceName: String!
      ) {
        core_removeDeviceIntercepts(envName: $envName, deviceName: $deviceName)
      }
    `,
    {
      transformer: (data: any) => data.core_removeDeviceIntercepts,
      vars(_: any) {},
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
  cli_cloneEnvironment: executor(
    gql`
      mutation Core_cloneEnvironment(
        $clusterName: String!
        $sourceEnvName: String!
        $destinationEnvName: String!
        $displayName: String!
        $environmentRoutingMode: Github__com___kloudlite___operator___apis___crds___v1__EnvironmentRoutingMode!
      ) {
        core_cloneEnvironment(
          clusterName: $clusterName
          sourceEnvName: $sourceEnvName
          destinationEnvName: $destinationEnvName
          displayName: $displayName
          environmentRoutingMode: $environmentRoutingMode
        ) {
          id
          displayName
          clusterName
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
          spec {
            targetNamespace
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.core_cloneEnvironment,
      vars(_: any) {},
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
      query Core_listApps($pq: CursorPaginationIn, $envName: String!) {
        apps: core_listExternalApps(pq: $pq, envName: $envName) {
          edges {
            node {
              spec {
                intercept {
                  enabled
                  portMappings {
                    devicePort
                    appPort
                  }
                  toDevice
                }
              }
              displayName
              environmentName
              markedForDeletion
              metadata {
                name
                annotations
                namespace
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
        mapps: core_listApps(pq: $pq, envName: $envName) {
          edges {
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
                intercept {
                  enabled
                  toDevice
                  portMappings {
                    devicePort
                    appPort
                  }
                }
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
      transformer: (data: AuthCli_ListAppsQuery) => {
        if (data.apps) {
          data.apps.edges = data.apps.edges.map((edge) => ({
            node: {
              ...edge.node,
              mapp: false,
            },
          }));
        }

        if (data.mapps) {
          data.mapps.edges = data.mapps.edges.map((edge) => ({
            node: {
              ...edge.node,
              mapp: true,
            },
          }));
        }
        data.apps?.edges.push(...(data.mapps?.edges || []));
        return data.apps;
      },
      vars: (_: any) => {},
    }
  ),
  cli_listConfigs: executor(
    gql`
      query Core_listConfigs($pq: CursorPaginationIn, $envName: String!) {
        core_listConfigs(pq: $pq, envName: $envName) {
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
              isReadyOnly
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
            hasPrevPage
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
  cli_createClusterReference: executor(
    gql`
      mutation Infra_createBYOKCluster($cluster: BYOKClusterIn!) {
        infra_createBYOKCluster(cluster: $cluster) {
          id
          clusterToken
          displayName
          metadata {
            name
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.infra_createBYOKCluster,
      vars(_: any) {},
    }
  ),
  cli_deleteClusterReference: executor(
    gql`
      mutation Infra_deleteBYOKCluster($name: String!) {
        infra_deleteBYOKCluster(name: $name)
      }
    `,
    {
      transformer: (data: any) => data.infra_deleteBYOKCluster,
      vars(_: any) {},
    }
  ),
  cli_clusterReferenceInstructions: executor(
    gql`
      query Infrat_getBYOKClusterSetupInstructions($name: String!) {
        infrat_getBYOKClusterSetupInstructions(
          name: $name
          onlyHelmValues: true
        ) {
          command
          title
        }
      }
    `,
    {
      transformer: (data: any) => {
        const instructions = JSON.parse(
          data.infrat_getBYOKClusterSetupInstructions[0].command
        );
        return instructions;
      },
      vars(_: any) {},
    }
  ),
  cli_listImportedManagedResources: executor(
    gql`
      query Core_listImportedManagedResources(
        $envName: String!
        $search: SearchImportedManagedResources
        $pq: CursorPaginationIn
      ) {
        core_listImportedManagedResources(
          envName: $envName
          search: $search
          pq: $pq
        ) {
          edges {
            cursor
            node {
              accountName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
              environmentName
              id
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              managedResourceRef {
                id
                name
                namespace
              }
              markedForDeletion
              name
              recordVersion
              secretRef {
                name
                namespace
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
              managedResource {
                accountName
                apiVersion
                creationTime
                displayName
                enabled
                environmentName
                id
                isImported
                kind
                managedServiceName
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
                mresRef
                recordVersion
                spec {
                  resourceNamePrefix
                  resourceTemplate {
                    apiVersion
                    kind
                    msvcRef {
                      apiVersion
                      kind
                      name
                      namespace
                    }
                    spec
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
                syncedOutputSecretRef {
                  apiVersion
                  data
                  immutable
                  kind
                  stringData
                  type
                }
                updateTime
              }
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
      transformer: (data: any) => data.core_listImportedManagedResources,
      vars(_: any) {},
    }
  ),
  cli_listByokClusters: executor(
    gql`
      query Infra_listBYOKClusters(
        $search: SearchCluster
        $pagination: CursorPaginationIn
      ) {
        infra_listBYOKClusters(search: $search, pagination: $pagination) {
          edges {
            cursor
            node {
              clusterToken
              displayName
              id
              metadata {
                name
                namespace
              }
              updateTime
            }
          }
          totalCount
        }
      }
    `,
    {
      transformer: (data: any) => {
        return data.infra_listBYOKClusters;
      },
      vars(_: any) {},
    }
  ),
});
