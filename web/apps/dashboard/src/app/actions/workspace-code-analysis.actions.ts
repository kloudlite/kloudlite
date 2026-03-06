'use server'

import type { CodeAnalysisResponse } from './workspace-code-analysis.types'

/**
 * Server action to get code analysis reports for a workspace.
 */
export async function getCodeAnalysis(
  workspaceName: string,
  namespace: string = 'default',
): Promise<{ success: boolean; data?: CodeAnalysisResponse; error?: string }> {
  try {
    const { env } = await import('@/lib/env')
    const { getAuthToken } = await import('@/lib/get-session')

    const token = await getAuthToken()
    if (!token) {
      return {
        success: false,
        error: 'Not authenticated',
      }
    }

    const url = `${env.apiUrl}/api/v1/namespaces/${namespace}/workspaces/${workspaceName}/code-analysis`
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    })

    if (!response.ok) {
      const errorText = await response.text()
      return {
        success: false,
        error: errorText || 'Failed to get code analysis',
      }
    }

    const data = await response.json()
    return { success: true, data }
  } catch (err) {
    console.error('Get code analysis error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to trigger a manual code analysis for a workspace.
 */
export async function triggerCodeAnalysis(workspaceName: string, namespace: string = 'default') {
  try {
    const { env } = await import('@/lib/env')
    const { getAuthToken } = await import('@/lib/get-session')

    const token = await getAuthToken()
    if (!token) {
      return {
        success: false,
        error: 'Not authenticated',
      }
    }

    const url = `${env.apiUrl}/api/v1/namespaces/${namespace}/workspaces/${workspaceName}/code-analysis`
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      const errorText = await response.text()
      return {
        success: false,
        error: errorText || 'Failed to trigger code analysis',
      }
    }

    const data = await response.json()
    return { success: true, data }
  } catch (err) {
    console.error('Trigger code analysis error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}
