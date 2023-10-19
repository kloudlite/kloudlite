import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateVpnDeviceMutation,
  ConsoleCreateVpnDeviceMutationVariables,
  ConsoleDeleteVpnDeviceMutation,
  ConsoleDeleteVpnDeviceMutationVariables,
  ConsoleGetVpnDeviceQuery,
  ConsoleGetVpnDeviceQueryVariables,
  ConsoleListVpnDevicesQuery,
  ConsoleListVpnDevicesQueryVariables,
  ConsoleUpdateVpnDeviceMutation,
  ConsoleUpdateVpnDeviceMutationVariables,
} from '~/root/src/generated/gql/server';

export type IDevices = NN<ConsoleListVpnDevicesQuery['core_listVPNDevices']>;

export const vpnQueries = (executor: IExecutor) => ({
  createVpnDevice: executor(
    gql`
      mutation Mutation($clusterName: String!, $vpnDevice: VPNDeviceIn!) {
        core_createVPNDevice(clusterName: $clusterName, vpnDevice: $vpnDevice) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateVpnDeviceMutation) =>
        data.core_createVPNDevice,
      vars(_: ConsoleCreateVpnDeviceMutationVariables) {},
    }
  ),

  updateVpnDevice: executor(
    gql`
      mutation Core_updateVPNDevice(
        $clusterName: String!
        $vpnDevice: VPNDeviceIn!
      ) {
        core_updateVPNDevice(clusterName: $clusterName, vpnDevice: $vpnDevice) {
          id
        }
      }
    `,
    {
      transformer: (v: ConsoleUpdateVpnDeviceMutation) => {
        return v.core_updateVPNDevice;
      },
      vars(_: ConsoleUpdateVpnDeviceMutationVariables) {},
    }
  ),
  listVpnDevices: executor(
    gql`
      query Core_listVPNDevices(
        $clusterName: String
        $search: SearchVPNDevices
        $pq: CursorPaginationIn
      ) {
        core_listVPNDevices(
          clusterName: $clusterName
          search: $search
          pq: $pq
        ) {
          edges {
            cursor
            node {
              accountName
              apiVersion
              clusterName
              createdBy {
                userEmail
                userId
                userName
              }
              creationTime
              displayName
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
                offset
                ports {
                  port
                  targetPort
                }
                serverName
              }
              status {
                checks
                isReady
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
      transformer(data: ConsoleListVpnDevicesQuery) {
        return data.core_listVPNDevices;
      },
      vars(_: ConsoleListVpnDevicesQueryVariables) {},
    }
  ),
  getVpnDevice: executor(
    gql`
      query Query($clusterName: String!, $name: String!) {
        core_getVPNDevice(clusterName: $clusterName, name: $name) {
          accountName
          apiVersion
          clusterName
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
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
            offset
            ports {
              port
              targetPort
            }
            serverName
          }
          status {
            checks
            isReady
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
      transformer(data: ConsoleGetVpnDeviceQuery) {
        return data.core_getVPNDevice;
      },
      vars(_: ConsoleGetVpnDeviceQueryVariables) {},
    }
  ),
  deleteVpnDevice: executor(
    gql`
      mutation Mutation($clusterName: String!, $deviceName: String!) {
        core_deleteVPNDevice(clusterName: $clusterName, deviceName: $deviceName)
      }
    `,
    {
      transformer(data: ConsoleDeleteVpnDeviceMutation) {
        return data.core_deleteVPNDevice;
      },
      vars(_: ConsoleDeleteVpnDeviceMutationVariables) {},
    }
  ),
});
