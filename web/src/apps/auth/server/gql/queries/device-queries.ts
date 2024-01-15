import gql from 'graphql-tag';
import { IExecutor } from '~/root/lib/server/helpers/execute-query-with-context';

export const vpnQueries = (executor: IExecutor) => ({
  cli_CoreUpdateDevicePorts: executor(
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
  cli_CoreUpdateDeviceEnv: executor(
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

  cli_listCoreDevices: executor(
    gql`
      query Core_listVPNDevicesForUser {
        core_listVPNDevicesForUser {
          displayName
          environmentName
          metadata {
            name
          }
          projectName
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
          metadata {
            name
          }
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

  cli_listDevices: executor(
    gql`
      query Infra_listVPNDevices($pq: CursorPaginationIn) {
        infra_listVPNDevices(pq: $pq) {
          edges {
            node {
              displayName
              metadata {
                name
              }
              spec {
                deviceNamespace
                disabled
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
        }
      }
    `,
    {
      transformer: (data: any) => data.infra_listVPNDevices,
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
  cli_createDevice: executor(
    gql`
      mutation Infra_createVPNDevice(
        $clusterName: String!
        $vpnDevice: VPNDeviceIn!
      ) {
        infra_createVPNDevice(
          clusterName: $clusterName
          vpnDevice: $vpnDevice
        ) {
          metadata {
            name
          }
        }
      }
    `,
    {
      transformer: (data: any) => data.infra_createVPNDevice,
      vars: (_: any) => {},
    }
  ),
});
