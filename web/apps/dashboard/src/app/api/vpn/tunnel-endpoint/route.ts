import { NextRequest, NextResponse } from 'next/server'
import { jwtVerify } from 'jose'
import { getTunnelEndpoint } from '@/app/actions/vpn.actions'

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

    // Get JWT_SECRET
    const jwtSecret = process.env.JWT_SECRET || process.env.NEXTAUTH_SECRET
    if (!jwtSecret) {
      console.error('JWT_SECRET/NEXTAUTH_SECRET environment variable not set')
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

    // Get username from claims
    const username = claims.username
    if (!username) {
      return NextResponse.json({ error: 'Username not found in token' }, { status: 401 })
    }

    // Get tunnel endpoint using Server Action
    const result = await getTunnelEndpoint(username)

    if (!result.success) {
      // Determine appropriate status code based on error
      let statusCode = 500
      if (result.error?.includes('not found')) {
        statusCode = 404
      } else if (result.error?.includes('not running') || result.error?.includes('not ready')) {
        statusCode = 503
      }

      return NextResponse.json({ error: result.error }, { status: statusCode })
    }

    return NextResponse.json(result.data, { status: 200 })
  } catch (error) {
    console.error('Tunnel endpoint error:', error)
    return NextResponse.json(
      { error: 'Failed to get tunnel endpoint' },
      { status: 500 }
    )
  }
}
