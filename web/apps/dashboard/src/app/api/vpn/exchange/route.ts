import { NextRequest, NextResponse } from 'next/server'
import { SignJWT, jwtVerify } from 'jose'

/**
 * VPN Token Exchange API
 * Exchanges short temporary token for permanent VPN token + connection data
 * This is a public endpoint (no auth required - validated via temporary token)
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { token: temporaryToken } = body

    console.log('[VPN Exchange] Received token exchange request')

    if (!temporaryToken || typeof temporaryToken !== 'string') {
      console.error('[VPN Exchange] Missing or invalid token in request')
      return NextResponse.json({ error: 'Temporary token required' }, { status: 400 })
    }

    // Get JWT_SECRET (shared with backend Go API)
    const jwtSecret = process.env.JWT_SECRET
    if (!jwtSecret) {
      console.error('[VPN Exchange] JWT_SECRET environment variable not set')
      return NextResponse.json({ error: 'Server configuration error' }, { status: 500 })
    }

    // Verify temporary JWT token
    const secret = new TextEncoder().encode(jwtSecret)
    let tokenData: { sub?: string; email: string; name?: string; username?: string; roles?: string[]; isActive?: boolean; type: string }

    try {
      // Add clock tolerance of 5 minutes to handle clock skew between servers
      const { payload } = await jwtVerify(temporaryToken, secret, {
        clockTolerance: 300, // 5 minutes in seconds
      })
      console.log('[VPN Exchange] Token verified successfully')
      tokenData = payload as typeof tokenData

      // Validate token type
      if (tokenData.type !== 'vpn-temp') {
        console.error('[VPN Exchange] Invalid token type:', tokenData.type)
        return NextResponse.json(
          { error: 'Invalid token type' },
          { status: 401 }
        )
      }
    } catch (error) {
      console.error('[VPN Exchange] JWT verification failed:', error)
      return NextResponse.json(
        { error: 'Invalid or expired temporary token' },
        { status: 401 }
      )
    }

    // Generate permanent VPN token (1 year expiry)
    // This token includes all user info needed for backend validation
    const permanentToken = await new SignJWT({
      sub: tokenData.sub,
      email: tokenData.email,
      name: tokenData.name,
      username: tokenData.username,
      roles: tokenData.roles,
      isActive: tokenData.isActive,
      type: 'vpn-permanent',
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('1y') // 1 year
      .setIssuer('kloudlite-vpn')
      .setSubject(tokenData.email)
      .sign(secret)

    console.log('[VPN Exchange] Generated permanent token for user:', tokenData.username)

    // Return permanent token - the kltun daemon will use this for all tunnel server API calls
    return NextResponse.json({
      connection_token: permanentToken,
    })
  } catch (error) {
    console.error('Token exchange error:', error)
    return NextResponse.json(
      { error: 'Failed to exchange token' },
      { status: 500 }
    )
  }
}
