'use server'

import { revalidatePath } from 'next/cache'
import { workspaceService } from '@/lib/services/workspace.service'
import {
  workspaceCreateSchema,
  workspaceUpdateSchema,
  workspaceNameSchema,
  packageUpdateSchema,
} from '@/lib/validations'
import type { WorkspaceListParams } from '@kloudlite/types'

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
export async function createWorkspace(namespace: string, data: unknown) {
  // Validate input
  const validated = workspaceCreateSchema.safeParse(data)
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    // Cast to WorkspaceCreateRequest - workmachine is auto-populated by webhook from namespace
    const result = await workspaceService.create(
      validated.data as import('@kloudlite/types').WorkspaceCreateRequest,
      namespace
    )
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
export async function updateWorkspace(name: string, namespace: string, data: unknown) {
  // Validate workspace name
  const nameValidation = workspaceNameSchema.safeParse(name)
  if (!nameValidation.success) {
    return {
      success: false,
      error: 'Invalid workspace name',
    }
  }

  // Validate update data
  const validated = workspaceUpdateSchema.safeParse(data)
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    // Cast to WorkspaceUpdateRequest - backend handles partial updates
    const result = await workspaceService.update(
      name,
      validated.data as import('@kloudlite/types').WorkspaceUpdateRequest,
      namespace
    )
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
 * Server action to fork a workspace (deprecated - use snapshot-based forking)
 */
export async function forkWorkspace(
  sourceWorkspaceName: string,
  data: unknown,
  namespace: string = 'default',
) {
  // Validate source workspace name
  const sourceNameValidation = workspaceNameSchema.safeParse(sourceWorkspaceName)
  if (!sourceNameValidation.success) {
    return {
      success: false,
      error: 'Invalid source workspace name',
    }
  }

  // Validate fork data
  const validated = workspaceCreateSchema.safeParse(data)
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    // Cast to WorkspaceCreateRequest - workmachine is auto-populated by webhook from namespace
    const result = await workspaceService.fork(
      sourceWorkspaceName,
      validated.data as import('@kloudlite/types').WorkspaceCreateRequest,
      namespace
    )
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Fork workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to update packages in a workspace's PackageRequest
 * Creates the PackageRequest if it doesn't exist
 */
export async function updatePackageRequest(
  workspaceName: string,
  packages: unknown,
  namespace: string = 'default',
) {
  // Validate workspace name
  const nameValidation = workspaceNameSchema.safeParse(workspaceName)
  if (!nameValidation.success) {
    return {
      success: false,
      error: 'Invalid workspace name',
    }
  }

  // Validate packages
  const validated = packageUpdateSchema.safeParse({ packages })
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    const { env } = await import('@/lib/env')
    const { getAuthToken } = await import('@/lib/get-session')

    const token = await getAuthToken()
    if (!token) {
      return {
        success: false,
        error: 'Not authenticated',
      }
    }

    const url = `${env.apiUrl}/api/v1/namespaces/${namespace}/workspaces/${workspaceName}/packages`
    const response = await fetch(url, {
      method: 'PUT',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ packages: validated.data.packages }),
    })

    if (!response.ok) {
      const errorText = await response.text()
      return {
        success: false,
        error: errorText || 'Failed to update packages',
      }
    }

    const data = await response.json()
    revalidatePath('/workspaces')
    return { success: true, data }
  } catch (err) {
    console.error('Update package request error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get package request status for a workspace
 * Uses the workspace packages endpoint which returns the PackageRequest (source of truth)
 */
export async function getPackageRequest(workspaceName: string, namespace: string = 'default') {
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

    // Use the workspace packages endpoint
    const url = `${env.apiUrl}/api/v1/namespaces/${namespace}/workspaces/${workspaceName}/packages`
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

/**
 * Code analysis report types
 */
export interface CodeAnalysisFinding {
  severity: string
  category: string
  file: string
  line: number
  title: string
  description: string
  recommendation: string
}

export interface CodeAnalysisReport {
  version: string
  type: string
  workspace: string
  analyzedAt: string
  summary: {
    score: number
    criticalCount: number
    highCount: number
    mediumCount: number
    lowCount: number
  }
  findings: CodeAnalysisFinding[]
}

export interface CodeAnalysisResponse {
  security: CodeAnalysisReport | null
  quality: CodeAnalysisReport | null
  status: {
    watching: boolean
    inProgress: boolean
    pendingAnalysis: boolean
    lastAnalysis?: string
  }
}

/**
 * Server action to get code analysis reports for a workspace
 */
export async function getCodeAnalysis(
  workspaceName: string,
  namespace: string = 'default',
): Promise<{ success: boolean; data?: CodeAnalysisResponse; error?: string }> {
  try {
    const { env } = await import('@/lib/env')
    const { getAuthToken } = await import('@/lib/get-session')

    const token = await getAuthToken()
    if (!token) {
      return {
        success: false,
        error: 'Not authenticated',
      }
    }

    const url = `${env.apiUrl}/api/v1/namespaces/${namespace}/workspaces/${workspaceName}/code-analysis`
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
        error: errorText || 'Failed to get code analysis',
      }
    }

    const data = await response.json()
    return { success: true, data }
  } catch (err) {
    console.error('Get code analysis error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to trigger a manual code analysis for a workspace
 */
export async function triggerCodeAnalysis(workspaceName: string, namespace: string = 'default') {
  try {
    const { env } = await import('@/lib/env')
    const { getAuthToken } = await import('@/lib/get-session')

    const token = await getAuthToken()
    if (!token) {
      return {
        success: false,
        error: 'Not authenticated',
      }
    }

    const url = `${env.apiUrl}/api/v1/namespaces/${namespace}/workspaces/${workspaceName}/code-analysis`
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      const errorText = await response.text()
      return {
        success: false,
        error: errorText || 'Failed to trigger code analysis',
      }
    }

    const data = await response.json()
    return { success: true, data }
  } catch (err) {
    console.error('Trigger code analysis error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}
