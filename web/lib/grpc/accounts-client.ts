import { cache } from 'react'

import { credentials, Metadata } from '@grpc/grpc-js'
import { getServerSession } from 'next-auth'

import { AccountsClient } from '@/grpc/accounts.external'
import { getAuthOptions } from '@/lib/auth/get-auth-options'

// Use 127.0.0.1 instead of localhost to avoid IPv6 issues
// Default to port 3001 if not specified
const BACKEND_URL = process.env.BACKEND_URL || '127.0.0.1:3001'

// Singleton client instance for connection reuse
let accountsClient: AccountsClient | null = null

export function getAccountsClient() {
  if (!accountsClient) {
    accountsClient = new AccountsClient(
      BACKEND_URL,
      credentials.createInsecure()
    )
  }
  return accountsClient
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
  // These might be redundant if the server validates JWT, but keeping for backward compatibility
  metadata.add('userId', session.user.id)
  metadata.add('userEmail', session.user.email || '')
  metadata.add('userName', session.user.name || '')
  
  return metadata
}

// Helper for parallel gRPC calls
export async function parallelGrpcCalls<T extends Record<string, Promise<any>>>(
  calls: T
): Promise<{ [K in keyof T]: Awaited<T[K]> }> {
  const keys = Object.keys(calls) as (keyof T)[]
  const promises = keys.map(key => calls[key])
  const results = await Promise.all(promises)
  
  const output = {} as { [K in keyof T]: Awaited<T[K]> }
  keys.forEach((key, index) => {
    output[key] = results[index]
  })
  
  return output
}