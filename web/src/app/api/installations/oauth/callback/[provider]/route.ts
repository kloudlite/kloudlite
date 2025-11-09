import { NextRequest, NextResponse } from 'next/server'
import { cookies } from 'next/headers'
import { SignJWT } from 'jose'
import {
  saveUserRegistration,
  getUserByEmail,
  type UserRegistration,
} from '@/lib/console/supabase-storage-service'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'
// Registration mode OAuth configuration - uses REGISTRATION_ prefixed env vars
const OAUTH_CONFIGS = {
  github: {
    tokenUrl: 'https://github.com/login/oauth/access_token',
    userUrl: 'https://api.github.com/user',
    clientId: process.env.REGISTRATION_GITHUB_CLIENT_ID!,
    clientSecret: process.env.REGISTRATION_GITHUB_CLIENT_SECRET!,
  },
  google: {
    tokenUrl: 'https://oauth2.googleapis.com/token',
    userUrl: 'https://www.googleapis.com/oauth2/v2/userinfo',
    clientId: process.env.REGISTRATION_GOOGLE_CLIENT_ID!,
    clientSecret: process.env.REGISTRATION_GOOGLE_CLIENT_SECRET!,
  },
  'microsoft-entra-id': {
    tokenUrl: `https://login.microsoftonline.com/${process.env.REGISTRATION_MICROSOFT_ENTRA_TENANT_ID}/oauth2/v2.0/token`,
    userUrl: 'https://graph.microsoft.com/v1.0/me',
    clientId: process.env.REGISTRATION_MICROSOFT_ENTRA_CLIENT_ID!,
    clientSecret: process.env.REGISTRATION_MICROSOFT_ENTRA_CLIENT_SECRET!,
  },
}

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ provider: string }> },
) {
  // Use environment variable for base URL to avoid internal container URLs in redirects
  const baseUrl = process.env.NEXTAUTH_URL || process.env.NEXT_PUBLIC_BASE_URL || request.url

  const searchParams = request.nextUrl.searchParams
  const code = searchParams.get('code')
  const state = searchParams.get('state')
  const error = searchParams.get('error')

  if (error) {
    return NextResponse.redirect(new URL(`/installations/login?error=${error}`, baseUrl))
  }

  if (!code || !state) {
    return NextResponse.redirect(new URL('/installations/login?error=missing_params', baseUrl))
  }

  // Verify state
  const cookieStore = await cookies()
  const storedState = cookieStore.get('oauth_state')?.value

  if (!storedState || storedState !== state) {
    return NextResponse.redirect(new URL('/installations/login?error=invalid_state', baseUrl))
  }

  const { provider: providerParam } = await params
  const provider = providerParam as keyof typeof OAUTH_CONFIGS
  const config = OAUTH_CONFIGS[provider]

  if (!config) {
    return NextResponse.redirect(new URL('/installations/login?error=invalid_provider', baseUrl))
  }

  // OAuth user data interfaces for different providers
  interface GitHubUser {
    id: number
    login: string
    email: string | null
    name: string
    avatar_url: string
  }

  interface GitHubEmail {
    email: string
    primary: boolean
    verified: boolean
    visibility: string | null
  }

  interface GoogleUser {
    id: string
    email: string
    name: string
    picture: string
  }

  interface AzureADUser {
    id: string
    userPrincipalName: string
    displayName: string
    mail?: string
    otherMails?: string[]
  }

  type OAuthUserData = GitHubUser | GoogleUser | AzureADUser

  let userData: OAuthUserData | Record<string, never> = {}

  try {
    // Exchange authorization code for access token
    const redirectUri = `${process.env.NEXTAUTH_URL}/api/installations/oauth/callback/${provider}`
    const tokenResponse = await fetch(config.tokenUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        Accept: 'application/json',
      },
      body: new URLSearchParams({
        client_id: config.clientId,
        client_secret: config.clientSecret,
        code,
        redirect_uri: redirectUri,
        grant_type: 'authorization_code',
      }),
    })

    if (!tokenResponse.ok) {
      throw new Error('Failed to exchange code for token')
    }

    const tokenData = await tokenResponse.json()
    const accessToken = tokenData.access_token

    // Fetch user information
    const userResponse = await fetch(config.userUrl, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
        Accept: 'application/json',
      },
    })

    if (!userResponse.ok) {
      throw new Error('Failed to fetch user data')
    }

    userData = (await userResponse.json()) as OAuthUserData

    // For GitHub, if email is not in the user response (private email), fetch from /user/emails
    if (provider === 'github' && 'login' in userData && !userData.email) {
      console.log('GitHub email is private, fetching from /user/emails')
      try {
        const emailsResponse = await fetch('https://api.github.com/user/emails', {
          headers: {
            Authorization: `Bearer ${accessToken}`,
            Accept: 'application/json',
          },
        })

        if (emailsResponse.ok) {
          const emails = (await emailsResponse.json()) as GitHubEmail[]
          // Find the primary verified email
          const primaryEmail = emails.find((e) => e.primary && e.verified)
          if (primaryEmail) {
            userData.email = primaryEmail.email
            console.log('Successfully retrieved primary email from /user/emails')
          }
        }
      } catch (error) {
        console.error('Failed to fetch GitHub emails:', error)
      }
    }

    // For Google, verify email is present
    if (provider === 'google' && 'picture' in userData && !userData.email) {
      console.error('Google OAuth did not return email')
      throw new Error('No email found in Google OAuth response')
    }

    // For Microsoft, ensure we have an email
    if (provider === 'microsoft-entra-id' && 'userPrincipalName' in userData) {
      if (!userData.mail && (!userData.otherMails || userData.otherMails.length === 0)) {
        console.log(
          'Microsoft email not in mail/otherMails, will extract from userPrincipalName',
        )
      }
    }
  } catch (err) {
    console.error('OAuth exchange error:', err)
    return NextResponse.redirect(
      new URL('/installations/login?error=oauth_exchange_failed', baseUrl),
    )
  }

  // Extract email from Microsoft userPrincipalName (handles guest users)
  // Guest users have format: original_email#EXT#@tenant.onmicrosoft.com
  const extractEmailFromUPN = (upn: string): string => {
    if (upn.includes('#EXT#')) {
      // Extract the part before #EXT# and replace last underscore with @
      const beforeExt = upn.split('#EXT#')[0]
      // Replace the last underscore with @ to restore original email
      const lastUnderscoreIndex = beforeExt.lastIndexOf('_')
      if (lastUnderscoreIndex !== -1) {
        return (
          beforeExt.substring(0, lastUnderscoreIndex) +
          '@' +
          beforeExt.substring(lastUnderscoreIndex + 1)
        )
      }
    }
    return upn
  }

  // Extract user data from OAuth response based on provider
  const getUserInfo = (data: OAuthUserData, prov: typeof provider) => {
    if (prov === 'github' && 'login' in data) {
      return {
        email: data.email,
        name: data.name || data.login,
        id: data.id,
        avatar: data.avatar_url,
      }
    } else if (prov === 'google' && 'picture' in data) {
      return {
        email: data.email,
        name: data.name,
        id: data.id,
        avatar: data.picture,
      }
    } else if (prov === 'microsoft-entra-id' && 'userPrincipalName' in data) {
      // Try to get email from: mail field, otherMails array, or parse userPrincipalName
      let email = data.mail
      if (!email && data.otherMails && data.otherMails.length > 0) {
        email = data.otherMails[0]
      }
      if (!email) {
        email = extractEmailFromUPN(data.userPrincipalName)
      }

      return {
        email,
        name: data.displayName,
        id: data.id,
        avatar: undefined,
      }
    }
    throw new Error('Invalid provider or user data')
  }

  const userInfo = getUserInfo(userData, provider)
  const email = userInfo.email
  const name = userInfo.name
  const userId = `${provider}-${userInfo.id || email}` // Unique user ID combining provider and their ID

  if (!email) {
    console.error('No email found in OAuth response')
    return NextResponse.redirect(new URL('/installations/login?error=no_email', baseUrl))
  }

  // Check if user already exists by email
  const existingUser = await getUserByEmail(email)

  let userRegistration: UserRegistration

  if (existingUser) {
    // User already exists
    console.log('Existing user found:', email)

    // Add provider to array if not already present
    const currentProvider = (provider === 'microsoft-entra-id' ? 'azure-ad' : provider) as
      | 'github'
      | 'google'
      | 'azure-ad'
    if (!existingUser.providers.includes(currentProvider)) {
      existingUser.providers = [...existingUser.providers, currentProvider]
      console.log('Adding new provider:', currentProvider)

      try {
        await saveUserRegistration(existingUser)
        console.log('Updated providers array for:', email)
      } catch (error) {
        console.error('Failed to update providers:', error)
      }
    }

    userRegistration = existingUser
  } else {
    // New user - create user registration
    console.log('New user registration:', email)

    const normalizedProvider = (provider === 'microsoft-entra-id' ? 'azure-ad' : provider) as
      | 'github'
      | 'google'
      | 'azure-ad'
    userRegistration = {
      userId,
      email,
      name,
      providers: [normalizedProvider],
      registeredAt: new Date().toISOString(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    try {
      await saveUserRegistration(userRegistration)
      console.log('Saved new user registration for:', email)
    } catch (error) {
      console.error('Failed to save user registration:', error)
      // Continue anyway - user can retry
    }
  }

  // Create a JWT session token with user data
  const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
  const token = await new SignJWT({
    provider: provider, // Current provider used for login
    providers: userRegistration.providers, // All providers user has used
    email: userRegistration.email,
    name: userRegistration.name,
    image: userInfo.avatar,
    userId: userRegistration.userId,
  })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('30d')
    .sign(secret)

  // Always redirect to installations list - users click "New Installation" to create
  const redirectUrl = '/installations'
  console.log('Redirecting to installations list')

  // Set session cookie and redirect
  const response = NextResponse.redirect(new URL(redirectUrl, baseUrl))
  response.cookies.set('registration_session', token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 30 * 24 * 60 * 60, // 30 days
  })

  // Clear OAuth state cookie
  response.cookies.delete('oauth_state')

  return response
}
