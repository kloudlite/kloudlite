'use server'

import { workspaceRepository } from '@kloudlite/lib/k8s'

/**
 * Get workspace status
 * Used for polling workspace state
 */
export async function getWorkspaceStatus(namespace: string, name: string) {
  try {
    const workspace = await workspaceRepository.get(namespace, name)

    return {
      success: true,
      data: {
        name: workspace.metadata!.name,
        namespace: workspace.metadata!.namespace,
        phase: workspace.status?.phase || 'Unknown',
        state: workspace.status?.state || workspace.spec.state,
        isReady: workspace.status?.isReady ?? false,
        message: workspace.status?.message,
        podName: workspace.status?.podName,
        conditions: workspace.status?.conditions || [],
        lastUpdated: new Date().toISOString(),
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

/**
 * Get workspace metrics (CPU, memory, disk usage)
 */
export async function getWorkspaceMetrics(namespace: string, name: string) {
  try {
    const workspace = await workspaceRepository.get(namespace, name)

    // Return metrics from workspace status
    return {
      success: true,
      data: {
        cpu: workspace.status?.metrics?.cpu || 0,
        memory: workspace.status?.metrics?.memory || 0,
        disk: workspace.status?.metrics?.disk || 0,
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
