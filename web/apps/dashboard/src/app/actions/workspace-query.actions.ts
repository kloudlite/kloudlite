'use server'

import type { PackageRequest, Workspace, UserPreferences } from '@kloudlite/lib/k8s'
import { watchNamespace } from '@/lib/k8s-watcher'
import { resourceStore } from '@/lib/resource-store'
import { getSession } from '@/lib/get-session'
import { getWorkMachineForUser } from './workspace.actions.shared'

/**
 * Server action to get full workspaces list with work machine and preferences.
 */
export async function getWorkspacesListFull() {
  try {
    console.log('[STORE] getWorkspacesListFull: workspaces, workmachines, userpreferences')
    const session = await getSession()
    const username = session?.user?.username || session?.user?.email || ''
    const cachedNamespace = session?.user?.namespace

    await resourceStore.waitForReady('workmachines')
    await resourceStore.waitForReady('userpreferences')

    const workMachineResult = getWorkMachineForUser(username)
    const preferencesResult = resourceStore.getCluster<UserPreferences>('userpreferences', username)
    const namespace = cachedNamespace || workMachineResult?.spec?.targetNamespace || 'default'

    watchNamespace(namespace)
    await resourceStore.waitForReady('workspaces', namespace)
    const workspaces = resourceStore.list<Workspace>('workspaces', namespace)

    const pinnedWorkspaceIds =
      preferencesResult?.spec?.pinnedWorkspaces?.map((ws) => `${ws.namespace}/${ws.name}`) || []

    const workMachineRunning =
      workMachineResult?.status?.state === 'running' && workMachineResult?.status?.isReady === true

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
 * Server action to get a workspace by its hash or name.
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

    watchNamespace(namespace)
    await resourceStore.waitForReady('workspaces', namespace)

    let workspace = resourceStore.getByHash<Workspace>('workspaces', namespace, hashOrName)
    if (!workspace) {
      workspace = resourceStore.findByStatusField<Workspace>('workspaces', namespace, 'status.hash', hashOrName)
    }
    if (!workspace) {
      workspace = resourceStore.get<Workspace>('workspaces', namespace, hashOrName)
    }

    if (!workspace) {
      return { success: false, error: 'Workspace not found' }
    }

    const packageRequest =
      resourceStore.listByLabel<PackageRequest>(
        'packagerequests',
        namespace,
        'kloudlite.io/workspace',
        workspace.metadata!.name!,
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
 * Server action to list workspaces.
 */
export async function listWorkspaces(namespace: string = 'default') {
  try {
    console.log('[STORE] listWorkspaces:', namespace)
    watchNamespace(namespace)
    await resourceStore.waitForReady('workspaces', namespace)
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
 * Server action to get a workspace.
 */
export async function getWorkspace(name: string, namespace: string = 'default') {
  try {
    console.log('[STORE] getWorkspace:', name)
    watchNamespace(namespace)
    await resourceStore.waitForReady('workspaces', namespace)
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
 * Server action to get workspace metrics from Kubernetes metrics-server.
 */
export async function getWorkspaceMetrics(name: string, namespace: string = 'default') {
  try {
    const { metricsRepository } = await import('@kloudlite/lib/k8s')
    const podName = `ws-${name}`

    console.log('[K8S-API] getWorkspaceMetrics:', podName)
    const podMetrics = await metricsRepository.getPodMetrics(namespace, podName)

    let cpuUsage = 0
    if (podMetrics.containers && podMetrics.containers.length > 0) {
      for (const container of podMetrics.containers) {
        const cpuStr = container.usage.cpu
        if (cpuStr.endsWith('n')) cpuUsage += parseInt(cpuStr.slice(0, -1), 10) / 1_000_000
        else if (cpuStr.endsWith('m')) cpuUsage += parseInt(cpuStr.slice(0, -1), 10)
        else cpuUsage += parseFloat(cpuStr) * 1000
      }
    }

    let memoryUsage = 0
    if (podMetrics.containers && podMetrics.containers.length > 0) {
      for (const container of podMetrics.containers) {
        const memStr = container.usage.memory
        if (memStr.endsWith('Ki')) memoryUsage += parseInt(memStr.slice(0, -2), 10) * 1024
        else if (memStr.endsWith('Mi')) memoryUsage += parseInt(memStr.slice(0, -2), 10) * 1024 * 1024
        else if (memStr.endsWith('Gi'))
          memoryUsage += parseInt(memStr.slice(0, -2), 10) * 1024 * 1024 * 1024
        else memoryUsage += parseInt(memStr, 10)
      }
    }

    return {
      success: true,
      data: {
        cpu: { usage: Math.round(cpuUsage) },
        memory: { usage: memoryUsage, usagePercent: 0 },
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
