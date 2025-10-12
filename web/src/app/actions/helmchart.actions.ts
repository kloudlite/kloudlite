'use server'

import { auth } from '@/lib/auth'
import { helmChartService } from '@/lib/services/helmchart.service'
import type { HelmChartCreateRequest, HelmChartUpdateRequest } from '@/types/helmchart'

export async function createHelmChart(
  namespace: string,
  data: HelmChartCreateRequest,
  user: string
) {
  try {
    const session = await auth()
    if (!session) {
      return { success: false, error: 'Unauthorized' }
    }

    const result = await helmChartService.createHelmChart(namespace, data)

    return { success: true, data: result }
  } catch (error) {
    console.error('Failed to create helm chart:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to create helm chart',
    }
  }
}

export async function updateHelmChart(
  namespace: string,
  name: string,
  data: HelmChartUpdateRequest,
  user: string
) {
  try {
    const session = await auth()
    if (!session) {
      return { success: false, error: 'Unauthorized' }
    }

    const result = await helmChartService.updateHelmChart(namespace, name, data)

    return { success: true, data: result }
  } catch (error) {
    console.error('Failed to update helm chart:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to update helm chart',
    }
  }
}

export async function deleteHelmChart(namespace: string, name: string, user: string) {
  try {
    const session = await auth()
    if (!session) {
      return { success: false, error: 'Unauthorized' }
    }

    await helmChartService.deleteHelmChart(namespace, name)

    return { success: true }
  } catch (error) {
    console.error('Failed to delete helm chart:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to delete helm chart',
    }
  }
}
