'use server'

import { revalidatePath } from 'next/cache'
import { getSession } from '@/lib/get-session'
import { workspaceService } from '@/lib/services/workspace.service'
import type {
  WorkspaceCreateRequest,
  WorkspaceUpdateRequest,
  WorkspaceListParams,
} from '@/types/workspace'

/**
 * Server action to list workspaces
 */
export async function listWorkspaces(namespace: string = 'default', params?: WorkspaceListParams) {
  try {
    const result = await workspaceService.list(namespace, params)
    return { success: true, data: result }
  } catch (err) {
    console.error('List workspaces error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get a workspace
 */
export async function getWorkspace(name: string, namespace: string = 'default') {
  try {
    const result = await workspaceService.get(name, namespace)
    return { success: true, data: result }
  } catch (err) {
    console.error('Get workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to create a workspace
 */
export async function createWorkspace(namespace: string, data: WorkspaceCreateRequest) {
  try {
    const result = await workspaceService.create(data, namespace)
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Create workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to update a workspace
 */
export async function updateWorkspace(
  name: string,
  namespace: string,
  data: WorkspaceUpdateRequest,
) {
  try {
    const result = await workspaceService.update(name, data, namespace)
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Update workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to delete a workspace
 */
export async function deleteWorkspace(name: string, namespace: string = 'default') {
  try {
    await workspaceService.delete(name, namespace)
    revalidatePath('/workspaces')
    return { success: true }
  } catch (err) {
    console.error('Delete workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to suspend a workspace
 */
export async function suspendWorkspace(name: string, namespace: string = 'default') {
  try {
    const result = await workspaceService.suspend(name, namespace)
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Suspend workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to activate a workspace
 */
export async function activateWorkspace(name: string, namespace: string = 'default') {
  try {
    const result = await workspaceService.activate(name, namespace)
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Activate workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to archive a workspace
 */
export async function archiveWorkspace(name: string, namespace: string = 'default') {
  try {
    const result = await workspaceService.archive(name, namespace)
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Archive workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get workspace metrics
 */
export async function getWorkspaceMetrics(name: string, namespace: string = 'default') {
  try {
    // Import auth and env dynamically to ensure this only runs on server
    // Removed unused auth import
    const { env } = await import('@/lib/env')

    const session = await getSession()
    if (!session?.user?.backendToken) {
      return {
        success: false,
        error: 'Not authenticated',
      }
    }

    const url = `${env.apiUrl}/api/v1/namespaces/${namespace}/workspaces/${name}/metrics`
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${session.user.backendToken}`,
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    })

    if (!response.ok) {
      const errorText = await response.text()
      return {
        success: false,
        error: errorText || 'Failed to get workspace metrics',
      }
    }

    const data = await response.json()
    return { success: true, data }
  } catch (err) {
    console.error('Get workspace metrics error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get node metrics for a work machine
 * @param workMachineName - The name of the work machine to get metrics for
 */
export async function getWorkMachineMetrics(workMachineName: string) {
  try {
    if (!workMachineName) {
      return {
        success: false,
        error: 'Work machine name is required',
      }
    }

    // Import auth and env dynamically to ensure this only runs on server
    const { env } = await import('@/lib/env')

    const session = await getSession()
    if (!session?.user?.backendToken) {
      return {
        success: false,
        error: 'Not authenticated',
      }
    }

    const url = `${env.apiUrl}/api/v1/work-machines/${workMachineName}/metrics`
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${session.user.backendToken}`,
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    })

    if (!response.ok) {
      const errorText = await response.text()
      return {
        success: false,
        error: errorText || 'Failed to get work machine metrics',
      }
    }

    const data = await response.json()

    // Check if we actually have metrics data (not just an empty response)
    if (!data.cpu || !data.cpu.capacity || data.cpu.capacity === 0) {
      return {
        success: false,
        error: 'Metrics not yet available',
      }
    }

    return { success: true, data }
  } catch (err) {
    console.error('Get work machine metrics error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

export async function getWorkMachineGPUMetrics(workMachineName: string) {
  try {
    if (!workMachineName) {
      return {
        success: false,
        error: 'Work machine name is required',
      }
    }

    // Import auth and env dynamically to ensure this only runs on server
    const { env } = await import('@/lib/env')

    const session = await getSession()
    if (!session?.user?.backendToken) {
      return {
        success: false,
        error: 'Not authenticated',
      }
    }

    const url = `${env.apiUrl}/api/v1/work-machines/${workMachineName}/gpu-metrics`
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${session.user.backendToken}`,
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    })

    if (!response.ok) {
      const errorText = await response.text()
      return {
        success: false,
        error: errorText || 'Failed to get GPU metrics',
      }
    }

    const data = await response.json()

    // If no GPU detected, return success with detected=false
    if (!data.detected) {
      return { success: true, data: { detected: false } }
    }

    return { success: true, data }
  } catch (err) {
    console.error('Get work machine GPU metrics error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to clone a workspace
 */
export async function cloneWorkspace(
  sourceWorkspaceName: string,
  data: WorkspaceCreateRequest,
  namespace: string = 'default',
) {
  try {
    const result = await workspaceService.clone(sourceWorkspaceName, data, namespace)
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Clone workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}
