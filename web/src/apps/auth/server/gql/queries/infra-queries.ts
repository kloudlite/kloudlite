import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';

export const infraQueries = (executor: IExecutor) => ({
  cli_CoreCheckNameAvailability: executor(
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
  cli_listCoreDevices: executor(
    gql`
      query Core_listVPNDevicesForUser {
        core_listVPNDevicesForUser {
          displayName
          environmentName
          markedForDeletion
          metadata {
            name
            namespace
          }
          projectName
          spec {
            cnameRecords {
              host
              target
            }
            deviceNamespace
            disabled
            ports {
              port
              targetPort
            }
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.core_listVPNDevicesForUser,
      vars: (_: any) => {},
    }
  ),
  cli_getCoreDevice: executor(
    gql`
      query Core_getVPNDevice($name: String!) {
        core_getVPNDevice(name: $name) {
          displayName
          metadata {
            name
          }
          projectName
          spec {
            deviceNamespace
            disabled
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
      transformer: (data: any) => data.core_getVPNDevice,
      vars: (_: any) => {},
    }
  ),
  cli_createCoreDevice: executor(
    gql`
      mutation Core_createVPNDevice($vpnDevice: ConsoleVPNDeviceIn!) {
        core_createVPNDevice(vpnDevice: $vpnDevice) {
          id
        }
      }
    `,
    {
      transformer: (data: any) => data.core_createVPNDevice,
      vars: (_: any) => {},
    }
  ),

  cli_updateCoreDevicePorts: executor(
    gql`
      mutation Mutation($deviceName: String!, $ports: [PortIn!]!) {
        core_updateVPNDevicePorts(deviceName: $deviceName, ports: $ports)
      }
    `,
    {
      transformer: (data: any) => data.core_updateVPNDevicePorts,
      vars: (_: any) => {},
    }
  ),
});
