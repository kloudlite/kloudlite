import { NextRequest, NextResponse } from 'next/server'
import { jwtVerify } from 'jose'

/**
 * VPN Connect API Route
 * Validates permanent VPN token and proxies requests to backend API server
 * Used by kltun CLI for establishing VPN connections with permanent tokens
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

    // Get JWT secret
    const jwtSecret = process.env.AUTH_SECRET || process.env.NEXTAUTH_SECRET
    if (!jwtSecret) {
      console.error('AUTH_SECRET or NEXTAUTH_SECRET environment variable not set')
      return NextResponse.json({ error: 'Server configuration error' }, { status: 500 })
    }

    // Verify permanent VPN token
    let claims: any
    try {
      const secret = new TextEncoder().encode(jwtSecret)
      const { payload } = await jwtVerify(token, secret, {
        issuer: 'kloudlite-vpn',
      })
      claims = payload
    } catch (error) {
      console.error('Token verification failed:', error)
      return NextResponse.json(
        { error: 'Invalid or expired VPN token' },
        { status: 401 }
      )
    }

    // Validate token type
    if (claims.type !== 'permanent') {
      return NextResponse.json({ error: 'Invalid token type' }, { status: 401 })
    }

    // Extract backend token from claims
    const backendToken = claims.backendToken
    if (!backendToken) {
      return NextResponse.json({ error: 'Invalid token - missing backend token' }, { status: 401 })
    }

    // Get the backend API URL from environment
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

    // Forward the request to the backend Go API using the backend token
    const backendResponse = await fetch(`${backendUrl}/api/v1/vpn/connect`, {
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
    console.error('VPN connect proxy error:', error)
    return NextResponse.json(
      { error: 'Failed to connect to VPN service' },
      { status: 500 }
    )
  }
}
