'use server'

import { environmentRepository } from '@kloudlite/lib/k8s'

/**
 * Get environment status
 * Used for polling environment state
 */
export async function getEnvironmentStatus(name: string) {
  try {
    const environment = await environmentRepository.get(name)

    return {
      success: true,
      data: {
        name: environment.metadata.name,
        state: environment.status?.state || 'Unknown',
        activated: environment.spec.activated,
        message: environment.status?.message,
        conditions: environment.status?.conditions || [],
        resourceCount: {
          deployments: environment.status?.resourceCount?.deployments || 0,
          services: environment.status?.resourceCount?.services || 0,
          configmaps: environment.status?.resourceCount?.configmaps || 0,
          secrets: environment.status?.resourceCount?.secrets || 0,
        },
        lastUpdated: new Date().toISOString(),
      },
    }
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}
