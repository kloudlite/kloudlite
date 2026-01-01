'use server'

import { revalidatePath } from 'next/cache'
import { environmentService } from '@/lib/services/environment.service'
import { compositionService } from '@/lib/services/composition.service'
import {
  environmentCreateSchema,
  environmentUpdateSchema,
  environmentNameSchema,
  cloneEnvironmentSchema,
  importEnvironmentConfigSchema,
} from '@/lib/validations'

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
    const result = await environmentService.createEnvironment(validated.data)
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
    const result = await environmentService.updateEnvironment(name, validated.data)
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
    const result = await environmentService.deleteEnvironment(name)
    revalidatePath('/environments')
    return { success: true, data: result }
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
    const result = await environmentService.activateEnvironment(name)
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
    const result = await environmentService.deactivateEnvironment(name)
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
    const result = await environmentService.getEnvironmentStatus(name)
    return { success: true, data: result }
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
 * Server action to clone an environment
 */
export async function cloneEnvironment(
  sourceName: string,
  targetName: string,
  targetNamespace: string,
  cloneEnvVars: boolean,
  cloneFiles: boolean,
  currentUser: string,
) {
  // Validate all parameters
  const validated = cloneEnvironmentSchema.safeParse({
    sourceName,
    targetName,
    targetNamespace,
    cloneEnvVars,
    cloneFiles,
    currentUser,
  })
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    const result = await environmentService.cloneEnvironment(
      validated.data.sourceName,
      validated.data.targetName,
      validated.data.targetNamespace,
      validated.data.cloneEnvVars,
      validated.data.cloneFiles,
      validated.data.currentUser,
    )
    revalidatePath('/environments')
    return { success: true, data: result }
  } catch (err) {
    console.error('Clone environment error:', err)
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
    const createResult = await environmentService.createEnvironment({
      name: validated.data.newEnvName,
      spec: {
        targetNamespace: validated.data.targetNamespace,
        ownedBy: validated.data.currentUser,
        activated: false,
      },
    })

    if (!createResult) {
      return { success: false, error: 'Failed to create environment' }
    }

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
    if (validatedExportData.compositions) {
      for (const comp of validatedExportData.compositions) {
        try {
          await compositionService.createComposition(validated.data.targetNamespace, {
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
  data: { visibility: 'private' | 'shared' | 'public'; sharedWith?: string[] },
) {
  try {
    const result = await environmentService.updateEnvironment(name, {
      spec: {
        visibility: data.visibility,
        sharedWith: data.sharedWith,
      },
    })
    revalidatePath('/environments')
    revalidatePath(`/environments/${name}`)
    return { success: true, data: result }
  } catch (err) {
    console.error('Update environment access error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}
