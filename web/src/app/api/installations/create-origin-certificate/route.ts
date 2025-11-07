import { NextRequest, NextResponse } from 'next/server'
import {
  getInstallationByKey,
  getLatestCertificate,
  saveCertificate,
} from '@/lib/console/supabase-storage-service'
import { generateCertificate } from '@/lib/console/cloudflare-certificates'
import { createOrReuseEdgeCertificateForHostnames } from '@/lib/console/cloudflare-edge-certificates'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'

const CLOUDFLARE_DNS_DOMAIN = process.env.CLOUDFLARE_DNS_DOMAIN!

/**
 * Create origin certificate for a DomainRequest
 * Called by the DomainRequest controller to get or create a certificate
 *
 * Request format (JSON body):
 * {
 *   "installationKey": "abc-123",
 *   "domainRequestName": "wm-karthik", // Used as scopeIdentifier (composite key with installationId)
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
    const domainRequestName = body.domainRequestName as string | undefined
    const customHostnames = body.hostnames as string[] | undefined

    if (!installationKey) {
      return NextResponse.json({ error: 'Installation key is required' }, { status: 400 })
    }

    if (!domainRequestName) {
      return NextResponse.json({ error: 'domainRequestName is required' }, { status: 400 })
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

    // Check if origin certificate already exists for this (domainRequestName, installationId) composite key
    // Using 'workmachine' scope for now, can be made dynamic later
    const existingCert = await getLatestCertificate(
      installation.id,
      'workmachine',
      domainRequestName,
    )

    if (existingCert) {
      console.log(
        `Origin certificate already exists for domainRequest: ${domainRequestName}, installation: ${installation.id}, cert ID: ${existingCert.cloudflareCertId}`,
      )
      return NextResponse.json({
        success: true,
        certificate: existingCert.certificate,
        privateKey: existingCert.privateKey,
        certificateId: existingCert.cloudflareCertId,
        validFrom: existingCert.validFrom,
        validUntil: existingCert.validUntil,
        message: 'Origin certificate already exists for this DomainRequest',
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

    // Store origin certificate in tls_certificates table with domainRequest scope
    // Key: (installationId, scope='workmachine', scopeIdentifier=domainRequestName)
    await saveCertificate({
      installationId: installation.id,
      cloudflareCertId: originCert.id,
      certificate: originCert.certificate,
      privateKey: originCert.privateKey,
      hostnames: originCert.hostnames,
      scope: 'workmachine', // Using workmachine scope for DomainRequest certificates
      scopeIdentifier: domainRequestName, // DomainRequest name as identifier
      parentScopeIdentifier: null,
      validFrom: originCert.validFrom,
      validUntil: originCert.validUntil,
    })

    console.log(
      `Origin certificate saved for domainRequest: ${domainRequestName}, installation: ${installation.id}`,
    )

    // Create edge certificate for the same hostnames as the origin certificate
    // This allows Cloudflare to terminate TLS for browser-to-Cloudflare connections
    const edgeCertId = await createOrReuseEdgeCertificateForHostnames(
      installation.id,
      originCert.hostnames,
      domainRequestName
    )
    if (edgeCertId) {
      console.log(`Edge certificate created/reused for domainRequest: ${domainRequestName}, certId: ${edgeCertId}`)
    } else {
      console.warn(`Failed to create edge certificate for domainRequest: ${domainRequestName}`)
    }

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
