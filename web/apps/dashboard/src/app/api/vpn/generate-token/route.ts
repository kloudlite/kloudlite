import { NextRequest, NextResponse } from 'next/server'
import { getSession } from '@/lib/get-session'
import { SignJWT } from 'jose'

/**
 * VPN Temporary Token Generation API
 * Generates a short-lived JWT token (3 min) for kltun authentication
 * Requires user to be authenticated via NextAuth
 */
export async function POST(request: NextRequest) {
  try {
    // Get authenticated session
    const session = await getSession()

    if (!session?.user?.email) {
      return NextResponse.json({ error: 'Unauthorized - please sign in' }, { status: 401 })
    }

    // Use JWT_SECRET (shared with backend Go API for JWT validation)
    const jwtSecret = process.env.JWT_SECRET
    if (!jwtSecret) {
      console.error('JWT_SECRET environment variable not set')
      return NextResponse.json({ error: 'Server configuration error' }, { status: 500 })
    }

    // Generate temporary JWT (3 minutes expiry) with user info
    // This token will be validated by the backend using the same JWT_SECRET
    const secret = new TextEncoder().encode(jwtSecret)

    const temporaryToken = await new SignJWT({
      sub: session.user.id,
      email: session.user.email,
      name: session.user.name,
      username: session.user.username,
      roles: session.user.roles,
      isActive: session.user.isActive,
      type: 'vpn-temp', // Mark as temporary VPN token
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('3m') // 3 minutes
      .sign(secret)

    const expiresAt = Date.now() + 3 * 60 * 1000
    const expiresIn = 180 // seconds

    // Get server URL from request headers (for multi-tenant support)
    const protocol = request.headers.get('x-forwarded-proto') || 'https'
    const host = request.headers.get('host') || request.nextUrl.host
    const serverUrl = `${protocol}://${host}`

    return NextResponse.json({
      temporary_token: temporaryToken,
      expires_at: new Date(expiresAt).toISOString(),
      expires_in: expiresIn,
      server_url: serverUrl,
    })
  } catch (error) {
    console.error('Generate temporary token error:', error)
    return NextResponse.json({ error: 'Failed to generate temporary token' }, { status: 500 })
  }
}
