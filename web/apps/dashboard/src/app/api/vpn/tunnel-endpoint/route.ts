import { NextRequest, NextResponse } from 'next/server'
import { jwtVerify, SignJWT } from 'jose'

/**
 * Tunnel Endpoint API Route
 * Validates VPN tokens and returns the tunnel server endpoint (WorkMachine public IP:443)
 * Used by kltun CLI to connect to the user's tunnel server
 */
export async function GET(request: NextRequest) {
  try {
    // Get the authorization header from the incoming request
    const authHeader = request.headers.get('authorization')

    if (!authHeader) {
      return NextResponse.json(
        { error: 'Authorization header required' },
        { status: 401 }
      )
    }

    // Extract token from Bearer header
    const token = authHeader.replace('Bearer ', '')

    // Get JWT_SECRET (shared with backend Go API)
    const jwtSecret = process.env.JWT_SECRET
    if (!jwtSecret) {
      console.error('JWT_SECRET environment variable not set')
      return NextResponse.json({ error: 'Server configuration error' }, { status: 500 })
    }

    // Verify VPN token
    let claims: {
      sub?: string
      email?: string
      name?: string
      username?: string
      roles?: string[]
      isActive?: boolean
      type?: string
    }

    try {
      const secret = new TextEncoder().encode(jwtSecret)
      const { payload } = await jwtVerify(token, secret)
      claims = payload as typeof claims
    } catch (error) {
      console.error('Token verification failed:', error)
      return NextResponse.json(
        { error: 'Invalid or expired VPN token' },
        { status: 401 }
      )
    }

    // Validate token type
    if (claims.type !== 'vpn-temp' && claims.type !== 'vpn-permanent') {
      return NextResponse.json({ error: 'Invalid token type' }, { status: 401 })
    }

    // Generate a backend token from the VPN token claims
    const secret = new TextEncoder().encode(jwtSecret)
    const backendToken = await new SignJWT({
      sub: claims.sub,
      email: claims.email,
      name: claims.name,
      username: claims.username,
      roles: claims.roles,
      isActive: claims.isActive,
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('5m')
      .sign(secret)

    // Get the backend API URL from environment
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

    // Forward the request to the backend Go API
    const backendResponse = await fetch(`${backendUrl}/api/v1/vpn/tunnel-endpoint`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${backendToken}`,
        'Content-Type': 'application/json',
      },
    })

    // Get the response data
    const data = await backendResponse.json()

    // Return the backend response with the same status code
    return NextResponse.json(data, { status: backendResponse.status })
  } catch (error) {
    console.error('Tunnel endpoint proxy error:', error)
    return NextResponse.json(
      { error: 'Failed to get tunnel endpoint' },
      { status: 500 }
    )
  }
}
