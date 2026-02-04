'use server'

import { cache } from 'react'
import { revalidatePath } from 'next/cache'
import { workspaceRepository, packageRequestRepository, workMachineRepository, userPreferencesRepository } from '@kloudlite/lib/k8s'
import type { Workspace } from '@kloudlite/lib/k8s'
import { workspaceService } from '@/lib/services/workspace.service'
import { getSession } from '@/lib/get-session'
import { cachedFetch, CacheTTL, invalidateCache } from '@/lib/cache'
import {
  workspaceCreateSchema,
  workspaceUpdateSchema,
  workspaceNameSchema,
  packageUpdateSchema,
} from '@/lib/validations'

/**
 * Cached work machine fetch - deduplicates within the same request (React cache)
 * Also uses LRU cache for cross-request caching
 * Uses SHORT TTL (30s) because status (running/stopped) can change
 */
const getCachedWorkMachine = cache(async (username: string) => {
  return cachedFetch(
    `workMachine:${username}`,
    () => workMachineRepository.getByOwner(username).catch(() => null),
    CacheTTL.SHORT // 30 seconds - status changes (running/stopped)
  )
})

/**
 * Server action to get full workspaces list with work machine and preferences
 */
export async function getWorkspacesListFull() {
  try {
    const session = await getSession()
    const username = session?.user?.username || session?.user?.email || ''
    const cachedNamespace = session?.user?.namespace

    // If we have cached namespace, we can parallelize ALL fetches
    if (cachedNamespace) {
      const [workMachineResult, preferencesResult, workspacesResult] = await Promise.all([
        getCachedWorkMachine(username),
        cachedFetch(
          `preferences:${username}`,
          () => userPreferencesRepository.getByUser(username).catch(() => null),
          CacheTTL.NORMAL // 1 minute
        ),
        cachedFetch(
          `workspaces:${cachedNamespace}`,
          () => workspaceRepository.list(cachedNamespace).catch(() => ({ items: [] })),
          CacheTTL.SHORT // 30 seconds - workspaces change more often
        ),
      ])

      const pinnedWorkspaceIds = preferencesResult?.spec?.pinnedWorkspaces?.map(
        (ws) => `${ws.namespace}/${ws.name}`
      ) || []

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
    }

    // Fallback: No cached namespace - need to fetch work machine first
    const [workMachineResult, preferencesResult] = await Promise.all([
      getCachedWorkMachine(username),
      cachedFetch(
        `preferences:${username}`,
        () => userPreferencesRepository.getByUser(username).catch(() => null),
        CacheTTL.NORMAL
      ),
    ])

    const namespace = workMachineResult?.spec?.targetNamespace || 'default'

    const workspacesResult = await cachedFetch(
      `workspaces:${namespace}`,
      () => workspaceRepository.list(namespace).catch(() => ({ items: [] })),
      CacheTTL.SHORT
    )

    const pinnedWorkspaceIds = preferencesResult?.spec?.pinnedWorkspaces?.map(
      (ws) => `${ws.namespace}/${ws.name}`
    ) || []

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
    const cachedNamespace = session?.user?.namespace

    // If no cached namespace, fetch work machine first (required)
    if (!cachedNamespace) {
      const workMachine = await getCachedWorkMachine(username)

      if (!workMachine) {
        return { success: false, error: 'No work machine found' }
      }

      const namespace = workMachine.spec?.targetNamespace || 'default'

      // Now fetch workspace and package request in sequence (no parallelization possible without namespace)
      let workspace = await workspaceRepository.getByHash(namespace, hashOrName)
      if (!workspace) {
        const workspacesList = await workspaceRepository.list(namespace)
        workspace = workspacesList.items?.find((ws) => ws.status?.hash === hashOrName) || null
      }
      if (!workspace) {
        try { workspace = await workspaceRepository.get(namespace, hashOrName) } catch { /* not found */ }
      }
      if (!workspace) {
        return { success: false, error: 'Workspace not found' }
      }

      const packageRequest = await packageRequestRepository.getByWorkspace(namespace, workspace.metadata!.name!).catch(() => null)
      const workMachineRunning = workMachine.status?.state === 'running' && workMachine.status?.isReady === true

      return { success: true, data: { workspace, packageRequest, workMachineRunning } }
    }

    // OPTIMIZED PATH: namespace is cached in session
    const namespace = cachedNamespace

    // Parallel fetch: workspace + work machine status (both only need session data)
    const [workspaceResult, workMachine] = await Promise.all([
      workspaceRepository.getByHash(namespace, hashOrName),
      getCachedWorkMachine(username),
    ])

    let workspace = workspaceResult

    // Fallback 1: search by status.hash
    if (!workspace) {
      const workspacesList = await workspaceRepository.list(namespace)
      workspace = workspacesList.items?.find((ws) => ws.status?.hash === hashOrName) || null
    }

    // Fallback 2: direct name lookup
    if (!workspace) {
      try {
        workspace = await workspaceRepository.get(namespace, hashOrName)
      } catch { /* not found */ }
    }

    if (!workspace) {
      return { success: false, error: 'Workspace not found' }
    }

    // Fetch package request (needs workspace name)
    const packageRequest = await packageRequestRepository.getByWorkspace(namespace, workspace.metadata!.name!).catch(() => null)

    const workMachineRunning = workMachine?.status?.state === 'running' && workMachine?.status?.isReady === true

    return {
      success: true,
      data: { workspace, packageRequest, workMachineRunning },
    }
  } catch (err) {
    console.error('Get workspace by hash error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return { success: false, error: error.message }
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
    invalidateCache(`workspaces:${namespace}*`)
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
    invalidateCache(`workspaces:${namespace}*`)
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
    invalidateCache(`workspaces:${namespace}*`)
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
    invalidateCache(`workspaces:${namespace}*`)
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
    invalidateCache(`workspaces:${namespace}*`)
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
    invalidateCache(`workspaces:${namespace}*`)
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
 * Server action to get workspace metrics from Kubernetes metrics-server
 */
export async function getWorkspaceMetrics(name: string, namespace: string = 'default') {
  try {
    const { metricsRepository } = await import('@kloudlite/lib/k8s')

    // Workspace pod name follows the pattern: ws-{workspaceName}
    const podName = `ws-${name}`

    const podMetrics = await metricsRepository.getPodMetrics(namespace, podName)

    // Parse CPU (format: "123456789n" for nanocores or "123m" for millicores)
    let cpuUsage = 0
    if (podMetrics.containers && podMetrics.containers.length > 0) {
      for (const container of podMetrics.containers) {
        const cpuStr = container.usage.cpu
        if (cpuStr.endsWith('n')) {
          // Convert nanocores to millicores
          cpuUsage += parseInt(cpuStr.slice(0, -1), 10) / 1_000_000
        } else if (cpuStr.endsWith('m')) {
          cpuUsage += parseInt(cpuStr.slice(0, -1), 10)
        } else {
          // Assume cores, convert to millicores
          cpuUsage += parseFloat(cpuStr) * 1000
        }
      }
    }

    // Parse memory (format: "123456Ki" for kibibytes)
    let memoryUsage = 0
    if (podMetrics.containers && podMetrics.containers.length > 0) {
      for (const container of podMetrics.containers) {
        const memStr = container.usage.memory
        if (memStr.endsWith('Ki')) {
          memoryUsage += parseInt(memStr.slice(0, -2), 10) * 1024
        } else if (memStr.endsWith('Mi')) {
          memoryUsage += parseInt(memStr.slice(0, -2), 10) * 1024 * 1024
        } else if (memStr.endsWith('Gi')) {
          memoryUsage += parseInt(memStr.slice(0, -2), 10) * 1024 * 1024 * 1024
        } else {
          memoryUsage += parseInt(memStr, 10)
        }
      }
    }

    return {
      success: true,
      data: {
        cpu: {
          usage: Math.round(cpuUsage), // millicores
        },
        memory: {
          usage: memoryUsage, // bytes
          usagePercent: 0, // Would need limits to calculate
        },
        timestamp: podMetrics.timestamp,
      },
    }
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
