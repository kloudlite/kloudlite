'use server'

import { revalidatePath } from 'next/cache'
import { environmentRepository } from '@kloudlite/lib/k8s'
import type { Environment, WorkMachine } from '@kloudlite/lib/k8s'
import { compositionService } from '@/lib/services/composition.service'
import { environmentService } from '@/lib/services/environment.service'
import { getSession } from '@/lib/get-session'
import { resourceStore } from '@/lib/resource-store'
import { watchNamespace, watchResourceInNamespace } from '@/lib/k8s-watcher'
import type { K8sService } from '@kloudlite/types'
import {
  environmentCreateSchema,
  environmentUpdateSchema,
  environmentNameSchema,
  forkEnvironmentSchema,
  importEnvironmentConfigSchema,
} from '@/lib/validations'

/**
 * Map a raw K8s Service object to the flattened K8sService type
 */
function mapRawService(svc: any): K8sService {
  return {
    name: svc.metadata?.name || '',
    namespace: svc.metadata?.namespace || '',
    type: svc.spec?.type || 'ClusterIP',
    clusterIP: svc.spec?.clusterIP || '',
    ports: (svc.spec?.ports || []).map((p: any) => ({
      name: p.name || '',
      protocol: p.protocol || 'TCP',
      port: p.port || 0,
      targetPort: String(p.targetPort || ''),
    })),
    selector: svc.spec?.selector || {},
    replicas: 0,
    image: undefined,
  }
}

/**
 * Get work machine for a user from the in-memory store
 */
function getWorkMachineForUser(username: string): WorkMachine | null {
  const machines = resourceStore.listClusterByLabel<WorkMachine>('workmachines', 'kloudlite.io/owned-by', username)
  return machines[0] || null
}

/**
 * Get the WorkMachine's namespace where environments should be created
 * Environments must be created in the WorkMachine namespace (wm-*)
 */
async function getWorkMachineNamespace(): Promise<string> {
  const session = await getSession()
  if (!session?.user?.username) {
    throw new Error('Not authenticated')
  }

  await resourceStore.waitForReady('workmachines')
  const workMachine = getWorkMachineForUser(session.user.username)
  if (!workMachine) {
    throw new Error(`No WorkMachine found for user ${session.user.username}`)
  }

  // Environments are created in the WorkMachine's targetNamespace
  return workMachine.spec.targetNamespace
}

/**
 * Server action to get full environments list with work machine and preferences
 */
export async function getEnvironmentsListFull() {
  try {
    console.log('[STORE] getEnvironmentsListFull: environments, workmachines, userpreferences')
    const session = await getSession()
    const username = session?.user?.username || session?.user?.email || ''

    // Ensure cluster-scoped stores are ready
    await resourceStore.waitForReady('workmachines')
    await resourceStore.waitForReady('userpreferences')

    const workMachineResult = getWorkMachineForUser(username)
    const preferencesResult = resourceStore.getCluster('userpreferences', username)

    // Get namespace from work machine's targetNamespace
    const namespace = workMachineResult?.spec?.targetNamespace
    if (!namespace) {
      return {
        success: false,
        error: 'No WorkMachine found',
        data: {
          environments: [],
          workMachine: null,
          preferences: null,
          pinnedEnvironmentIds: [],
          workMachineRunning: false,
        },
      }
    }

    // Ensure namespace watches are running and wait only for environments
    watchNamespace(namespace)
    await resourceStore.waitForReady('environments', namespace)

    const environments = resourceStore.list<Environment>('environments', namespace)

    // Get pinned environment names from preferences
    const pinnedEnvironmentIds = preferencesResult?.spec?.pinnedEnvironments || []

    // Check if work machine is running
    const workMachineRunning = workMachineResult?.status?.state === 'running' &&
      workMachineResult?.status?.isReady === true

    return {
      success: true,
      data: {
        environments,
        workMachine: workMachineResult,
        preferences: preferencesResult,
        pinnedEnvironmentIds,
        workMachineRunning,
      },
    }
  } catch (err) {
    console.error('Get environments list full error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
      data: {
        environments: [],
        workMachine: null,
        preferences: null,
        pinnedEnvironmentIds: [],
        workMachineRunning: false,
      },
    }
  }
}

/**
 * Server action to create an environment
 */
