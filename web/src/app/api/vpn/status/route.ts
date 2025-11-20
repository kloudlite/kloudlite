import { NextResponse } from 'next/server'
import { auth } from '@/lib/auth'
import { getSession } from '@/lib/get-session'

/**
 * VPN Status Check API Route
 * Tests VPN connectivity by attempting to reach the VPN health check endpoint
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

    // Get user session to extract subdomain
    const userSession = await getSession()
    const subdomain = userSession?.user?.subdomain

    if (!subdomain) {
      return NextResponse.json(
        { connected: false, message: 'No subdomain configured' },
        { status: 400 }
      )
    }

    // Get base domain from environment
    const baseDomain = process.env.CLOUDFLARE_DNS_DOMAIN || 'khost.dev'

    // Construct VPN check URL
    const vpnCheckUrl = `https://vpn-check.${subdomain}.${baseDomain}`

    try {
      // Try to reach the VPN check endpoint with a short timeout
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 3000) // 3 second timeout

      const response = await fetch(vpnCheckUrl, {
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
        message: 'VPN check endpoint unreachable'
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
