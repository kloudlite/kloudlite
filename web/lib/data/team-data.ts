import { cache } from 'react'

import { getAccountsClient, getAuthMetadata } from '@/lib/grpc/accounts-client'

// Cache user teams for the duration of the request
export const getUserTeams = cache(async () => {
  const client = getAccountsClient()
  const metadata = await getAuthMetadata()
  
  const session = await (await import('next-auth')).getServerSession(await (await import('@/lib/auth/get-auth-options')).getAuthOptions())
  if (!session?.user) {
    return { teams: [] }
  }
  
  return new Promise<any>((resolve, reject) => {
    client.listTeams({ userId: session.user.id }, metadata, (error, response) => {
      if (error) {
        console.error('Failed to fetch teams:', error)
        resolve({ teams: [] })
      } else {
        resolve(response)
      }
    })
  })
})

// Transform team data to match UI expectations
export async function listUserTeams() {
  const response = await getUserTeams()
  return response.teams || []
}