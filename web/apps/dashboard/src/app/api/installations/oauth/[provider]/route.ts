import { NextRequest, NextResponse } from 'next/server'

// Registration mode OAuth configuration - uses REGISTRATION_ prefixed env vars
// These are separate from dashboard OAuth to prevent conflicts
const OAUTH_CONFIGS = {
  github: {
    authUrl: 'https://github.com/login/oauth/authorize',
    clientId: process.env.REGISTRATION_GITHUB_CLIENT_ID!,
    scope: 'read:user user:email',
  },
  google: {
    authUrl: 'https://accounts.google.com/o/oauth2/v2/auth',
    clientId: process.env.REGISTRATION_GOOGLE_CLIENT_ID!,
    scope: 'openid email profile',
  },
  'microsoft-entra-id': {
    authUrl: `https://login.microsoftonline.com/${process.env.REGISTRATION_MICROSOFT_ENTRA_TENANT_ID}/oauth2/v2.0/authorize`,
    clientId: process.env.REGISTRATION_MICROSOFT_ENTRA_CLIENT_ID!,
    scope: 'openid email profile User.Read',
  },
}

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ provider: string }> },
) {
  const { provider: providerParam } = await params
  const provider = providerParam as keyof typeof OAUTH_CONFIGS
  const config = OAUTH_CONFIGS[provider]

  if (!config) {
    return NextResponse.json({ error: 'Invalid provider' }, { status: 400 })
  }

  const redirectUri = `${process.env.NEXTAUTH_URL}/api/installations/oauth/callback/${provider}`
  const state = crypto.randomUUID()

  // Build OAuth URL
  const oauthParams = new URLSearchParams({
    client_id: config.clientId,
    redirect_uri: redirectUri,
    scope: config.scope,
    state,
    response_type: 'code',
  })

  // Store state in cookie for verification
  const response = NextResponse.redirect(`${config.authUrl}?${oauthParams}`)
  response.cookies.set('oauth_state', state, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 60 * 10, // 10 minutes
  })

  return response
}
