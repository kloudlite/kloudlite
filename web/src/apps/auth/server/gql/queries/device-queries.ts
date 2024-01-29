import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';

export const vpnQueries = (executor: IExecutor) => ({
  cli_updateDeviceCluster: executor(
    gql`
      mutation Core_updateVpnClusterName(
        $deviceName: String!
        $clusterName: String!
      ) {
        core_updateVpnClusterName(
          deviceName: $deviceName
          clusterName: $clusterName
        )
      }
    `,
    {
      transformer: (data: any) => data.core_updateVpnClusterName,
      vars: (_: any) => {},
    }
  ),
  cli_updateDeviceNs: executor(
    gql`
      mutation Core_updateVpnDeviceNs($deviceName: String!, $ns: String!) {
        core_updateVpnDeviceNs(deviceName: $deviceName, ns: $ns)
      }
    `,
    {
      transformer: (data: any) => data.core_updateVpnDeviceNs,
      vars: (_: any) => {},
    }
  ),
  cli_updateDevicePorts: executor(
    gql`
      mutation Core_updateVPNDevicePorts(
        $deviceName: String!
        $ports: [PortIn!]!
      ) {
        core_updateVPNDevicePorts(deviceName: $deviceName, ports: $ports)
      }
    `,
    {
      transformer: (data: any) => data.core_updateVPNDevicePorts,
      vars: (_: any) => {},
    }
  ),
  cli_updateDeviceEnv: executor(
    gql`
      mutation Core_updateVPNDeviceEnv(
        $deviceName: String!
        $projectName: String!
        $envName: String!
      ) {
        core_updateVPNDeviceEnv(
          deviceName: $deviceName
          projectName: $projectName
          envName: $envName
        )
      }
    `,
    {
      transformer: (data: any) => data.core_updateVPNDeviceEnv,
      vars: (_: any) => {},
    }
  ),

  cli_listDevices: executor(
    gql`
      query Core_listVPNDevicesForUser {
        core_listVPNDevicesForUser {
          displayName
          environmentName
          metadata {
            name
          }
          projectName
          clusterName
          status {
            isReady
            message {
              RawMessage
            }
          }
          spec {
            cnameRecords {
              host
              target
            }
            activeNamespace
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
  cli_getDevice: executor(
    gql`
      query Core_getVPNDevice($name: String!) {
        core_getVPNDevice(name: $name) {
          displayName
          metadata {
            name
          }
          clusterName
          projectName
          spec {
            activeNamespace
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
  cli_createDevice: executor(
    gql`
      mutation Core_createVPNDevice($vpnDevice: ConsoleVPNDeviceIn!) {
        core_createVPNDevice(vpnDevice: $vpnDevice) {
          metadata {
            name
          }
          wireguardConfig {
            encoding
            value
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.core_createVPNDevice,
      vars: (_: any) => {},
    }
  ),
});
