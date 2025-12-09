import { NextRequest, NextResponse } from 'next/server'
import { env } from '@kloudlite/lib'

/**
 * Server-side validation of superadmin login token
 *
 * This proxies the validation request to the Go API server.
 * Note: Cookie/session management is now handled by NextAuth through the
 * credentials provider - this endpoint just validates and returns user info.
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

    // Return validation result - session management handled by NextAuth
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
