import { NextRequest, NextResponse } from 'next/server'
import { env } from '@/lib/env'

/**
 * Server-side validation of superadmin login token
 *
 * This proxies the validation request to the Go API server
 * and handles the response to set up the session securely
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

    if (!response.ok) {
      const error = await response.json()
      return NextResponse.json(
        { error: error.error || 'Token validation failed' },
        { status: response.status }
      )
    }

    const data = await response.json()

    // Return the validation response
    // The client will handle storing the JWT and user data
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error validating superadmin login token:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
