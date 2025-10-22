import { NextRequest, NextResponse } from 'next/server'
import { cookies } from 'next/headers'
import { SignJWT } from 'jose'
import { saveUserRegistration, getUserByEmail, type UserRegistration } from '@/lib/registration/storage-service'

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
  'azure-ad': {
    tokenUrl: `https://login.microsoftonline.com/${process.env.REGISTRATION_MICROSOFT_ENTRA_TENANT_ID}/oauth2/v2.0/token`,
    userUrl: 'https://graph.microsoft.com/v1.0/me',
    clientId: process.env.REGISTRATION_MICROSOFT_ENTRA_CLIENT_ID!,
    clientSecret: process.env.REGISTRATION_MICROSOFT_ENTRA_CLIENT_SECRET!,
  },
}

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ provider: string }> }
) {
  const searchParams = request.nextUrl.searchParams
  const code = searchParams.get('code')
  const state = searchParams.get('state')
  const error = searchParams.get('error')

  if (error) {
    return NextResponse.redirect(new URL(`/register?error=${error}`, request.url))
  }

  if (!code || !state) {
    return NextResponse.redirect(new URL('/register?error=missing_params', request.url))
  }

  // Verify state
  const cookieStore = await cookies()
  const storedState = cookieStore.get('oauth_state')?.value

  if (!storedState || storedState !== state) {
    return NextResponse.redirect(new URL('/register?error=invalid_state', request.url))
  }

  const { provider: providerParam } = await params
  const provider = providerParam as keyof typeof OAUTH_CONFIGS
  const config = OAUTH_CONFIGS[provider]

  if (!config) {
    return NextResponse.redirect(new URL('/register?error=invalid_provider', request.url))
  }

  let userData: any = {}

  try {
    // Exchange authorization code for access token
    const redirectUri = `${process.env.NEXTAUTH_URL}/api/register/oauth/callback/${provider}`
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

    userData = await userResponse.json()
  } catch (err) {
    console.error('OAuth exchange error:', err)
    return NextResponse.redirect(new URL('/register?error=oauth_exchange_failed', request.url))
  }

  // Extract user data from OAuth response
  const email = userData.email || userData.login || userData.userPrincipalName
  const name = userData.name || userData.login || userData.displayName
  const userId = `${provider}-${userData.id || email}` // Unique user ID combining provider and their ID

  if (!email) {
    console.error('No email found in OAuth response')
    return NextResponse.redirect(new URL('/register?error=no_email', request.url))
  }

  // Check if user already exists by email (primary key)
  let existingUser = await getUserByEmail(email)

  let userRegistration: UserRegistration

  if (existingUser) {
    // User already exists - reuse existing installation key
    console.log('Existing user found:', email, 'with installation key:', existingUser.installationKey)
    console.log('Reusing existing installation. No new keys will be generated.')

    // Update user information that might have changed
    existingUser.name = name // Update name in case it changed in OAuth provider
    existingUser.lastHealthCheck = new Date().toISOString() // Track last login

    // Save updated user information back to storage
    try {
      await saveUserRegistration(existingUser)
      console.log('Updated user information for:', email)
    } catch (error) {
      console.error('Failed to update user registration:', error)
    }

    userRegistration = existingUser
  } else {
    // New user - generate installation key ONLY (secret key generated on first deployment verification)
    console.log('New user registration:', email)
    const installationKey = crypto.randomUUID()
    console.log('Generated new installation key:', installationKey)
    console.log('Secret key will be generated when deployment first verifies')

    userRegistration = {
      userId,
      email,
      name,
      provider: provider as 'github' | 'google' | 'azure-ad',
      registeredAt: new Date().toISOString(),
      installationKey,
      hasCompletedInstallation: false, // Will be set to true when deployment verifies and gets secret
      lastHealthCheck: new Date().toISOString(),
    }

    try {
      await saveUserRegistration(userRegistration)
      console.log('Saved new user registration for:', email)
    } catch (error) {
      console.error('Failed to save user registration:', error)
      // Continue anyway - user can still proceed with installation
    }
  }

  // Create a JWT session token with user data
  const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
  const token = await new SignJWT({
    provider: userRegistration.provider,
    email: userRegistration.email,
    name: userRegistration.name,
    image: userData.avatar_url || userData.picture,
    installationKey: userRegistration.installationKey,
    userId: userRegistration.userId,
  })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('30d')
    .sign(secret)

  // Set session cookie and redirect to installation instructions
  const response = NextResponse.redirect(new URL('/register/install', request.url))
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
