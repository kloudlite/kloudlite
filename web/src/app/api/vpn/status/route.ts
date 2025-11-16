import { NextResponse } from 'next/server'
import { auth } from '@/lib/auth'

/**
 * VPN Status Check API Route
 * Tests VPN connectivity by attempting to reach a known cluster service
 * Returns connection status for the VPN indicator in the dashboard
 */
export async function GET() {
  try {
    // Check if user is authenticated
    const session = await auth()

    if (!session?.user?.email) {
      return NextResponse.json(
        { connected: false, message: 'Not authenticated' },
        { status: 401 }
      )
    }

    // Attempt to check VPN connectivity by calling the backend API
    // This will fail if VPN is not connected since the browser can't reach cluster services directly
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

    try {
      // Try to reach a lightweight endpoint with a short timeout
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 3000) // 3 second timeout

      const response = await fetch(`${backendUrl}/api/v1/health`, {
        method: 'GET',
        signal: controller.signal,
        headers: {
          'Content-Type': 'application/json',
        },
      })

      clearTimeout(timeoutId)

      if (response.ok) {
        return NextResponse.json({
          connected: true,
          message: 'VPN connected'
        })
      }

      return NextResponse.json({
        connected: false,
        message: 'Backend unreachable'
      })
    } catch (_error) {
      // Network error likely means VPN is not connected
      return NextResponse.json({
        connected: false,
        message: 'VPN not connected'
      })
    }
  } catch (_error) {
    console.error('VPN status check error:', _error)
    return NextResponse.json(
      { connected: false, message: 'Status check failed' },
      { status: 500 }
    )
  }
}
