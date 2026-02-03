'use server'

import { revalidatePath } from 'next/cache'
import { environmentRepository, workMachineRepository, userPreferencesRepository, serviceRepository } from '@kloudlite/lib/k8s'
import type { Environment } from '@kloudlite/lib/k8s'
import { compositionService } from '@/lib/services/composition.service'
import { environmentService } from '@/lib/services/environment.service'
import { getSession } from '@/lib/get-session'
import {
  environmentCreateSchema,
  environmentUpdateSchema,
  environmentNameSchema,
  forkEnvironmentSchema,
  importEnvironmentConfigSchema,
} from '@/lib/validations'

/**
 * Get the WorkMachine's namespace where environments should be created
 * Environments must be created in the WorkMachine namespace (wm-*)
 */
async function getWorkMachineNamespace(): Promise<string> {
  const session = await getSession()
  if (!session?.user?.username) {
    throw new Error('Not authenticated')
  }

  const workMachine = await workMachineRepository.getByOwner(session.user.username)
  if (!workMachine) {
    throw new Error(`No WorkMachine found for user ${session.user.username}`)
  }

  // Return the WorkMachine's own namespace (wm-*), not the targetNamespace
  return workMachine.metadata.namespace || `wm-${session.user.username}`
}

/**
 * Server action to get full environments list with work machine and preferences
 */
export async function getEnvironmentsListFull() {
  try {
    const session = await getSession()
    const username = session?.user?.username || session?.user?.email || ''

    // Fetch work machine and preferences in parallel
    const [workMachineResult, preferencesResult] = await Promise.all([
      workMachineRepository.getByOwner(username).catch(() => null),
      userPreferencesRepository.getByUser(username).catch(() => null),
    ])

    // Get namespace from work machine
    const namespace = workMachineResult?.metadata?.namespace || `wm-${username}`

    // Fetch environments
    const environmentsResult = await environmentRepository.list(namespace).catch(() => ({ items: [] }))

    // Get pinned environment names from preferences
    const pinnedEnvironmentIds = preferencesResult?.spec?.pinnedEnvironments || []

    // Check if work machine is running
    const workMachineRunning = workMachineResult?.status?.state === 'running' &&
      workMachineResult?.status?.isReady === true

    return {
      success: true,
      data: {
        environments: environmentsResult.items || [],
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
    const namespace = await getWorkMachineNamespace()
    const environment = await environmentRepository.get(namespace, name)
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
    const namespace = await getWorkMachineNamespace()
    const result = await environmentRepository.get(namespace, name)
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
    const namespace = await getWorkMachineNamespace()
    const result = await environmentRepository.list(namespace)
    // Kubernetes list returns { items: [...] }
    return { success: true, data: result.items || [] }
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
 * Server action to get an environment by its hash
 * The hash is a unique 8-char hex identifier generated from envName-owner
 */
export async function getEnvironmentByHash(hash: string) {
  try {
    const namespace = await getWorkMachineNamespace()
    const environmentsList = await environmentRepository.list(namespace)
    const environment = environmentsList.items?.find(
      (env) => env.status?.hash === hash
    )

    if (!environment) {
      return {
        success: false,
        error: 'Environment not found',
      }
    }

    // Get services from the target namespace if environment is active
    let services: import('@kloudlite/types').K8sService[] = []
    const targetNamespace = environment.spec?.targetNamespace
    if (targetNamespace && environment.status?.state === 'active') {
      try {
        services = await serviceRepository.list(targetNamespace)
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
        namespace: targetNamespace || '',
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
    const namespace = await getWorkMachineNamespace()

    // Fetch environment
    const environment = await environmentRepository.get(namespace, name)

    // Get services from the target namespace if environment is active
    let services: import('@kloudlite/types').K8sService[] = []
    const targetNamespace = environment.spec?.targetNamespace
    if (targetNamespace && environment.status?.state === 'active') {
      try {
        services = await serviceRepository.list(targetNamespace)
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
        namespace: targetNamespace || '',
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
    const namespace = await getWorkMachineNamespace()
    const environment = await environmentRepository.get(namespace, name)

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

    // Read-modify-write pattern: get environment, update compose, then save
    const environment = await environmentRepository.get(namespace, name)

    // Update the compose field (convert null to undefined for type compatibility)
    const updatedEnvironment = {
      ...environment,
      spec: {
        ...environment.spec,
        compose: compose ?? undefined,
      },
    }

    // Save back using update (full replace)
    const result = await environmentRepository.update(namespace, name, updatedEnvironment)

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