export async function createEnvironment(data: unknown) {
  // Validate input
  const validated = environmentCreateSchema.safeParse(data)
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    const namespace = await getWorkMachineNamespace()
    const createData = validated.data as import('@kloudlite/types').EnvironmentCreateRequest

    // Build Environment CRD object
    const environment: Environment = {
      apiVersion: 'environments.kloudlite.io/v1',
      kind: 'Environment',
      metadata: {
        name: createData.name,
        namespace,
      },
      spec: {
        ...createData.spec,
      },
    }

    console.log('[K8S-API] createEnvironment:', environment.metadata?.name)
    const result = await environmentRepository.create(namespace, environment)
    revalidatePath('/environments')
    return { success: true, data: result }
  } catch (err) {
    console.error('Create environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to update an environment
 */
export async function updateEnvironment(name: string, data: unknown) {
  // Validate environment name
  const nameValidation = environmentNameSchema.safeParse(name)
  if (!nameValidation.success) {
    return {
      success: false,
      error: 'Invalid environment name',
    }
  }

  // Validate update data
  const validated = environmentUpdateSchema.safeParse(data)
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    const namespace = await getWorkMachineNamespace()
    const updateData = validated.data as import('@kloudlite/types').EnvironmentUpdateRequest

    // Use patch for partial updates
    console.log('[K8S-API] updateEnvironment:', name)
    const result = await environmentRepository.patch(namespace, name, {
      spec: updateData.spec,
    })
    revalidatePath('/environments')
    revalidatePath(`/environments/${name}`)
    return { success: true, data: result }
  } catch (err) {
    console.error('Update environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to delete an environment
 */
export async function deleteEnvironment(name: string) {
  try {
    const namespace = await getWorkMachineNamespace()
    console.log('[K8S-API] deleteEnvironment:', name)
    await environmentRepository.delete(namespace, name)
    revalidatePath('/environments')
    return { success: true }
  } catch (err) {
    console.error('Delete environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to activate an environment
 */
export async function activateEnvironment(name: string) {
  try {
    const namespace = await getWorkMachineNamespace()
    console.log('[K8S-API] activateEnvironment:', name)
    const result = await environmentRepository.activate(namespace, name)
    revalidatePath('/environments')
    revalidatePath(`/environments/${name}`)
    return { success: true, data: result }
  } catch (err) {
    console.error('Activate environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to deactivate an environment
 */
export async function deactivateEnvironment(name: string) {
  try {
    const namespace = await getWorkMachineNamespace()
    console.log('[K8S-API] deactivateEnvironment:', name)
    const result = await environmentRepository.deactivate(namespace, name)
    revalidatePath('/environments')
    revalidatePath(`/environments/${name}`)
    return { success: true, data: result }
  } catch (err) {
    console.error('Deactivate environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get environment status
 */
export async function getEnvironmentStatus(name: string) {
  try {
    console.log('[STORE] getEnvironmentStatus:', name)
    const namespace = await getWorkMachineNamespace()
    watchNamespace(namespace)
    await resourceStore.waitForReady('environments', namespace)
    const environment = resourceStore.get<Environment>('environments', namespace, name)
    if (!environment) {
      return { success: false, error: 'Environment not found' }
    }
    return { success: true, data: environment.status }
  } catch (err) {
    console.error('Get environment status error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get an environment by name
 */
export async function getEnvironment(name: string) {
  try {
    console.log('[STORE] getEnvironment:', name)
    const namespace = await getWorkMachineNamespace()
    watchNamespace(namespace)
    await resourceStore.waitForReady('environments', namespace)
    const result = resourceStore.get<Environment>('environments', namespace, name)
    if (!result) {
      return { success: false, error: 'Environment not found' }
    }
    return { success: true, data: result }
  } catch (err) {
    console.error('Get environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to list all environments in the user's namespace
 */
export async function listEnvironments() {
  try {
    console.log('[STORE] listEnvironments')
    const namespace = await getWorkMachineNamespace()
    watchNamespace(namespace)
    await resourceStore.waitForReady('environments', namespace)
    const items = resourceStore.list<Environment>('environments', namespace)
    return { success: true, data: items }
  } catch (err) {
    console.error('List environments error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
      data: [],
    }
  }
}

/**
 * Server action to fork an environment
 */
export async function forkEnvironment(
  sourceName: string,
  targetName: string,
  targetNamespace: string,
  forkEnvVars: boolean,
  forkFiles: boolean,
  currentUser: string,
) {
  // Validate all parameters
  const validated = forkEnvironmentSchema.safeParse({
    sourceName,
    targetName,
    targetNamespace,
    forkEnvVars,
    forkFiles,
    currentUser,
  })
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    const result = await environmentService.forkEnvironment(
      validated.data.sourceName,
      validated.data.targetName,
      validated.data.targetNamespace,
      validated.data.forkEnvVars,
      validated.data.forkFiles,
      validated.data.currentUser,
    )
    revalidatePath('/environments')
    return { success: true, data: result }
  } catch (err) {
    console.error('Fork environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to export environment config
 * Returns all compositions, configs, secrets, and files as a YAML-compatible object
 */
export async function exportEnvironmentConfig(envName: string, targetNamespace: string) {
  try {
    // Fetch all data in parallel
    const [envVarsResult, filesResult, compositionsResult] = await Promise.all([
      environmentService.getEnvVars(envName).catch(() => ({ envVars: [], count: 0 })),
      environmentService.listFiles(envName).catch(() => ({ files: [], count: 0 })),
      compositionService.listCompositions(targetNamespace).catch(() => ({ compositions: [], count: 0 })),
    ])

    // Separate configs and secrets from envVars
    const configs: Record<string, string> = {}
    const secrets: Record<string, string> = {}
    for (const envVar of envVarsResult.envVars || []) {
      if (envVar.type === 'config') {
        configs[envVar.key] = envVar.value
      } else if (envVar.type === 'secret') {
        secrets[envVar.key] = envVar.value
      }
    }

    // Build export object
    const exportData = {
      apiVersion: 'kloudlite.io/v1',
      kind: 'EnvironmentExport',
      metadata: {
        name: envName,
        exportedAt: new Date().toISOString(),
      },
      configs,
      secrets,
      files: filesResult.files || [],
      compositions: (compositionsResult.compositions || []).map((comp) => ({
        name: comp.metadata?.name,
        spec: comp.spec,
      })),
    }

    return { success: true, data: exportData }
  } catch (err) {
    console.error('Export environment config error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to import environment config
 * Creates a new environment and imports all configs, secrets, files, and compositions
 */
export async function importEnvironmentConfig(
  newEnvName: string,
  targetNamespace: string,
  currentUser: string,
  exportData: unknown,
) {
  // Validate all parameters
  const validated = importEnvironmentConfigSchema.safeParse({
    newEnvName,
    targetNamespace,
    currentUser,
    exportData,
  })
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  const errors: string[] = []

  try {
    // Step 1: Create the environment
    // Don't pass targetNamespace - webhook will auto-generate it as env-{owner}--{name}
    const createResult = await environmentService.createEnvironment({
      name: validated.data.newEnvName,
      spec: {
        ownedBy: validated.data.currentUser,
        activated: false,
      },
    })

    if (!createResult) {
      return { success: false, error: 'Failed to create environment' }
    }

    // Get the targetNamespace from the created environment
    const targetNamespace = createResult.environment?.spec?.targetNamespace || ''

    const { exportData: validatedExportData } = validated.data

    // Step 2: Import configs
    if (validatedExportData.configs) {
      for (const [key, value] of Object.entries(validatedExportData.configs)) {
        try {
          await environmentService.createEnvVar(validated.data.newEnvName, key, value, 'config')
        } catch (err) {
          errors.push(`Failed to import config "${key}": ${err instanceof Error ? err.message : 'Unknown error'}`)
        }
      }
    }

    // Step 3: Import secrets
    if (validatedExportData.secrets) {
      for (const [key, value] of Object.entries(validatedExportData.secrets)) {
        try {
          await environmentService.createEnvVar(validated.data.newEnvName, key, value, 'secret')
        } catch (err) {
          errors.push(`Failed to import secret "${key}": ${err instanceof Error ? err.message : 'Unknown error'}`)
        }
      }
    }

    // Step 4: Import files
    if (validatedExportData.files) {
      for (const file of validatedExportData.files) {
        try {
          await environmentService.setFile(validated.data.newEnvName, file.name, file.content)
        } catch (err) {
          errors.push(`Failed to import file "${file.name}": ${err instanceof Error ? err.message : 'Unknown error'}`)
        }
      }
    }

    // Step 5: Import compositions
    if (validatedExportData.compositions && targetNamespace) {
      for (const comp of validatedExportData.compositions) {
        try {
          await compositionService.createComposition(targetNamespace, {
            name: comp.name,
            spec: comp.spec as Parameters<typeof compositionService.createComposition>[1]['spec'],
          })
        } catch (err) {
          errors.push(`Failed to import composition "${comp.name}": ${err instanceof Error ? err.message : 'Unknown error'}`)
        }
      }
    }

    revalidatePath('/environments')

    if (errors.length > 0) {
      return {
        success: true,
        data: { name: validated.data.newEnvName },
        warnings: errors,
      }
    }

    return { success: true, data: { name: validated.data.newEnvName } }
  } catch (err) {
    console.error('Import environment config error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to update environment access settings (visibility and sharedWith)
 */
export async function updateEnvironmentAccess(
  name: string,
  data: { visibility: 'private' | 'shared' | 'open'; sharedWith?: string[] },
) {
  try {
    // Use the existing updateEnvironment action with partial spec
    const result = await updateEnvironment(name, {
      spec: {
        visibility: data.visibility,
        sharedWith: data.sharedWith,
      },
    })
    return result
  } catch (err) {
    console.error('Update environment access error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get an environment by its hash or name
 * Uses the in-memory store with fallback strategies
 */
export async function getEnvironmentByHash(hashOrName: string) {
  try {
    console.log('[STORE] getEnvironmentByHash:', hashOrName)
    const namespace = await getWorkMachineNamespace()

    // Ensure namespace watches are running and wait only for environments
    watchNamespace(namespace)
    await resourceStore.waitForReady('environments', namespace)

    // Try to find by hash using label index (efficient)
    let environment = resourceStore.getByHash<Environment>('environments', namespace, hashOrName)

    // Fallback 1: search by status.hash
    if (!environment) {
      environment = resourceStore.findByStatusField<Environment>('environments', namespace, 'status.hash', hashOrName)
    }

    // Fallback 2: try direct name lookup
    if (!environment) {
      environment = resourceStore.get<Environment>('environments', namespace, hashOrName)
    }

    if (!environment) {
      return {
        success: false,
        error: 'Environment not found',
      }
    }

    // Get services from the target namespace
    // Only watch 'services' in targetNamespace to avoid connection accumulation (Issue 3)
    let services: K8sService[] = []
    const targetNs = environment.spec?.targetNamespace
    if (targetNs) {
      try {
        await watchResourceInNamespace('services', targetNs)
        services = resourceStore.list('services', targetNs).map(mapRawService)
      } catch (err) {
        console.error('Failed to fetch services:', err)
      }
    }

    return {
      success: true,
      data: {
        environment,
        services,
        compose: environment.spec?.compose || null,
        composeStatus: environment.status?.composeStatus || null,
        namespace: targetNs || '',
        isActive: environment.status?.state === 'active',
        envHash: environment.status?.hash || '',
        subdomain: environment.status?.subdomain || '',
      },
    }
  } catch (err) {
    console.error('Get environment by hash error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get environment details (environment + services + composition)
 */
export async function getEnvironmentDetails(name: string) {
  try {
    console.log('[STORE] getEnvironmentDetails:', name)
    const namespace = await getWorkMachineNamespace()
    watchNamespace(namespace)
    await resourceStore.waitForReady('environments', namespace)

    const environment = resourceStore.get<Environment>('environments', namespace, name)
    if (!environment) {
      return { success: false, error: 'Environment not found' }
    }

    // Get services from the target namespace
    let services: K8sService[] = []
    const targetNs = environment.spec?.targetNamespace
    if (targetNs) {
      try {
        await watchResourceInNamespace('services', targetNs)
        services = resourceStore.list('services', targetNs).map(mapRawService)
      } catch (err) {
        console.error('Failed to fetch services:', err)
      }
    }

    return {
      success: true,
      data: {
        environment,
        services,
        compose: environment.spec?.compose || null,
        composeStatus: environment.status?.composeStatus || null,
        namespace: targetNs || '',
        isActive: environment.status?.state === 'active',
        envHash: environment.status?.hash || '',
        subdomain: environment.status?.subdomain || '',
      },
    }
  } catch (err) {
    console.error('Get environment details error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get environment compose
 */
export async function getEnvironmentCompose(name: string) {
  try {
    console.log('[STORE] getEnvironmentCompose:', name)
    const namespace = await getWorkMachineNamespace()
    watchNamespace(namespace)
    await resourceStore.waitForReady('environments', namespace)
    const environment = resourceStore.get<Environment>('environments', namespace, name)
    if (!environment) {
      return { success: false, error: 'Environment not found' }
    }

    return {
      success: true,
      data: {
        name,
        compose: environment.spec?.compose || null,
        composeStatus: environment.status?.composeStatus || null,
      },
    }
  } catch (err) {
    console.error('Get environment compose error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to update environment compose
 */
export async function updateEnvironmentCompose(
  name: string,
  compose: import('@kloudlite/types').CompositionSpec | null,
) {
  try {
    const namespace = await getWorkMachineNamespace()

    console.log('[K8S-API] updateEnvironmentCompose:', name)
    const result = await environmentRepository.patch(namespace, name, {
      spec: {
        compose: compose ?? undefined,
      },
    })

    revalidatePath('/environments')
    revalidatePath(`/environments/${name}`)

    return {
      success: true,
      data: {
        message: 'Composition updated successfully',
        compose: result.spec?.compose || null,
        composeStatus: result.status?.composeStatus || null,
      },
    }
  } catch (err) {
    console.error('Update environment compose error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}
