import { cache } from 'react'

import { credentials, Metadata } from '@grpc/grpc-js'
import { getServerSession } from 'next-auth'

import { AuthClient } from '@/grpc/auth.external'
import { getAuthOptions } from '@/lib/auth/get-auth-options'

// Use 127.0.0.1 instead of localhost to avoid IPv6 issues
// Default to port 3001 if not specified
const BACKEND_URL = process.env.BACKEND_URL || '127.0.0.1:3001'

// Singleton client instance for connection reuse
let authClient: AuthClient | null = null

export function getAuthClient() {
  if (!authClient) {
    authClient = new AuthClient(
      BACKEND_URL,
      credentials.createInsecure()
    )
  }
  return authClient
}

// Cache session per request using React cache
const getCachedSession = cache(async () => {
  const authOpts = await getAuthOptions()
  return getServerSession(authOpts)
})

// Helper function to get metadata with user context and JWT token
export async function getAuthMetadata() {
  const session = await getCachedSession()
  if (!session?.user) {
    throw new Error('User not authenticated')
  }
  
  const metadata = new Metadata()
  // Add JWT token for authentication
  if ((session as any).accessToken) {
    metadata.add('authorization', `Bearer ${(session as any).accessToken}`)
  }
  
  return metadata
}