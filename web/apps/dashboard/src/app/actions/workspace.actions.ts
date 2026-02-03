'use server'

import { revalidatePath } from 'next/cache'
import { workspaceRepository, packageRequestRepository, workMachineRepository, userPreferencesRepository } from '@kloudlite/lib/k8s'
import type { Workspace } from '@kloudlite/lib/k8s'
import { workspaceService } from '@/lib/services/workspace.service'
import { getSession } from '@/lib/get-session'
import {
  workspaceCreateSchema,
  workspaceUpdateSchema,
  workspaceNameSchema,
  packageUpdateSchema,
} from '@/lib/validations'

/**
 * Server action to get full workspaces list with work machine and preferences
 */
export async function getWorkspacesListFull() {
  try {
    const session = await getSession()
    const username = session?.user?.username || session?.user?.email || ''

    // Fetch work machine and preferences in parallel
    const [workMachineResult, preferencesResult] = await Promise.all([
      workMachineRepository.getByOwner(username).catch(() => null),
      userPreferencesRepository.getByUser(username).catch(() => null),
    ])

    // Get namespace from work machine
    const namespace = workMachineResult?.spec?.targetNamespace || 'default'

    // Fetch workspaces
    const workspacesResult = await workspaceRepository.list(namespace).catch(() => ({ items: [] }))

    // Get pinned workspace IDs from preferences
    const pinnedWorkspaceIds = preferencesResult?.spec?.pinnedWorkspaces?.map(
      (ws) => `${ws.namespace}/${ws.name}`
    ) || []

    // Check if work machine is running
    const workMachineRunning = workMachineResult?.status?.state === 'running' &&
      workMachineResult?.status?.isReady === true

    return {
      success: true,
      data: {
        workspaces: workspacesResult.items || [],
        workMachine: workMachineResult,
        preferences: preferencesResult,
        pinnedWorkspaceIds,
        workMachineRunning,
      },
    }
  } catch (err) {
    console.error('Get workspaces list full error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
      data: {
        workspaces: [],
        workMachine: null,
        preferences: null,
        pinnedWorkspaceIds: [],
        workMachineRunning: false,
      },
    }
  }
}

/**
 * Server action to get a workspace by its hash or name
 * Uses label selector for efficient lookup by hash (kloudlite.io/hash label)
 * Falls back to direct name lookup if hash not found
 */
export async function getWorkspaceByHash(hashOrName: string) {
  try {
    const session = await getSession()
    const username = session?.user?.username || session?.user?.email || ''

    // Get work machine to find the namespace
    const workMachine = await workMachineRepository.getByOwner(username)
    if (!workMachine) {
      return {
        success: false,
        error: 'No work machine found',
      }
    }

    const namespace = workMachine.spec?.targetNamespace || 'default'

    // Try to find by hash using label selector (efficient)
    let workspace = await workspaceRepository.getByHash(namespace, hashOrName)

    // Fallback: try direct name lookup if hash not found
    if (!workspace) {
      try {
        workspace = await workspaceRepository.get(namespace, hashOrName)
      } catch {
        // Not found by name either
      }
    }

    if (!workspace) {
      return {
        success: false,
        error: 'Workspace not found',
      }
    }

    // Get package request for this workspace
    let packageRequest = null
    try {
      packageRequest = await packageRequestRepository.getByWorkspace(namespace, workspace.metadata!.name!)
    } catch {
      // PackageRequest may not exist
    }

    // Check if work machine is running
    const workMachineRunning = workMachine.status?.state === 'running' &&
      workMachine.status?.isReady === true

    return {
      success: true,
      data: {
        workspace,
        packageRequest,
        workMachineRunning,
      },
    }
  } catch (err) {
    console.error('Get workspace by hash error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to list workspaces
 */
export async function listWorkspaces(namespace: string = 'default') {
  try {
    const result = await workspaceRepository.list(namespace)
    return { success: true, data: { items: result.items, metadata: result.metadata } }
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
    const result = await workspaceRepository.get(namespace, name)
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
 * Security: ownedBy and workmachine are derived from the authenticated session,
 * not from frontend input, to prevent privilege escalation
 */
export async function createWorkspace(data: unknown) {
  // Get authenticated session
  const session = await getSession()
  if (!session?.user) {
    return {
      success: false,
      error: 'Not authenticated',
    }
  }

  const username = session.user.username || session.user.email || ''
  if (!username) {
    return {
      success: false,
      error: 'Unable to determine username',
    }
  }

  // Validate input
  const validated = workspaceCreateSchema.safeParse(data)
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    // Get the user's work machine to determine namespace
    const workMachine = await workMachineRepository.getByOwner(username)
    if (!workMachine) {
      return {
        success: false,
        error: 'No work machine found. Please set up your work machine first.',
      }
    }

    const namespace = workMachine.spec?.targetNamespace || 'default'
    const workmachineName = workMachine.metadata?.name || `wm-${username}`

    const createData = validated.data as import('@kloudlite/types').WorkspaceCreateRequest

    // Build Workspace CRD object with secure values from session
    const workspace: Workspace = {
      apiVersion: 'workspaces.kloudlite.io/v1',
      kind: 'Workspace',
      metadata: {
        name: createData.name,
        namespace,
      },
      spec: {
        ...createData.spec,
        // Override with secure values from session - don't trust frontend
        ownedBy: username,
        workmachine: workmachineName,
      },
    }

    const result = await workspaceRepository.create(namespace, workspace)
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
    const updateData = validated.data as import('@kloudlite/types').WorkspaceUpdateRequest

    // Use patch for partial updates
    const result = await workspaceRepository.patch(namespace, name, {
      spec: updateData.spec,
    })
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
    await workspaceRepository.delete(namespace, name)
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
    const result = await workspaceRepository.suspend(namespace, name)
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
    const result = await workspaceRepository.activate(namespace, name)
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
    const result = await workspaceRepository.archive(namespace, name)
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
    // Try to get existing PackageRequest by workspace label
    const existingPkgReq = await packageRequestRepository.getByWorkspace(namespace, workspaceName)

    if (existingPkgReq) {
      // Update existing PackageRequest
      const result = await packageRequestRepository.updatePackages(
        namespace,
        existingPkgReq.metadata!.name!,
        validated.data.packages as import('@kloudlite/lib/k8s').PackageSpec[]
      )
      revalidatePath('/workspaces')
      return { success: true, data: result }
    } else {
      // Create new PackageRequest
      const packageRequest: import('@kloudlite/lib/k8s').PackageRequest = {
        apiVersion: 'packages.kloudlite.io/v1',
        kind: 'PackageRequest',
        metadata: {
          name: `${workspaceName}-packages`,
          namespace,
          labels: {
            'kloudlite.io/workspace': workspaceName,
          },
        },
        spec: {
          workspaceRef: workspaceName,
          profileName: `${workspaceName}-packages`,
          packages: validated.data.packages as import('@kloudlite/lib/k8s').PackageSpec[],
        },
      }

      const result = await packageRequestRepository.create(namespace, packageRequest)
      revalidatePath('/workspaces')
      return { success: true, data: result }
    }
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
    const packageRequest = await packageRequestRepository.getByWorkspace(namespace, workspaceName)

    if (!packageRequest) {
      // PackageRequest doesn't exist yet (workspace has no packages configured)
      return { success: true, data: null }
    }

    return { success: true, data: packageRequest }
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
