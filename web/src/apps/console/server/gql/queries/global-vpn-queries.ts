import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  ConsoleCreateGlobalVpnDeviceMutation,
  ConsoleCreateGlobalVpnDeviceMutationVariables,
  ConsoleGetGlobalVpnDeviceQuery,
  ConsoleGetGlobalVpnDeviceQueryVariables,
  ConsoleListGlobalVpnDevicesQuery,
  ConsoleListGlobalVpnDevicesQueryVariables,
  ConsoleUpdateGlobalVpnDeviceMutation,
  ConsoleUpdateGlobalVpnDeviceMutationVariables,
  ConsoleDeleteGlobalVpnDeviceMutation,
  ConsoleDeleteGlobalVpnDeviceMutationVariables,
} from '~/root/src/generated/gql/server';

export type IGlobalVpnDevices = NN<
  ConsoleListGlobalVpnDevicesQuery['infra_listGlobalVPNDevices']
>;

export const globalVpnQueries = (executor: IExecutor) => ({
  deleteGlobalVpnDevice: executor(
    gql`
      mutation Infra_deleteGlobalVPNDevice(
        $gvpn: String!
        $deviceName: String!
      ) {
        infra_deleteGlobalVPNDevice(gvpn: $gvpn, deviceName: $deviceName)
      }
    `,
    {
      transformer: (data: ConsoleDeleteGlobalVpnDeviceMutation) =>
        data.infra_deleteGlobalVPNDevice,
      vars(_: ConsoleDeleteGlobalVpnDeviceMutationVariables) {},
    }
  ),
  createGlobalVpnDevice: executor(
    gql`
      mutation Infra_createGlobalVPNDevice($gvpnDevice: GlobalVPNDeviceIn!) {
        infra_createGlobalVPNDevice(gvpnDevice: $gvpnDevice) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleCreateGlobalVpnDeviceMutation) =>
        data.infra_createGlobalVPNDevice,
      vars(_: ConsoleCreateGlobalVpnDeviceMutationVariables) {},
    }
  ),
  updateGlobalVpnDevice: executor(
    gql`
      mutation Infra_updateGlobalVPNDevice($gvpnDevice: GlobalVPNDeviceIn!) {
        infra_updateGlobalVPNDevice(gvpnDevice: $gvpnDevice) {
          id
        }
      }
    `,
    {
      transformer: (data: ConsoleUpdateGlobalVpnDeviceMutation) =>
        data.infra_updateGlobalVPNDevice,
      vars(_: ConsoleUpdateGlobalVpnDeviceMutationVariables) {},
    }
  ),
  getGlobalVpnDevice: executor(
    gql`
      query Infra_getGlobalVPNDevice($gvpn: String!, $deviceName: String!) {
        infra_getGlobalVPNDevice(gvpn: $gvpn, deviceName: $deviceName) {
          accountName
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          displayName
          globalVPNName
          id
          ipAddr
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
      transformer: (data: ConsoleGetGlobalVpnDeviceQuery) =>
        data.infra_getGlobalVPNDevice,
      vars(_: ConsoleGetGlobalVpnDeviceQueryVariables) {},
    }
  ),
  listGlobalVpnDevices: executor(
    gql`
      query Infra_listGlobalVPNDevices(
        $gvpn: String!
        $search: SearchGlobalVPNDevices
        $pagination: CursorPaginationIn
      ) {
        infra_listGlobalVPNDevices(
          gvpn: $gvpn
          search: $search
          pagination: $pagination
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
              creationMethod
              creationTime
              displayName
              globalVPNName
              id
              ipAddr
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
              privateKey
              publicKey
              recordVersion
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
      transformer: (data: ConsoleListGlobalVpnDevicesQuery) =>
        data.infra_listGlobalVPNDevices,
      vars(_: ConsoleListGlobalVpnDevicesQueryVariables) {},
    }
  ),
});
