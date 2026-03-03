import { NextRequest, NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import {
  getInstallationByKey,
  updateInstallationRootDns,
  markDeploymentReady,
} from '@/lib/console/storage'
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
 *    - {subdomain}.khost.dev -> ALB DNS
 * 2. A record mode: Creates an A record pointing directly to IP
 *    - {subdomain}.khost.dev -> IP address
 *
 * Proxy mode:
 * - proxied=true: Cloudflare handles TLS termination (Flexible SSL mode)
 * - proxied=false: DNS-only, origin handles TLS
 *
 * Request format:
 * {
 *   "installationKey": "abc-123",
 *   "target": "kl-xxx-alb-123456789.us-east-1.elb.amazonaws.com",  // or "1.2.3.4"
 *   "type": "cname" | "a",  // "cname" for load balancers, "a" for direct IPs
 *   "proxied": true | false  // optional, defaults to false
 * }
 */
export async function POST(request: NextRequest) {
  try {
    const authHeader = request.headers.get('authorization')
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return apiError('Missing or invalid authorization header', 401)
    }

    const secretKey = authHeader.substring(7)
    const body = await request.json()
    const { installationKey, target, type, proxied = false } = body

    if (!installationKey) {
      return apiError('installationKey is required', 400)
    }

    if (!target) {
      return apiError('target is required', 400)
    }

    if (!type || !['cname', 'a'].includes(type)) {
      return apiError('type must be "cname" or "a"', 400)
    }

    const installation = await getInstallationByKey(installationKey)
    if (!installation) {
      return apiError('Invalid installation key', 404)
    }

    if (installation.secretKey !== secretKey) {
      return apiError('Invalid secret key', 403)
    }

    if (!installation.subdomain) {
      return apiError('Installation must have a subdomain assigned', 400)
    }

    const fullDomain = `${installation.subdomain}.${CLOUDFLARE_DNS_DOMAIN}`
    let recordId: string | null = null

    if (type === 'cname') {
      // CNAME mode: for load balancers (AWS ALB, etc.)
      console.log(`Creating root CNAME: ${fullDomain} -> ${target} (proxied=${proxied})`)
      recordId = await createCnameRecord(fullDomain, target, proxied)
    } else {
      // A record mode: for direct IP addresses
      console.log(`Creating root A record: ${fullDomain} -> ${target} (proxied=${proxied})`)
      recordId = await createDnsRecord(fullDomain, target, proxied)
    }

    if (!recordId) {
      return apiError('Failed to create DNS record', 500)
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
      proxied,
    })

    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Configure root DNS error:', error)
    return apiError('Internal server error', 500)
  }
}
