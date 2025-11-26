import { NextRequest, NextResponse } from 'next/server'
import { env } from '@kloudlite/lib'
import { cookies } from 'next/headers'
import { SignJWT } from 'jose'

/**
 * Server-side validation of superadmin login token
 *
 * This proxies the validation request to the Go API server
 * and creates a NextAuth-compatible session cookie
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { token } = body

    if (!token) {
      return NextResponse.json(
        { error: 'Token is required' },
        { status: 400 }
      )
    }

    // Call the Go API server to validate the token
    const response = await fetch(`${env.apiUrl}/api/v1/superadmin-login/validate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ token }),
    })

    const responseText = await response.text()

    if (!response.ok) {
      console.error('Token validation failed:', response.status, responseText)
      try {
        const error = JSON.parse(responseText)
        return NextResponse.json(
          { error: error.error || 'Token validation failed' },
          { status: response.status }
        )
      } catch {
        return NextResponse.json(
          { error: 'Token validation failed' },
          { status: response.status }
        )
      }
    }

    let data
    try {
      data = JSON.parse(responseText)
    } catch {
      console.error('Failed to parse response:', responseText)
      return NextResponse.json(
        { error: 'Invalid response from server' },
        { status: 500 }
      )
    }

    if (!data.valid || !data.user) {
      return NextResponse.json(
        { error: 'Invalid token' },
        { status: 401 }
      )
    }

    // Create NextAuth-compatible JWT session using shared secret
    const secret = new TextEncoder().encode(process.env.AUTH_SECRET)

    // Create a session token that mimics NextAuth's JWT structure
    const sessionToken = await new SignJWT({
      email: data.user.email,
      name: data.user.displayName || data.user.email,
      sub: data.user.email, // Subject (user ID)
      roles: data.roles,
      isActive: true,
      provider: 'superadmin-login',
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('8h') // Same as typical NextAuth session
      .sign(secret)

    // Set the NextAuth session cookie
    const cookieStore = await cookies()
    const cookieName = process.env.NODE_ENV === 'production'
      ? '__Secure-next-auth.session-token'
      : 'next-auth.session-token'

    cookieStore.set(cookieName, sessionToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 8 * 60 * 60, // 8 hours
      path: '/',
    })

    // Return success
    return NextResponse.json({
      success: true,
      user: data.user,
      roles: data.roles,
    })
  } catch (error) {
    console.error('Error validating superadmin login token:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
