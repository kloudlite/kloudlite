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
        name: workspace.metadata.name,
        namespace: workspace.metadata.namespace,
        phase: workspace.status?.phase || 'Unknown',
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

