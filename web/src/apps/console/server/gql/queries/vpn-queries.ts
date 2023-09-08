import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import {
  ConsoleCreateVpnDeviceMutation,
  ConsoleCreateVpnDeviceMutationVariables,
  ConsoleGetVpnDeviceQuery,
  ConsoleGetVpnDeviceQueryVariables,
  ConsoleListVpnDevicesQuery,
  ConsoleListVpnDevicesQueryVariables,
  ConsoleUpdateVpnDeviceMutation,
  ConsoleUpdateVpnDeviceMutationVariables,
} from '~/root/src/generated/gql/server';

export const vpnQueries = (executor: IExecutor) => ({
  createVpnDevice: executor(
    gql`
      mutation Core_createVPNDevice($vpnDevice: VPNDeviceIn!) {
        core_createVPNDevice(vpnDevice: $vpnDevice) {
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
      mutation Core_createVPNDevice($vpnDevice: VPNDeviceIn!) {
        core_updateVPNDevice(vpnDevice: $vpnDevice) {
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
      query Query($search: SearchVPNDevices, $pq: CursorPaginationIn) {
        core_listVPNDevices(search: $search, pq: $pq) {
          edges {
            cursor
            node {
              metadata {
                name
              }
              clusterName
              displayName

              spec {
                serverName
                ports {
                  port
                  targetPort
                }
                offset
              }
              createdBy {
                userId
                userName
                userEmail
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
      transformer(data: ConsoleListVpnDevicesQuery) {
        return data.core_listVPNDevices;
      },
      vars(_: ConsoleListVpnDevicesQueryVariables) {},
    }
  ),
  getVpnDevice: executor(
    gql`
      query Core_getVPNDevice($name: String!) {
        core_getVPNDevice(name: $name) {
          updateTime
          spec {
            serverName
            ports {
              port
              targetPort
            }
            offset
          }
          clusterName
          displayName
          metadata {
            name
          }
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
});
