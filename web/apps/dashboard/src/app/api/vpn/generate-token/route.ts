import { NextRequest, NextResponse } from 'next/server'
import { auth } from '@/lib/auth'
import { SignJWT } from 'jose'

/**
 * VPN Temporary Token Generation API
 * Generates a short-lived compact JWT token (3 min) for kltun authentication
 * Requires user to be authenticated via NextAuth
 */
export async function POST(_request: NextRequest) {
  try {
    // Get authenticated session
    const session = await auth()

    if (!session?.user?.email) {
      return NextResponse.json({ error: 'Unauthorized - please sign in' }, { status: 401 })
    }

    // Get JWT secret
    const jwtSecret = process.env.AUTH_SECRET || process.env.NEXTAUTH_SECRET
    if (!jwtSecret) {
      console.error('AUTH_SECRET or NEXTAUTH_SECRET environment variable not set')
      return NextResponse.json({ error: 'Server configuration error' }, { status: 500 })
    }

    // Generate compact temporary JWT (3 minutes expiry)
    // Using minimal claims to keep token size small
    const secret = new TextEncoder().encode(jwtSecret)

    // Get backend token from session
    const backendToken = (session.user as { backendToken?: string }).backendToken

    console.log('[VPN Generate] Creating token for email:', session.user.email)
    console.log('[VPN Generate] Current time:', Math.floor(Date.now() / 1000))
    console.log('[VPN Generate] JWT secret length:', jwtSecret.length)
    console.log('[VPN Generate] Backend token available:', !!backendToken)

    const temporaryToken = await new SignJWT({
      e: session.user.email, // 'e' instead of 'email' to save bytes
      b: backendToken, // 'b' instead of 'backendToken'
      t: 'temp', // 't' instead of 'type'
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('3m') // 3 minutes
      .sign(secret)

    const expiresAt = Date.now() + 3 * 60 * 1000
    const expiresIn = 180 // seconds

    console.log('[VPN Generate] Token generated, expires at:', new Date(expiresAt).toISOString())

    // Get server URL from request headers (for multi-tenant support)
    const protocol = _request.headers.get('x-forwarded-proto') || 'https'
    const host = _request.headers.get('host') || _request.nextUrl.host
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
