import { NextRequest, NextResponse } from 'next/server'
import { SignJWT, jwtVerify } from 'jose'
import { env } from '@/lib/env'

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

    // Get AUTH_SECRET (shared with backend)
    const authSecret = process.env.AUTH_SECRET
    if (!authSecret) {
      console.error('[VPN Exchange] AUTH_SECRET environment variable not set')
      return NextResponse.json({ error: 'Server configuration error' }, { status: 500 })
    }

    // Verify temporary JWT token
    const secret = new TextEncoder().encode(authSecret)
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

    // Fetch VPN configuration from backend
    // Generate a backend token from the user info
    const backendToken = await new SignJWT({
      sub: tokenData.sub,
      email: tokenData.email,
      name: tokenData.name,
      username: tokenData.username,
      roles: tokenData.roles,
      isActive: tokenData.isActive,
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('5m')
      .sign(secret)

    const backendUrl = env.apiUrl
    let vpnConfig: unknown

    console.log('[VPN Exchange] Fetching VPN config from backend:', backendUrl)

    try {
      const vpnResponse = await fetch(`${backendUrl}/api/v1/vpn/connect`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${backendToken}`,
          'Content-Type': 'application/json',
        },
      })

      console.log('[VPN Exchange] Backend response status:', vpnResponse.status)

      if (!vpnResponse.ok) {
        const errorText = await vpnResponse.text()
        console.error('[VPN Exchange] Backend VPN connect failed:', errorText)
        return NextResponse.json(
          { error: 'Failed to retrieve VPN configuration' },
          { status: vpnResponse.status }
        )
      }

      vpnConfig = await vpnResponse.json()
      console.log('[VPN Exchange] VPN config received successfully')
    } catch (error) {
      console.error('[VPN Exchange] Failed to fetch VPN config from backend:', error)
      return NextResponse.json(
        { error: 'Failed to connect to VPN service' },
        { status: 500 }
      )
    }

    // Return permanent token and VPN configuration
    return NextResponse.json({
      connection_token: permanentToken,
      vpn_config: vpnConfig,
    })
  } catch (error) {
    console.error('Token exchange error:', error)
    return NextResponse.json(
      { error: 'Failed to exchange token' },
      { status: 500 }
    )
  }
}
