import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';
import { NN } from '~/root/lib/types/common';
import {
  IotconsoleCreateIotDeviceMutation,
  IotconsoleCreateIotDeviceMutationVariables,
  IotconsoleGetIotDeviceQuery,
  IotconsoleGetIotDeviceQueryVariables,
  IotconsoleListIotDevicesQuery,
  IotconsoleListIotDevicesQueryVariables,
  IotconsoleUpdateIotDeviceMutation,
  IotconsoleUpdateIotDeviceMutationVariables,
  IotconsoleDeleteIotDeviceMutation,
  IotconsoleDeleteIotDeviceMutationVariables,
} from '~/root/src/generated/gql/server';

export type IDevices = NN<IotconsoleListIotDevicesQuery['iot_listDevices']>;
export type IDevice = NN<IotconsoleGetIotDeviceQuery['iot_getDevice']>;

export const iotDeviceQueries = (executor: IExecutor) => ({
  deleteIotDevice: executor(
    gql`
      mutation Iot_deleteDevice(
        $projectName: String!
        $deploymentName: String!
        $name: String!
      ) {
        iot_deleteDevice(
          projectName: $projectName
          deploymentName: $deploymentName
          name: $name
        )
      }
    `,
    {
      transformer: (data: IotconsoleDeleteIotDeviceMutation) =>
        data.iot_deleteDevice,
      vars(_: IotconsoleDeleteIotDeviceMutationVariables) {},
    }
  ),
  createIotDevice: executor(
    gql`
      mutation Iot_createDevice(
        $deploymentName: String!
        $device: IOTDeviceIn!
        $projectName: String!
      ) {
        iot_createDevice(
          projectName: $projectName
          deploymentName: $deploymentName
          device: $device
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: IotconsoleCreateIotDeviceMutation) =>
        data.iot_createDevice,
      vars(_: IotconsoleCreateIotDeviceMutationVariables) {},
    }
  ),
  updateIotDevice: executor(
    gql`
      mutation Iot_updateDevice(
        $deploymentName: String!
        $device: IOTDeviceIn!
        $projectName: String!
      ) {
        iot_updateDevice(
          projectName: $projectName
          deploymentName: $deploymentName
          device: $device
        ) {
          id
        }
      }
    `,
    {
      transformer: (data: IotconsoleUpdateIotDeviceMutation) =>
        data.iot_updateDevice,
      vars(_: IotconsoleUpdateIotDeviceMutationVariables) {},
    }
  ),
  getIotDevice: executor(
    gql`
      query Iot_getDevice(
        $projectName: String!
        $deploymentName: String!
        $name: String!
      ) {
        iot_getDevice(
          projectName: $projectName
          deploymentName: $deploymentName
          name: $name
        ) {
          accountName
          createdBy {
            userEmail
            userId
            userName
          }
          creationTime
          deploymentName
          displayName
          id
          ip
          lastUpdatedBy {
            userEmail
            userId
            userName
          }
          markedForDeletion
          name
          podCIDR
          publicKey
          recordVersion
          serviceCIDR
          updateTime
          version
        }
      }
    `,
    {
      transformer: (data: IotconsoleGetIotDeviceQuery) => data.iot_getDevice,
      vars(_: IotconsoleGetIotDeviceQueryVariables) {},
    }
  ),
  listIotDevices: executor(
    gql`
      query Iot_listDevices(
        $deploymentName: String!
        $search: SearchIOTDevices
        $pq: CursorPaginationIn
        $projectName: String!
      ) {
        iot_listDevices(
          projectName: $projectName
          deploymentName: $deploymentName
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
              deploymentName
              displayName
              id
              ip
              lastUpdatedBy {
                userEmail
                userId
                userName
              }
              markedForDeletion
              name
              podCIDR
              publicKey
              recordVersion
              serviceCIDR
              updateTime
              version
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
      transformer: (data: IotconsoleListIotDevicesQuery) => {
        return data.iot_listDevices;
      },
      vars(_: IotconsoleListIotDevicesQueryVariables) {},
    }
  ),
});
