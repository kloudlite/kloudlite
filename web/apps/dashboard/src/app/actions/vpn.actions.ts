'use server'

import { getK8sClient } from '@kloudlite/lib/k8s'
import { workMachineRepository } from '@kloudlite/lib/k8s'
import type { V1Secret, V1Service, V1Ingress, V1Pod } from '@kubernetes/client-node'

export interface HostEntry {
  hostname: string
  ip: string
}

export interface TunnelEndpointInfo {
  tunnel_endpoint: string // hostname:443 for connection
  hostname: string // vpn-connect.{subdomain}.{domain}
  ip: string // WorkMachine public IP
}

/**
 * Get CA certificate for VPN connection
 * Fetches the wildcard TLS certificate CA from kloudlite namespace
 */
export async function getCACert() {
  try {
    const client = getK8sClient()

    // Fetch CA certificate from kloudlite namespace
    const caSecret = await client.core.readNamespacedSecret({
      name: 'kloudlite-wildcard-cert-tls',
      namespace: 'kloudlite',
    }) as V1Secret

    const caCert = caSecret.data?.['ca.crt']
    if (!caCert) {
      return {
        success: false,
        error: 'CA certificate not found in secret',
      }
    }

    // Decode from base64
    const caCertDecoded = Buffer.from(caCert, 'base64').toString('utf-8')

    return {
      success: true,
      data: { ca_cert: caCertDecoded },
    }
  } catch (err) {
    console.error('Get CA cert error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Get hosts entries for a user's WorkMachine namespace
 * Returns list of hostname -> IP mappings for /etc/hosts configuration
 */
export async function getHosts(username: string) {
  try {
    // Find user's WorkMachine to get target namespace
    const workMachine = await workMachineRepository.getByOwner(username)
    if (!workMachine) {
      return {
        success: false,
        error: 'No work machine found for user',
      }
    }

    const targetNamespace = workMachine.spec.targetNamespace
    if (!targetNamespace) {
      return {
        success: false,
        error: 'WorkMachine has no target namespace',
      }
    }

    const client = getK8sClient()
    const hosts: HostEntry[] = []

    // Get all ingresses to build host entries
    const ingressList = await client.custom.listClusterCustomObject({
      group: 'networking.k8s.io',
      version: 'v1',
      plural: 'ingresses',
    }) as { items: V1Ingress[] }

    // Get the ingress controller service IP (wm-ingress-controller)
    let routerIP = ''
    try {
      const routerSvc = await client.core.readNamespacedService({
        name: 'wm-ingress-controller',
        namespace: targetNamespace,
      }) as V1Service

      routerIP = routerSvc.spec?.clusterIP || ''
    } catch (err) {
      // Service might not exist yet, continue without it
      console.warn('Ingress controller service not found:', err)
    }

    // Build host entries from ingress rules
    if (routerIP && ingressList.items) {
      for (const ingress of ingressList.items) {
        if (ingress.spec?.rules) {
          for (const rule of ingress.spec.rules) {
            if (rule.host) {
              hosts.push({
                hostname: rule.host,
                ip: routerIP,
              })
            }
          }
        }
      }
    }

    return {
      success: true,
      data: { hosts },
    }
  } catch (err) {
    console.error('Get hosts error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Get tunnel endpoint for VPN connection
 * Returns the tunnel server hostname and IP for the user's WorkMachine
 */
export async function getTunnelEndpoint(username: string) {
  try {
    // Find user's WorkMachine
    const workMachine = await workMachineRepository.getByOwner(username)
    if (!workMachine) {
      return {
        success: false,
        error: 'No work machine found for user',
      }
    }

    // Get public IP from WorkMachine status
    const publicIP = workMachine.status?.publicIP
    if (!publicIP) {
      return {
        success: false,
        error: 'WorkMachine has no public IP (may not be running)',
      }
    }

    const targetNamespace = workMachine.spec.targetNamespace
    if (!targetNamespace) {
      return {
        success: false,
        error: 'WorkMachine has no target namespace',
      }
    }

    // Check if tunnel server is ready
    const isTunnelReady = await checkTunnelServerReady(targetNamespace)
    if (!isTunnelReady) {
      return {
        success: false,
        error: 'Tunnel server is not ready yet (WorkMachine may still be starting)',
      }
    }

    // Get domain info from environment
    const fullDomain = process.env.HOSTED_SUBDOMAIN
    if (!fullDomain) {
      return {
        success: false,
        error: 'HOSTED_SUBDOMAIN environment variable not set',
      }
    }

    // Parse subdomain and domain (e.g., "beanbag.khost.dev" -> ["beanbag", "khost.dev"])
    const parts = fullDomain.split('.')
    if (parts.length < 2) {
      return {
        success: false,
        error: `Invalid domain format: ${fullDomain} (expected subdomain.domain)`,
      }
    }

    const subdomain = parts[0]
    const domain = parts.slice(1).join('.')

    // Build vpn-connect hostname
    const hostname = `vpn-connect.${subdomain}.${domain}`

    const result: TunnelEndpointInfo = {
      tunnel_endpoint: `${hostname}:443`,
      hostname,
      ip: publicIP,
    }

    return {
      success: true,
      data: result,
    }
  } catch (err) {
    console.error('Get tunnel endpoint error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Check if tunnel-server pod is ready in the namespace
 */
async function checkTunnelServerReady(namespace: string): Promise<boolean> {
  try {
    const client = getK8sClient()

    // Get tunnel-server pod (StatefulSet creates pod with name: tunnel-server-0)
    const pod = await client.core.readNamespacedPod({
      name: 'tunnel-server-0',
      namespace,
    }) as V1Pod

    // Check if pod is ready
    if (pod.status?.conditions) {
      for (const condition of pod.status.conditions) {
        if (condition.type === 'Ready' && condition.status === 'True') {
          return true
        }
      }
    }

    return false
  } catch (err) {
    // Pod doesn't exist yet
    return false
  }
}
