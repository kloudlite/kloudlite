import { NextRequest, NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { requireInstallationAccess } from '@/lib/console/authorization'
import { getInstallationById } from '@/lib/console/storage'
import dns from 'node:dns'
import https from 'node:https'

export const runtime = 'nodejs'

// Use Cloudflare DNS (1.1.1.1) to avoid stale negative cache from system resolver.
// Newly created Cloudflare DNS records may not resolve via the local resolver for minutes
// due to negative NXDOMAIN caching, but 1.1.1.1 picks them up immediately.
const resolver = new dns.Resolver()
resolver.setServers(['1.1.1.1', '1.0.0.1'])

function resolveHostname(hostname: string): Promise<string> {
  return new Promise((resolve, reject) => {
    resolver.resolve4(hostname, (err, addresses) => {
      if (err) reject(err)
      else resolve(addresses[0])
    })
  })
}

function httpsHead(ip: string, hostname: string, timeoutMs: number): Promise<{ status: number }> {
  return new Promise((resolve, reject) => {
    const req = https.request(
      {
        hostname: ip,
        port: 443,
        path: '/',
        method: 'HEAD',
        headers: { Host: hostname },
        servername: hostname, // TLS SNI — Cloudflare needs this to route correctly
        timeout: timeoutMs,
      },
      (res) => {
        res.resume() // drain response
        resolve({ status: res.statusCode ?? 0 })
      },
    )
    req.on('timeout', () => {
      req.destroy()
      reject(new Error('timeout'))
    })
    req.on('error', reject)
    req.end()
  })
}

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
    const { id } = await params

    // Verify access via org membership
    try {
      await requireInstallationAccess(id)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unauthorized'
      if (message.includes('No session')) return apiError('Not authenticated', 401)
      if (message.includes('Not found')) return apiError('Installation not found', 404)
      return apiError('Forbidden', 403)
    }

    // Get installation details
    const installation = await getInstallationById(id)

    if (!installation) {
      return apiError('Installation not found', 404)
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
    const hostname = `${installation.subdomain}.${domain}`
    const url = `https://${hostname}`

    // Resolve via Cloudflare DNS (1.1.1.1) to bypass stale system resolver cache
    const startTime = Date.now()
    let resolvedIp: string
    try {
      resolvedIp = await resolveHostname(hostname)
    } catch {
      const latency = Date.now() - startTime
      return NextResponse.json({
        active: false,
        reason: 'dns_failed',
        latency,
        message: `DNS resolution failed for ${hostname} via 1.1.1.1`
      })
    }

    // Ping using resolved IP with correct SNI so Cloudflare routes properly
    try {
      const { status } = await httpsHead(resolvedIp, hostname, 5000)
      const latency = Date.now() - startTime

      if (status >= 200 && status < 500) {
        // Any non-5xx response means the server is up (redirects, auth-required, etc.)
        return NextResponse.json({
          active: true,
          latency,
          status,
          url
        })
      } else {
        return NextResponse.json({
          active: false,
          reason: 'error_response',
          status,
          latency,
          message: `Server returned ${status}`
        })
      }
    } catch (fetchError) {
      const latency = Date.now() - startTime

      if (fetchError instanceof Error && fetchError.message === 'timeout') {
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
    return apiError('Internal server error', 500)
  }
}
