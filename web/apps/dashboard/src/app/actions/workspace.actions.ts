'use server'

import { revalidatePath } from 'next/cache'
import { workspaceRepository, packageRequestRepository } from '@kloudlite/lib/k8s'
import type { Workspace, WorkMachine, PackageRequest } from '@kloudlite/lib/k8s'
import { workspaceService } from '@/lib/services/workspace.service'
import { getSession } from '@/lib/get-session'
import { resourceStore } from '@/lib/resource-store'
import { watchNamespace } from '@/lib/k8s-watcher'
import {
  workspaceCreateSchema,
  workspaceUpdateSchema,
  workspaceNameSchema,
  packageUpdateSchema,
} from '@/lib/validations'

/**
 * Get work machine for a user from the in-memory store
 */
function getWorkMachineForUser(username: string): WorkMachine | null {
  const machines = resourceStore.listClusterByLabel<WorkMachine>('workmachines', 'kloudlite.io/owned-by', username)
  return machines[0] || null
}

/**
 * Server action to get full workspaces list with work machine and preferences
 */
export async function getWorkspacesListFull() {
  try {
    console.log('[STORE] getWorkspacesListFull: workspaces, workmachines, userpreferences')
    const session = await getSession()
    const username = session?.user?.username || session?.user?.email || ''
    const cachedNamespace = session?.user?.namespace

    // Ensure cluster-scoped stores are ready
    await resourceStore.waitForReady('workmachines')
    await resourceStore.waitForReady('userpreferences')

    const workMachineResult = getWorkMachineForUser(username)
    const preferencesResult = resourceStore.getCluster('userpreferences', username)

    // Determine namespace from session cache or work machine
    const namespace = cachedNamespace || workMachineResult?.spec?.targetNamespace || 'default'

    // Ensure namespace watches are running and ready
    await watchNamespace(namespace)

    const workspaces = resourceStore.list<Workspace>('workspaces', namespace)

    const pinnedWorkspaceIds = preferencesResult?.spec?.pinnedWorkspaces?.map(
      (ws: any) => `${ws.namespace}/${ws.name}`
    ) || []

    const workMachineRunning = workMachineResult?.status?.state === 'running' &&
      workMachineResult?.status?.isReady === true

    return {
      success: true,
      data: {
        workspaces,
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
 * Looks up from the in-memory store with multiple fallback strategies
 */
export async function getWorkspaceByHash(hashOrName: string) {
  try {
    console.log('[STORE] getWorkspaceByHash:', hashOrName)
    const session = await getSession()
    const username = session?.user?.username || session?.user?.email || ''
    const cachedNamespace = session?.user?.namespace

    await resourceStore.waitForReady('workmachines')

    const workMachine = getWorkMachineForUser(username)
    const namespace = cachedNamespace || workMachine?.spec?.targetNamespace || 'default'

    // Ensure namespace watches are running
    await watchNamespace(namespace)

    // Try to find by hash label (most common)
    let workspace = resourceStore.getByHash<Workspace>('workspaces', namespace, hashOrName)

    // Fallback 1: search by status.hash
    if (!workspace) {
      workspace = resourceStore.findByStatusField<Workspace>('workspaces', namespace, 'status.hash', hashOrName)
    }

    // Fallback 2: direct name lookup
    if (!workspace) {
      workspace = resourceStore.get<Workspace>('workspaces', namespace, hashOrName)
    }

    if (!workspace) {
      return { success: false, error: 'Workspace not found' }
    }

    // Get package request from store
    const packageRequest = resourceStore.listByLabel<PackageRequest>(
      'packagerequests', namespace, 'kloudlite.io/workspace', workspace.metadata!.name!
    )[0] || null

    const workMachineRunning = workMachine?.status?.state === 'running' && workMachine?.status?.isReady === true

    return { success: true, data: { workspace, packageRequest, workMachineRunning } }
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
    console.log('[STORE] listWorkspaces:', namespace)
    await watchNamespace(namespace)
    const items = resourceStore.list<Workspace>('workspaces', namespace)
    return { success: true, data: { items, metadata: {} } }
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
    console.log('[STORE] getWorkspace:', name)
    await watchNamespace(namespace)
    const result = resourceStore.get<Workspace>('workspaces', namespace, name)
    if (!result) {
      return { success: false, error: 'Workspace not found' }
    }
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
    await resourceStore.waitForReady('workmachines')
    const workMachine = getWorkMachineForUser(username)
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

    console.log('[K8S-API] createWorkspace:', workspace.metadata?.name)
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
    console.log('[K8S-API] updateWorkspace:', name)
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
    console.log('[K8S-API] deleteWorkspace:', name)
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
    console.log('[K8S-API] suspendWorkspace:', name)
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
    console.log('[K8S-API] activateWorkspace:', name)
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
    console.log('[K8S-API] archiveWorkspace:', name)
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
 * Server action to get workspace metrics from Kubernetes metrics-server
 */
export async function getWorkspaceMetrics(name: string, namespace: string = 'default') {
  try {
    const { metricsRepository } = await import('@kloudlite/lib/k8s')

    // Workspace pod name follows the pattern: ws-{workspaceName}
    const podName = `ws-${name}`

    console.log('[K8S-API] getWorkspaceMetrics:', podName)
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
    // Try to get existing PackageRequest by workspace label from store
    console.log('[STORE] updatePackageRequest: checking existing for', workspaceName)
    await watchNamespace(namespace)
    const existingPkgReq = resourceStore.listByLabel<PackageRequest>(
      'packagerequests', namespace, 'kloudlite.io/workspace', workspaceName
    )[0] || null

    if (existingPkgReq) {
      // Update existing PackageRequest
      console.log('[K8S-API] updatePackageRequest: updating', existingPkgReq.metadata!.name!)
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

      console.log('[K8S-API] updatePackageRequest: creating', packageRequest.metadata?.name)
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
    console.log('[STORE] getPackageRequest:', workspaceName)
    await watchNamespace(namespace)
    const packageRequest = resourceStore.listByLabel<PackageRequest>(
      'packagerequests', namespace, 'kloudlite.io/workspace', workspaceName
    )[0] || null

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
