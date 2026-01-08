'use server'

import { revalidatePath } from 'next/cache'
import { snapshotService } from '@/lib/services/snapshot.service'
import type { CreateSnapshotRequest, CreateWorkspaceFromSnapshotRequest, CreateEnvironmentFromSnapshotRequest } from '@/lib/services/snapshot.service'

/**
 * Server action to list snapshots for a workspace
 */
export async function listSnapshots(workspaceName: string, namespace: string) {
  try {
    const result = await snapshotService.listWorkspaceSnapshots(workspaceName, namespace)
    return { success: true, data: result }
  } catch (err) {
    console.error('List snapshots error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to create a snapshot for a workspace
 */
export async function createSnapshot(
  workspaceName: string,
  namespace: string,
  data?: CreateSnapshotRequest,
) {
  try {
    const result = await snapshotService.createWorkspaceSnapshot(workspaceName, namespace, data)
    revalidatePath(`/workspaces/${namespace}/${workspaceName}`)
    return { success: true, data: result }
  } catch (err) {
    console.error('Create snapshot error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to list snapshots for an environment
 */
export async function listEnvironmentSnapshots(environmentName: string) {
  try {
    const result = await snapshotService.listEnvironmentSnapshots(environmentName)
    return { success: true, data: result }
  } catch (err) {
    console.error('List environment snapshots error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to create a snapshot for an environment
 */
export async function createEnvironmentSnapshot(
  environmentName: string,
  data?: CreateSnapshotRequest,
) {
  try {
    const result = await snapshotService.createEnvironmentSnapshot(environmentName, data)
    revalidatePath(`/environments/${environmentName}`)
    return { success: true, data: result }
  } catch (err) {
    console.error('Create environment snapshot error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get a snapshot by name
 */
export async function getSnapshot(snapshotName: string) {
  try {
    const result = await snapshotService.get(snapshotName)
    return { success: true, data: result }
  } catch (err) {
    console.error('Get snapshot error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to restore a snapshot
 */
export async function restoreSnapshot(snapshotName: string) {
  try {
    const result = await snapshotService.restore(snapshotName)
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Restore snapshot error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to delete a snapshot
 */
export async function deleteSnapshot(snapshotName: string) {
  try {
    await snapshotService.delete(snapshotName)
    revalidatePath('/workspaces')
    return { success: true }
  } catch (err) {
    console.error('Delete snapshot error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to push a snapshot to the registry
 */
export async function pushSnapshot(snapshotName: string, tag: string) {
  try {
    const result = await snapshotService.push(snapshotName, tag)
    revalidatePath('/workspaces')
    revalidatePath('/environments')
    return { success: true, data: result }
  } catch (err) {
    console.error('Push snapshot error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to pull a snapshot from the registry
 */
export async function pullSnapshot(
  repository: string,
  tag: string,
  name?: string,
) {
  try {
    const result = await snapshotService.pull(repository, tag, name)
    revalidatePath('/workspaces')
    revalidatePath('/environments')
    return { success: true, data: result }
  } catch (err) {
    console.error('Pull snapshot error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to list pushed snapshots available for forking
 */
export async function listPushedSnapshots(type?: 'workspace' | 'environment') {
  try {
    const result = await snapshotService.listPushed(type)
    return { success: true, data: result }
  } catch (err) {
    console.error('List pushed snapshots error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to create a workspace from a pushed snapshot
 */
export async function createWorkspaceFromSnapshot(data: CreateWorkspaceFromSnapshotRequest) {
  try {
    const result = await snapshotService.createWorkspaceFromSnapshot(data)
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Create workspace from snapshot error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to create an environment from a pushed snapshot
 */
export async function createEnvironmentFromSnapshot(data: CreateEnvironmentFromSnapshotRequest) {
  try {
    const result = await snapshotService.createEnvironmentFromSnapshot(data)
    revalidatePath('/environments')
    return { success: true, data: result }
  } catch (err) {
    console.error('Create environment from snapshot error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}
