import { NextRequest, NextResponse } from 'next/server'
import { auth } from '@/lib/auth'
import { SignJWT } from 'jose'

/**
 * VPN Temporary Token Generation API
 * Generates a short-lived compact JWT token (3 min) for kltun authentication
 * Requires user to be authenticated via NextAuth
 */
export async function POST(request: NextRequest) {
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
    const temporaryToken = await new SignJWT({
      e: session.user.email, // 'e' instead of 'email' to save bytes
      b: (session as any).backendToken, // 'b' instead of 'backendToken'
      t: 'temp', // 't' instead of 'type'
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('3m') // 3 minutes
      .sign(secret)

    const expiresAt = Date.now() + 3 * 60 * 1000
    const expiresIn = 180 // seconds

    return NextResponse.json({
      temporary_token: temporaryToken,
      expires_at: new Date(expiresAt).toISOString(),
      expires_in: expiresIn,
      server_url: process.env.NEXT_PUBLIC_WEB_URL || 'http://localhost:3000',
    })
  } catch (error) {
    console.error('Generate temporary token error:', error)
    return NextResponse.json({ error: 'Failed to generate temporary token' }, { status: 500 })
  }
}
