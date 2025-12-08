'use server'

import { revalidatePath } from 'next/cache'
import { workspaceService } from '@/lib/services/workspace.service'
import type {
  WorkspaceCreateRequest,
  WorkspaceUpdateRequest,
  WorkspaceListParams,
} from '@kloudlite/types'

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
    // Import required modules dynamically to ensure this only runs on server
    const { env } = await import('@/lib/env')
    const { getAuthToken } = await import('@/lib/get-session')

    const token = await getAuthToken()
    if (!token) {
      return {
        success: false,
        error: 'Not authenticated',
      }
    }

    const url = `${env.apiUrl}/api/v1/namespaces/${namespace}/workspaces/${name}/metrics`
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${token}`,
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

/**
 * Server action to get package request status for a workspace
 * PackageRequest is cluster-scoped with name format: {workspace-name}-packages
 */
export async function getPackageRequest(workspaceName: string) {
  try {
    // Import required modules dynamically to ensure this only runs on server
    const { env } = await import('@/lib/env')
    const { getAuthToken } = await import('@/lib/get-session')

    const token = await getAuthToken()
    if (!token) {
      return {
        success: false,
        error: 'Not authenticated',
      }
    }

    // PackageRequest is cluster-scoped with naming convention: {workspace-name}-packages
    const packageRequestName = `${workspaceName}-packages`
    const url = `${env.apiUrl}/apis/workspaces.kloudlite.io/v1/packagerequests/${packageRequestName}`
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    })

    if (!response.ok) {
      if (response.status === 404) {
        // PackageRequest doesn't exist yet (workspace has no packages configured)
        return { success: true, data: null }
      }
      const errorText = await response.text()
      return {
        success: false,
        error: errorText || 'Failed to get package request',
      }
    }

    const data = await response.json()
    return { success: true, data }
  } catch (err) {
    console.error('Get package request error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}
