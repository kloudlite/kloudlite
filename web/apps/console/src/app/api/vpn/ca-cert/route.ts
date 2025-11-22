import { NextRequest, NextResponse } from 'next/server'
import { jwtVerify } from 'jose'

/**
 * CA Certificate API Route
 * Validates VPN tokens and proxies CA certificate requests to backend API server
 * Used by kltun CLI for getting CA certificates
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

    // Verify VPN token (both temporary and permanent tokens)
    let claims: {
      type?: string
      backendToken?: string
      t?: string
      b?: string
      e?: string
      email?: string
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

    // Extract backend token
    const backendToken = claims.backendToken || claims.b
    if (!backendToken) {
      return NextResponse.json({ error: 'Invalid token - missing backend token' }, { status: 401 })
    }

    // Get the backend API URL from environment
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

    // Forward the request to the backend Go API using the backend token
    const backendResponse = await fetch(`${backendUrl}/api/v1/vpn/ca-cert`, {
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
    console.error('CA cert proxy error:', error)
    return NextResponse.json(
      { error: 'Failed to get CA certificate' },
      { status: 500 }
    )
  }
}
