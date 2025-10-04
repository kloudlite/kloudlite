'use server'

import { auth } from '@/lib/auth'
import { apiClient } from '@/lib/api-client'

interface Provider {
  type: string
  enabled: boolean
  clientId: string
  clientSecret?: string
}

export async function updateProvider(type: string, provider: Provider) {
  // Check authentication using NextAuth
  const session = await auth()

  if (!session || !session.user?.roles?.includes('super-admin')) {
    return { success: false, error: 'Not authenticated or insufficient permissions' }
  }

  try {
    // Since we're using NextAuth, we need to make the request directly
    // The backend should validate based on the user's session/JWT
    await apiClient.put(`/api/v1/providers/${type}`, provider)

    return { success: true }
  } catch (error: any) {
    console.error('Error updating provider:', error)
    return { success: false, error: error.message || 'Failed to update provider' }
  }
}