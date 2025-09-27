import { cache } from 'react'

import { getServerSession } from 'next-auth'

import { getAuthOptions } from '@/lib/auth/get-auth-options'
import { getAccountsClient, getAuthMetadata as getAccountsAuthMetadata } from '@/lib/grpc/accounts-client'
import { getAuthClient, getAuthMetadata } from '@/lib/grpc/auth-client'

// Cache platform role check for the duration of the request
export const getPlatformRole = cache(async () => {
  const client = getAuthClient()
  const metadata = await getAuthMetadata()
  
  return new Promise<any>((resolve, reject) => {
    client.getPlatformRole({}, metadata, (error, response) => {
      if (error) {reject(error)}
      else {resolve(response)}
    })
  })
})

// Cache platform settings for the duration of the request
export const getPlatformSettings = cache(async () => {
  const client = getAccountsClient()
  const metadata = await getAccountsAuthMetadata()
  
  return new Promise<any>((resolve, reject) => {
    client.getPlatformSettings({}, metadata, (error, response) => {
      if (error) {reject(error)}
      else {resolve(response)}
    })
  })
})

// Batch fetch platform data
export async function fetchPlatformData() {
  const [platformRole, platformSettings] = await Promise.all([
    getPlatformRole(),
    getPlatformSettings()
  ])
  
  // Only fetch additional data if user is super admin
  let platformUsers = null
  let platformInvitations = null
  let teamRequests = { requests: [] }
  
  if (platformRole.role === "super_admin") {
    const authClient = getAuthClient()
    const accountsClient = getAccountsClient()
    const authMetadata = await getAuthMetadata()
    const accountsMetadata = await getAccountsAuthMetadata()
    
    const [usersData, invitationsData, requestsData] = await Promise.all([
      new Promise<any>((resolve, reject) => {
        authClient.listPlatformUsers({ role: '' }, authMetadata, (error, response) => {
          if (error) {reject(error)}
          else {resolve(response)}
        })
      }),
      new Promise<any>((resolve, reject) => {
        accountsClient.listPlatformInvitations({ status: 'pending' }, accountsMetadata, (error, response) => {
          if (error) {
            console.error("Failed to list platform invitations:", error)
            resolve({ invitations: [] })
          } else {
            resolve(response)
          }
        })
      }),
      new Promise<any>((resolve, reject) => {
        accountsClient.listTeamRequests({ status: "pending" }, accountsMetadata, (error, response) => {
          if (error) {
            console.error("Failed to list team requests:", error)
            resolve({ requests: [] })
          } else {
            resolve(response)
          }
        })
      })
    ])
    
    platformUsers = usersData.users
    platformInvitations = invitationsData.invitations
    teamRequests = requestsData
  } else {
    // Non-super admins only get team requests
    const client = getAccountsClient()
    const metadata = await getAccountsAuthMetadata()
    
    teamRequests = await new Promise<any>((resolve, reject) => {
      client.listTeamRequests({ status: "pending" }, metadata, (error, response) => {
        if (error) {
          console.error("Failed to list team requests:", error)
          resolve({ requests: [] })
        } else {
          resolve(response)
        }
      })
    })
  }
  
  return {
    platformRole,
    platformSettings: platformSettings.settings,
    platformUsers,
    platformInvitations,
    teamRequests: teamRequests.requests
  }
}