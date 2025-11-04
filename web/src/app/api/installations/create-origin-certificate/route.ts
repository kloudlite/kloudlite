import { NextRequest, NextResponse } from 'next/server'
import {
  getInstallationByKey,
  updateInstallation,
} from '@/lib/console/supabase-storage-service'
import { generateCertificate } from '@/lib/console/cloudflare-certificates'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'

const CLOUDFLARE_DNS_DOMAIN = process.env.CLOUDFLARE_DNS_DOMAIN!

/**
 * Create origin certificate for an installation
 * Called by the DomainRequest controller when origin certificate doesn't exist
 *
 * Request format (JSON body):
 * {
 *   "installationKey": "abc-123",
 *   "hostnames": ["example.com", "*.example.com"] // Optional, defaults to ["subdomain.domain", "*.subdomain.domain"]
 * }
 *
 * Requires Authorization: Bearer <secret-key>
 */
export async function POST(request: NextRequest) {
  try {
    // Extract and validate bearer token
    const authHeader = request.headers.get('authorization')
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return NextResponse.json(
        { error: 'Missing or invalid authorization header' },
        { status: 401 },
      )
    }

    const secretKey = authHeader.substring(7) // Remove "Bearer " prefix

    // Parse request body
    const body = await request.json()
    const installationKey = body.installationKey
    const customHostnames = body.hostnames as string[] | undefined

    if (!installationKey) {
      return NextResponse.json({ error: 'Installation key is required' }, { status: 400 })
    }

    // Look up installation by installation key
    const installation = await getInstallationByKey(installationKey)

    if (!installation) {
      return NextResponse.json({ error: 'Invalid installation key' }, { status: 404 })
    }

    // Verify secret key matches
    if (installation.secretKey !== secretKey) {
      return NextResponse.json({ error: 'Invalid secret key' }, { status: 403 })
    }

    // Check if origin certificate already exists
    if (installation.originCertificate && installation.originPrivateKey) {
      console.log(
        `Origin certificate already exists for installation: ${installation.id}, cert ID: ${installation.originCertId}`,
      )
      return NextResponse.json({
        success: true,
        certificate: installation.originCertificate,
        privateKey: installation.originPrivateKey,
        certificateId: installation.originCertId,
        validFrom: installation.originCertValidFrom,
        validUntil: installation.originCertValidUntil,
        message: 'Origin certificate already exists',
      })
    }

    // Check if installation has a subdomain
    if (!installation.subdomain) {
      return NextResponse.json(
        { error: 'Installation does not have a subdomain assigned' },
        { status: 400 },
      )
    }

    // Determine hostnames for certificate
    let originCertHostnames: string[]
    if (customHostnames && customHostnames.length > 0) {
      // Use custom hostnames provided by DomainRequest
      originCertHostnames = customHostnames
      console.log(`Using custom origin certificate hostnames for installation: ${installation.id}`, customHostnames)
    } else {
      // Default to valid single-wildcard pattern (Cloudflare only allows ONE wildcard per hostname)
      originCertHostnames = [
        `${installation.subdomain}.${CLOUDFLARE_DNS_DOMAIN}`,
        `*.${installation.subdomain}.${CLOUDFLARE_DNS_DOMAIN}`,
      ]
      console.log(`Using default origin certificate hostnames for installation: ${installation.id}`, originCertHostnames)
    }

    const originCert = await generateCertificate(originCertHostnames)

    if (!originCert) {
      console.error(`Failed to generate origin certificate for installation: ${installation.id}`)
      return NextResponse.json(
        { error: 'Failed to generate origin certificate from Cloudflare' },
        { status: 500 },
      )
    }

    console.log(`Origin certificate generated: ${originCert.id}`)

    // Store origin certificate in installation
    await updateInstallation(installation.id, {
      originCertificate: originCert.certificate,
      originPrivateKey: originCert.privateKey,
      originCertId: originCert.id,
      originCertValidFrom: originCert.validFrom,
      originCertValidUntil: originCert.validUntil,
    })

    console.log(`Origin certificate saved to installation: ${installation.id}`)

    const response = NextResponse.json({
      success: true,
      certificate: originCert.certificate,
      privateKey: originCert.privateKey,
      certificateId: originCert.id,
      validFrom: originCert.validFrom,
      validUntil: originCert.validUntil,
      message: 'Origin certificate created successfully',
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Create origin certificate error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
