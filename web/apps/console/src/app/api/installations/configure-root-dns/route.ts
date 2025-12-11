import { NextRequest, NextResponse } from 'next/server'
import {
  getInstallationByKey,
  updateInstallationRootDns,
  markDeploymentReady,
} from '@/lib/console/supabase-storage-service'
import {
  CLOUDFLARE_DNS_DOMAIN,
  createCnameRecord,
  createDnsRecord,
} from '@/lib/console/cloudflare-dns'

export const runtime = 'nodejs'

/**
 * Configure root DNS for an installation
 *
 * Supports two modes:
 * 1. CNAME mode: Creates a CNAME record pointing to load balancer DNS
 *    - {subdomain}.khost.dev -> ALB DNS (DNS-only, load balancer handles TLS)
 * 2. A record mode: Creates an A record pointing directly to IP
 *    - {subdomain}.khost.dev -> IP address
 *
 * Request format:
 * {
 *   "installationKey": "abc-123",
 *   "target": "kl-xxx-alb-123456789.us-east-1.elb.amazonaws.com",  // or "1.2.3.4"
 *   "type": "cname" | "a"  // "cname" for load balancers, "a" for direct IPs
 * }
 */
export async function POST(request: NextRequest) {
  try {
    const authHeader = request.headers.get('authorization')
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return NextResponse.json(
        { error: 'Missing or invalid authorization header' },
        { status: 401 },
      )
    }

    const secretKey = authHeader.substring(7)
    const body = await request.json()
    const { installationKey, target, type } = body

    if (!installationKey) {
      return NextResponse.json({ error: 'installationKey is required' }, { status: 400 })
    }

    if (!target) {
      return NextResponse.json({ error: 'target is required' }, { status: 400 })
    }

    if (!type || !['cname', 'a'].includes(type)) {
      return NextResponse.json({ error: 'type must be "cname" or "a"' }, { status: 400 })
    }

    const installation = await getInstallationByKey(installationKey)
    if (!installation) {
      return NextResponse.json({ error: 'Invalid installation key' }, { status: 404 })
    }

    if (installation.secretKey !== secretKey) {
      return NextResponse.json({ error: 'Invalid secret key' }, { status: 403 })
    }

    if (!installation.subdomain) {
      return NextResponse.json(
        { error: 'Installation must have a subdomain assigned' },
        { status: 400 },
      )
    }

    const fullDomain = `${installation.subdomain}.${CLOUDFLARE_DNS_DOMAIN}`
    let recordId: string | null = null

    if (type === 'cname') {
      // CNAME mode: for load balancers (AWS ALB, etc.)
      // Use proxied=false (DNS-only) since load balancer handles TLS termination
      console.log(`Creating root CNAME: ${fullDomain} -> ${target} (DNS-only)`)
      recordId = await createCnameRecord(fullDomain, target, false)
    } else {
      // A record mode: for direct IP addresses
      // Use proxied=false for direct access
      console.log(`Creating root A record: ${fullDomain} -> ${target} (DNS-only)`)
      recordId = await createDnsRecord(fullDomain, target, false)
    }

    if (!recordId) {
      return NextResponse.json(
        { error: 'Failed to create DNS record' },
        { status: 500 },
      )
    }

    // Store DNS info in installation record
    await updateInstallationRootDns(installation.id, target, type, recordId)

    // Mark deployment as ready - this changes status from "Configuring" to "Active"
    await markDeploymentReady(installation.id, true)

    const response = NextResponse.json({
      success: true,
      domain: fullDomain,
      target,
      type,
      recordId,
      proxied: false,
    })

    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Configure root DNS error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
