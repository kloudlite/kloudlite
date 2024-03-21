import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleListConsoleVpnDevicesForUserQuery,
  ConsoleListConsoleVpnDevicesForUserQueryVariables,
  ConsoleGetConsoleVpnDeviceQuery,
  ConsoleGetConsoleVpnDeviceQueryVariables,
  ConsoleListConsoleVpnDevicesQuery,
  ConsoleListConsoleVpnDevicesQueryVariables,
  ConsoleCreateConsoleVpnDeviceMutation,
  ConsoleCreateConsoleVpnDeviceMutationVariables,
  ConsoleUpdateConsoleVpnDeviceMutation,
  ConsoleUpdateConsoleVpnDeviceMutationVariables,
  ConsoleDeleteConsoleVpnDeviceMutation,
  ConsoleDeleteConsoleVpnDeviceMutationVariables,
} from '~/root/src/generated/gql/server';

export type IConsoleDevices = NN<
  ConsoleListConsoleVpnDevicesQuery['core_listVPNDevices']
>;

export type IConsoleDevicesForUser = NN<
  ConsoleListConsoleVpnDevicesForUserQuery['core_listVPNDevicesForUser']
>;

export const consoleVpnQueries = (executor: IExecutor) => ({
  createConsoleVpnDevice: executor(
    gql`
      mutation Core_createVPNDevice($vpnDevice: ConsoleVPNDeviceIn!) {
        core_createVPNDevice(vpnDevice: $vpnDevice) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateConsoleVpnDeviceMutation) =>
        data.core_createVPNDevice,
      vars(_: ConsoleCreateConsoleVpnDeviceMutationVariables) {},
    }
  ),

  updateConsoleVpnDevice: executor(
    gql`
      mutation Core_updateVPNDevice($vpnDevice: ConsoleVPNDeviceIn!) {
        core_updateVPNDevice(vpnDevice: $vpnDevice) {
          id
        }
      }
    `,
    {
      transformer: (v: ConsoleUpdateConsoleVpnDeviceMutation) => {
        return v.core_updateVPNDevice;
      },
      vars(_: ConsoleUpdateConsoleVpnDeviceMutationVariables) {},
    }
  ),
  listConsoleVpnDevices: executor(
    gql`
      query Core_listVPNDevices(
        $search: CoreSearchVPNDevices
        $pq: CursorPaginationIn
      ) {
        core_listVPNDevices(search: $search, pq: $pq) {
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
              environmentName
              projectName
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
              syncStatus {
                action
                error
                lastSyncedAt
                recordVersion
                state
                syncScheduledAt
              }
              spec {
                cnameRecords {
                  host
                  target
                }
                activeNamespace
                disabled
                nodeSelector
                ports {
                  port
                  targetPort
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
      transformer(data: ConsoleListConsoleVpnDevicesQuery) {
        return data.core_listVPNDevices;
      },
      vars(_: ConsoleListConsoleVpnDevicesQueryVariables) {},
    }
  ),
  getConsoleVpnDevice: executor(
    gql`
      query Core_getVPNDevice($name: String!) {
        core_getVPNDevice(name: $name) {
          displayName
          environmentName
          metadata {
            name
            namespace
          }
          projectName
          recordVersion
          spec {
            cnameRecords {
              host
              target
            }
            activeNamespace
            disabled
            nodeSelector
            ports {
              port
              targetPort
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
      transformer(data: ConsoleGetConsoleVpnDeviceQuery) {
        return data.core_getVPNDevice;
      },
      vars(_: ConsoleGetConsoleVpnDeviceQueryVariables) {},
    }
  ),
  listConsoleVpnDevicesForUser: executor(
    gql`
      query Core_listVPNDevicesForUser {
        core_listVPNDevicesForUser {
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
          environmentName
          projectName
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
          syncStatus {
            action
            error
            lastSyncedAt
            recordVersion
            state
            syncScheduledAt
          }
          spec {
            cnameRecords {
              host
              target
            }
            activeNamespace
            disabled
            nodeSelector
            ports {
              port
              targetPort
            }
          }
          updateTime
        }
      }
    `,
    {
      transformer(data: ConsoleListConsoleVpnDevicesForUserQuery) {
        return data.core_listVPNDevicesForUser;
      },
      vars(_: ConsoleListConsoleVpnDevicesForUserQueryVariables) {},
    }
  ),
  deleteConsoleVpnDevice: executor(
    gql`
      mutation Core_deleteVPNDevice($deviceName: String!) {
        core_deleteVPNDevice(deviceName: $deviceName)
      }
    `,
    {
      transformer(data: ConsoleDeleteConsoleVpnDeviceMutation) {
        return data.core_deleteVPNDevice;
      },
      vars(_: ConsoleDeleteConsoleVpnDeviceMutationVariables) {},
    }
  ),
});
