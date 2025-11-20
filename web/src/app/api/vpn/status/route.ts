import { NextResponse } from 'next/server'
import { NextRequest } from 'next/server'
import { auth } from '@/lib/auth'

/**
 * VPN Status Check API Route
 * Tests VPN connectivity by attempting to reach the VPN health check endpoint
 * Returns connection status for the VPN indicator in the dashboard
 */
export async function GET(request: NextRequest) {
  try {
    // Check if user is authenticated
    const session = await auth()

    if (!session?.user?.email) {
      return NextResponse.json(
        { connected: false, message: 'Not authenticated' },
        { status: 401 }
      )
    }

    // Extract subdomain from current hostname
    // Expected format: subdomain.khost.dev or *.subdomain.khost.dev
    const hostname = request.headers.get('host') || ''
    const baseDomain = process.env.CLOUDFLARE_DNS_DOMAIN || 'khost.dev'

    // Parse subdomain from hostname
    // Examples:
    // - "test.khost.dev" -> "test"
    // - "console.test.khost.dev" -> "test"
    const hostParts = hostname.split('.')
    const baseParts = baseDomain.split('.')

    let subdomain: string | null = null

    if (hostParts.length > baseParts.length) {
      // Get the part before the base domain
      // For "console.test.khost.dev" with base "khost.dev", we want "test"
      subdomain = hostParts[hostParts.length - baseParts.length - 1]
    }

    if (!subdomain) {
      return NextResponse.json(
        { connected: false, message: 'Could not determine subdomain from hostname' },
        { status: 400 }
      )
    }

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
