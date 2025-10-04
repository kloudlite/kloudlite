'use server'

import { revalidatePath } from 'next/cache'
import { workspaceService } from '@/lib/services/workspace.service'
import type {
  WorkspaceCreateRequest,
  WorkspaceUpdateRequest,
} from '@/types/workspace'

/**
 * Server action to list workspaces
 */
export async function listWorkspaces(
  namespace: string = 'default',
  params?: {
    owner?: string
    workMachine?: string
    status?: string
    limit?: number
    continue?: string
  }
) {
  try {
    const result = await workspaceService.list(namespace, params)
    return { success: true, data: result }
  } catch (error: any) {
    console.error('List workspaces error:', error)
    return {
      success: false,
      error: error.message || 'Failed to list workspaces',
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
  } catch (error: any) {
    console.error('Get workspace error:', error)
    return {
      success: false,
      error: error.message || 'Failed to get workspace',
    }
  }
}

/**
 * Server action to create a workspace
 */
export async function createWorkspace(
  namespace: string,
  data: WorkspaceCreateRequest
) {
  try {
    const result = await workspaceService.create(data, namespace)
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Create workspace error:', error)
    return {
      success: false,
      error: error.message || 'Failed to create workspace',
    }
  }
}

/**
 * Server action to update a workspace
 */
export async function updateWorkspace(
  name: string,
  namespace: string,
  data: WorkspaceUpdateRequest
) {
  try {
    const result = await workspaceService.update(name, data, namespace)
    revalidatePath('/workspaces')
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Update workspace error:', error)
    return {
      success: false,
      error: error.message || 'Failed to update workspace',
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
  } catch (error: any) {
    console.error('Delete workspace error:', error)
    return {
      success: false,
      error: error.message || 'Failed to delete workspace',
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
  } catch (error: any) {
    console.error('Suspend workspace error:', error)
    return {
      success: false,
      error: error.message || 'Failed to suspend workspace',
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
  } catch (error: any) {
    console.error('Activate workspace error:', error)
    return {
      success: false,
      error: error.message || 'Failed to activate workspace',
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
  } catch (error: any) {
    console.error('Archive workspace error:', error)
    return {
      success: false,
      error: error.message || 'Failed to archive workspace',
    }
  }
}
