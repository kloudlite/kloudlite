import { NextRequest, NextResponse } from 'next/server'
import { env } from '@kloudlite/lib'
import { cookies } from 'next/headers'
import { encode } from 'next-auth/jwt'

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

    // Create NextAuth-compatible JWT session using NextAuth's encode function
    // This ensures the token format and cookie name match NextAuth's expectations
    const sessionToken = await encode({
      token: {
        email: data.user.email,
        name: data.user.displayName || data.user.email,
        sub: data.user.email,
        roles: data.roles,
        isActive: true,
        provider: 'superadmin-login',
      },
      secret: process.env.AUTH_SECRET!,
      maxAge: 8 * 60 * 60, // 8 hours
    })

    // Set the NextAuth session cookie using NextAuth's default cookie name
    // NextAuth v5 uses 'authjs' prefix by default
    const cookieStore = await cookies()
    const cookieName = process.env.NODE_ENV === 'production'
      ? '__Secure-authjs.session-token'
      : 'authjs.session-token'

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
