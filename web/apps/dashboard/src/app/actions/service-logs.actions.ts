'use server'

import { getK8sClient } from '@kloudlite/lib/k8s'

/**
 * Get service logs
 * Returns recent log lines from a service's pod
 */
export async function getServiceLogs(
  namespace: string,
  serviceName: string,
  options?: {
    tailLines?: number
    sinceSeconds?: number
    previous?: boolean
  }
) {
  try {
    const client = getK8sClient()

    // Get pods for the service
    const { body: podList } = await client.core.listNamespacedPod({
      namespace,
      labelSelector: `app=${serviceName}`,
    })

    if (!podList.items || podList.items.length === 0) {
      return {
        success: false,
        error: 'No pods found for service',
      }
    }

    // Get logs from the first pod
    const pod = podList.items[0]
    const podName = pod.metadata!.name!

    const { body: logs } = await client.core.readNamespacedPodLog({
      name: podName,
      namespace,
      tailLines: options?.tailLines || 100,
      sinceSeconds: options?.sinceSeconds,
      previous: options?.previous,
    })

    return {
      success: true,
      data: {
        podName,
        logs: logs as string,
        timestamp: new Date().toISOString(),
      },
    }
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}
