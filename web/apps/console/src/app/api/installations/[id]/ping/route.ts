import { NextRequest, NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationById } from '@/lib/console/supabase-storage-service'

export const runtime = 'nodejs'

/**
 * Check if an installation's dashboard is reachable (active)
 * GET /api/installations/[id]/ping
 *
 * Returns:
 * - active: boolean - whether the dashboard URL is reachable
 * - reason: string - why it's not active (if applicable)
 * - latency: number - response time in ms (if active)
 */
export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const session = await getRegistrationSession()

    if (!session?.user) {
      return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })
    }

    const { id } = await params

    // Get installation
    const installation = await getInstallationById(id)

    if (!installation) {
      return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
    }

    // Verify ownership
    if (installation.userId !== session.user.id) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 403 })
    }

    // Check if installation has a valid subdomain and is deployment ready
    if (!installation.subdomain) {
      return NextResponse.json({
        active: false,
        reason: 'no_subdomain',
        message: 'No subdomain configured'
      })
    }

    if (!installation.deploymentReady) {
      return NextResponse.json({
        active: false,
        reason: 'not_ready',
        message: 'Deployment not ready yet'
      })
    }

    // Validate subdomain is not a placeholder
    if (installation.subdomain === '0.0.0.0' || installation.subdomain.includes('0.0.0.0')) {
      return NextResponse.json({
        active: false,
        reason: 'invalid_subdomain',
        message: 'Invalid subdomain'
      })
    }

    // Build the dashboard URL
    const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'
    const url = `https://${installation.subdomain}.${domain}`

    // Ping the URL with a timeout
    const startTime = Date.now()
    try {
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 5000) // 5 second timeout

      const response = await fetch(url, {
        method: 'HEAD',
        signal: controller.signal,
        redirect: 'follow',
      })

      clearTimeout(timeoutId)

      const latency = Date.now() - startTime

      if (response.ok || response.status === 401 || response.status === 403) {
        // 401/403 means the server is up but requires auth - still considered active
        return NextResponse.json({
          active: true,
          latency,
          status: response.status,
          url
        })
      } else {
        return NextResponse.json({
          active: false,
          reason: 'error_response',
          status: response.status,
          latency,
          message: `Server returned ${response.status}`
        })
      }
    } catch (fetchError) {
      const latency = Date.now() - startTime

      if (fetchError instanceof Error && fetchError.name === 'AbortError') {
        return NextResponse.json({
          active: false,
          reason: 'timeout',
          latency,
          message: 'Request timed out (5s)'
        })
      }

      return NextResponse.json({
        active: false,
        reason: 'unreachable',
        latency,
        message: 'Could not connect to dashboard'
      })
    }
  } catch (error) {
    console.error('Error in ping endpoint:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
