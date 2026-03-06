'use server'

import { revalidatePath } from 'next/cache'
import { workspaceRepository } from '@kloudlite/lib/k8s'
import type { Workspace } from '@kloudlite/lib/k8s'
import { workspaceService } from '@/lib/services/workspace.service'
import { getSession } from '@/lib/get-session'
import { resourceStore } from '@/lib/resource-store'
import {
  workspaceCreateSchema,
  workspaceUpdateSchema,
  workspaceNameSchema,
} from '@/lib/validations'
import { getWorkMachineForUser } from './workspace.actions.shared'

/**
 * Server action to create a workspace.
 * Security: ownedBy and workmachine are derived from session.
 */
export async function createWorkspace(data: unknown) {
  const session = await getSession()
  if (!session?.user) {
    return { success: false, error: 'Not authenticated' }
  }

  const username = session.user.username || session.user.email || ''
  if (!username) {
    return { success: false, error: 'Unable to determine username' }
  }

  const validated = workspaceCreateSchema.safeParse(data)
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
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

    const workspace: Workspace = {
      apiVersion: 'workspaces.kloudlite.io/v1',
      kind: 'Workspace',
      metadata: {
        name: createData.name,
        namespace,
      },
      spec: {
        ...createData.spec,
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
    return { success: false, error: error.message }
  }
}

/**
 * Server action to update a workspace.
 */
export async function updateWorkspace(name: string, namespace: string, data: unknown) {
  const nameValidation = workspaceNameSchema.safeParse(name)
  if (!nameValidation.success) {
    return { success: false, error: 'Invalid workspace name' }
  }

  const validated = workspaceUpdateSchema.safeParse(data)
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    const updateData = validated.data as import('@kloudlite/types').WorkspaceUpdateRequest
    console.log('[K8S-API] updateWorkspace:', name)
    const result = await workspaceRepository.patch(namespace, name, {
      spec: updateData.spec,
    })
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Update workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return { success: false, error: error.message }
  }
}

/**
 * Server action to delete a workspace.
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
    return { success: false, error: error.message }
  }
}

/**
 * Server action to suspend a workspace.
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
    return { success: false, error: error.message }
  }
}

/**
 * Server action to activate a workspace.
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
    return { success: false, error: error.message }
  }
}

/**
 * Server action to archive a workspace.
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
    return { success: false, error: error.message }
  }
}

/**
 * Server action to fork a workspace (deprecated - use snapshot-based forking).
 */
export async function forkWorkspace(
  sourceWorkspaceName: string,
  data: unknown,
  namespace: string = 'default',
) {
  const sourceNameValidation = workspaceNameSchema.safeParse(sourceWorkspaceName)
  if (!sourceNameValidation.success) {
    return { success: false, error: 'Invalid source workspace name' }
  }

  const validated = workspaceCreateSchema.safeParse(data)
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    const result = await workspaceService.fork(
      sourceWorkspaceName,
      validated.data as import('@kloudlite/types').WorkspaceCreateRequest,
      namespace,
    )
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (err) {
    console.error('Fork workspace error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return { success: false, error: error.message }
  }
}
