'use server'

import { revalidatePath } from 'next/cache'
import { getServerSession } from 'next-auth'

import { type TeamRequest, type PlatformSettings } from '@/grpc/accounts.external'
import { getAuthOptions } from '@/lib/auth/get-auth-options'
import { getAccountsClient, getAuthMetadata } from '@/lib/grpc/accounts-client'
import { getAuthClient } from '@/lib/grpc/auth-client'

export async function createTeam(data: {
  slug: string
  displayName: string
  description?: string
  region: string
}) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  // First, try to create the team directly
  return new Promise((resolve, reject) => {
    client.createTeam(
      {
        ...data,
        description: data.description || ''
      },
      metadata,
      (error, response) => {
        if (error) {
          // If team creation requires approval, create a request instead
          if (error.message?.includes('requires approval')) {
            client.requestTeamCreation(
              {
                ...data,
                description: data.description || ''
              },
              metadata,
              (reqError, reqResponse) => {
                if (reqError) {
                  reject(reqError)
                } else {
                  revalidatePath('/teams')
                  revalidatePath('/overview')
                  // Return a special response indicating approval is pending
                  resolve({ teamId: reqResponse?.requestId, pending: true })
                }
              }
            )
          } else {
            reject(error)
          }
        } else {
          revalidatePath('/teams')
          revalidatePath('/overview')
          resolve(response)
        }
      }
    )
  })
}

export async function checkTeamSlugAvailability(slug: string) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  return new Promise<{ available: boolean; suggestions: string[] }>((resolve, reject) => {
    client.checkTeamSlugAvailability(
      { slug },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else {
          resolve({
            available: response?.result || false,
            suggestions: response?.suggestedSlugs || []
          })
        }
      }
    )
  })
}

export async function generateTeamSlugSuggestions(displayName: string) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  return new Promise<string[]>((resolve, reject) => {
    client.generateTeamSlugSuggestions(
      { displayName },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else {
          resolve(response?.suggestions || [])
        }
      }
    )
  })
}

export async function listUserTeams() {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  return new Promise((resolve, reject) => {
    client.listTeams(
      { userId: session.user.id },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else {
          resolve(response?.teams || [])
        }
      }
    )
  })
}

export async function listUserPendingTeamRequests() {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  // Get all team requests (pending)
  return new Promise<TeamRequest[]>((resolve, _reject) => {
    client.listTeamRequests(
      { status: 'pending' },
      metadata,
      (error, response) => {
        if (error) {
          // If user doesn't have permission or no requests, return empty array
          resolve([])
        } else {
          // Filter to only user's requests
          const userRequests = (response?.requests || []).filter(
            (req) => req.requestedBy === session.user.id || req.requestedByEmail === session.user.email
          )
          resolve(userRequests)
        }
      }
    )
  })
}


export async function approveTeamRequest(requestId: string) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  return new Promise((resolve, reject) => {
    client.approveTeamRequest(
      { requestId },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else {
          revalidatePath('/platform')
          resolve(response)
        }
      }
    )
  })
}

export async function rejectTeamRequest(requestId: string, reason: string) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  return new Promise((resolve, reject) => {
    client.rejectTeamRequest(
      { requestId, reason },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else {
          revalidatePath('/platform')
          resolve(response)
        }
      }
    )
  })
}

export async function updatePlatformUserRole(userId: string, role: string) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAuthClient()
  const metadata = await getAuthMetadata()

  return new Promise((resolve, reject) => {
    client.updatePlatformUserRole(
      { userId, role },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else {
          revalidatePath('/platform')
          resolve(response)
        }
      }
    )
  })
}

export async function updatePlatformSettings(settings: PlatformSettings) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  return new Promise((resolve, reject) => {
    client.updatePlatformSettings(
      { settings },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else {
          revalidatePath('/platform')
          resolve(response)
        }
      }
    )
  })
}

// Platform invitation actions

export async function invitePlatformUser(email: string, role: string) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  return new Promise((resolve, reject) => {
    client.invitePlatformUser(
      { email, role },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else if (response?.success) {
          revalidatePath('/platform')
          resolve(response)
        } else {
          reject(new Error(response?.error || 'Failed to invite user'))
        }
      }
    )
  })
}

export async function listPlatformInvitations(status?: string) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  return new Promise((resolve, reject) => {
    client.listPlatformInvitations(
      { status: status || '' },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else {
          resolve(response?.invitations || [])
        }
      }
    )
  })
}

export async function resendPlatformInvitation(invitationId: string) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  return new Promise((resolve, reject) => {
    client.resendPlatformInvitation(
      { invitationId },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else if (response?.success) {
          revalidatePath('/platform')
          resolve(response)
        } else {
          reject(new Error(response?.error || 'Failed to resend invitation'))
        }
      }
    )
  })
}

export async function cancelPlatformInvitation(invitationId: string) {
  const authOpts = await getAuthOptions()
  const session = await getServerSession(authOpts)
  if (!session?.user) {
    throw new Error('Unauthorized')
  }

  const client = getAccountsClient()
  const metadata = await getAuthMetadata()

  return new Promise((resolve, reject) => {
    client.cancelPlatformInvitation(
      { invitationId },
      metadata,
      (error, response) => {
        if (error) {
          reject(error)
        } else {
          revalidatePath('/platform')
          resolve(response)
        }
      }
    )
  })
}