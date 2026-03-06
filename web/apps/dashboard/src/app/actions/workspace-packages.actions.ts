'use server'

import { revalidatePath } from 'next/cache'
import { packageRequestRepository } from '@kloudlite/lib/k8s'
import type { PackageRequest } from '@kloudlite/lib/k8s'
import { watchNamespace } from '@/lib/k8s-watcher'
import { resourceStore } from '@/lib/resource-store'
import { packageUpdateSchema, workspaceNameSchema } from '@/lib/validations'

/**
 * Server action to update packages in a workspace's PackageRequest.
 * Creates the PackageRequest if it doesn't exist.
 */
export async function updatePackageRequest(
  workspaceName: string,
  packages: unknown,
  namespace: string = 'default',
) {
  const nameValidation = workspaceNameSchema.safeParse(workspaceName)
  if (!nameValidation.success) {
    return { success: false, error: 'Invalid workspace name' }
  }

  const validated = packageUpdateSchema.safeParse({ packages })
  if (!validated.success) {
    return {
      success: false,
      error: validated.error.errors.map((e) => e.message).join(', '),
    }
  }

  try {
    console.log('[STORE] updatePackageRequest: checking existing for', workspaceName)
    watchNamespace(namespace)
    await resourceStore.waitForReady('packagerequests', namespace)
    const existingPkgReq =
      resourceStore.listByLabel<PackageRequest>(
        'packagerequests',
        namespace,
        'kloudlite.io/workspace',
        workspaceName,
      )[0] || null

    if (existingPkgReq) {
      console.log('[K8S-API] updatePackageRequest: updating', existingPkgReq.metadata!.name!)
      const result = await packageRequestRepository.updatePackages(
        namespace,
        existingPkgReq.metadata!.name!,
        validated.data.packages as import('@kloudlite/lib/k8s').PackageSpec[],
      )
      revalidatePath('/workspaces')
      return { success: true, data: result }
    }

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
  } catch (err) {
    console.error('Update package request error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return { success: false, error: error.message }
  }
}

/**
 * Server action to get package request status for a workspace.
 */
export async function getPackageRequest(workspaceName: string, namespace: string = 'default') {
  try {
    console.log('[STORE] getPackageRequest:', workspaceName)
    watchNamespace(namespace)
    await resourceStore.waitForReady('packagerequests', namespace)
    const packageRequest =
      resourceStore.listByLabel<PackageRequest>(
        'packagerequests',
        namespace,
        'kloudlite.io/workspace',
        workspaceName,
      )[0] || null

    if (!packageRequest) {
      return { success: true, data: null }
    }

    return { success: true, data: packageRequest }
  } catch (err) {
    console.error('Get package request error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return { success: false, error: error.message }
  }
}
