'use server'

import type { MachineTypeCreateRequest, MachineTypeUpdateRequest } from '@kloudlite/types'
import type { MachineType as DashboardMachineType } from '@/types/machine'
import { getPlatformApiClient, type KubernetesResource } from '@/lib/platform-api'

type MachineType = DashboardMachineType & KubernetesResource<DashboardMachineType['spec']>

const readOnlyMachineTypeError = 'Machine types are managed by platform-api defaults and are read-only.'

/**
 * Server action to list all machine types
 * Reads from in-memory store (populated by K8s watcher)
 */
export async function listMachineTypes() {
  try {
    console.log('[STORE] listMachineTypes')
    const client = await getPlatformApiClient()
    const items = await client.listClusterResources<MachineType>('machinetypes')
    return { success: true, data: items }
  } catch (err) {
    console.error('List machine types error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get a specific machine type
 */
export async function getMachineType(name: string) {
  try {
    console.log('[STORE] getMachineType:', name)
    const client = await getPlatformApiClient()
    const result = await client.getClusterResource<MachineType>('machinetypes', name)
    return { success: true, data: result }
  } catch (err) {
    console.error('Get machine type error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to create a machine type
 */
export async function createMachineType(data: MachineTypeCreateRequest) {
  void data
  return { success: false, error: readOnlyMachineTypeError }
}

/**
 * Server action to update a machine type
 */
export async function updateMachineType(name: string, data: MachineTypeUpdateRequest) {
  void name
  void data
  return { success: false, error: readOnlyMachineTypeError }
}

/**
 * Server action to delete a machine type
 */
export async function deleteMachineType(name: string) {
  void name
  return { success: false, error: readOnlyMachineTypeError }
}

/**
 * Server action to activate a machine type
 */
export async function activateMachineType(name: string) {
  void name
  return { success: false, error: readOnlyMachineTypeError }
}

/**
 * Server action to deactivate a machine type
 */
export async function deactivateMachineType(name: string) {
  void name
  return { success: false, error: readOnlyMachineTypeError }
}

/**
 * Server action to set a machine type as default
 */
export async function setMachineTypeAsDefault(name: string) {
  void name
  return { success: false, error: readOnlyMachineTypeError }
}
