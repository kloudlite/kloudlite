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

    if (!temporaryToken || typeof temporaryToken !== 'string') {
      return NextResponse.json({ error: 'Temporary token required' }, { status: 400 })
    }

    // Get JWT secret
    const jwtSecret = process.env.AUTH_SECRET || process.env.NEXTAUTH_SECRET
    if (!jwtSecret) {
      console.error('AUTH_SECRET or NEXTAUTH_SECRET environment variable not set')
      return NextResponse.json({ error: 'Server configuration error' }, { status: 500 })
    }

    // Verify temporary JWT token
    const secret = new TextEncoder().encode(jwtSecret)
    let tokenData: { e: string; b: string; t: string }

    try {
      // Add clock tolerance of 5 minutes to handle clock skew between servers
      const { payload } = await jwtVerify(temporaryToken, secret, {
        clockTolerance: 300, // 5 minutes in seconds
      })
      tokenData = payload as { e: string; b: string; t: string }

      // Validate token type
      if (tokenData.t !== 'temp') {
        return NextResponse.json(
          { error: 'Invalid token type' },
          { status: 401 }
        )
      }
    } catch (error) {
      console.error('JWT verification failed:', error)
      return NextResponse.json(
        { error: 'Invalid or expired temporary token' },
        { status: 401 }
      )
    }

    // Generate permanent VPN token (1 year expiry)
    const permanentToken = await new SignJWT({
      email: tokenData.e, // Use full 'email' key for permanent token
      type: 'permanent',
      backendToken: tokenData.b, // Use full 'backendToken' key for permanent token
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('1y') // 1 year
      .setIssuer('kloudlite-vpn')
      .setSubject(tokenData.e)
      .sign(secret)

    // Fetch VPN configuration from backend using the backend token
    const backendUrl = env.apiUrl
    let vpnConfig: any

    try {
      const vpnResponse = await fetch(`${backendUrl}/api/v1/vpn/connect`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${tokenData.b}`,
          'Content-Type': 'application/json',
        },
      })

      if (!vpnResponse.ok) {
        const errorText = await vpnResponse.text()
        console.error('Backend VPN connect failed:', errorText)
        return NextResponse.json(
          { error: 'Failed to retrieve VPN configuration' },
          { status: vpnResponse.status }
        )
      }

      vpnConfig = await vpnResponse.json()
    } catch (error) {
      console.error('Failed to fetch VPN config from backend:', error)
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
