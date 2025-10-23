import { NextRequest, NextResponse } from 'next/server'
import { getUserByInstallationKey, saveCertificate, type CertificateScope } from '@/lib/registration/supabase-storage-service'
import { generateCertificate, generateHostnames } from '@/lib/registration/cloudflare-certificates'

const CLOUDFLARE_DNS_DOMAIN = process.env.CLOUDFLARE_DNS_DOMAIN!

/**
 * Generate TLS certificates using Cloudflare Origin CA
 * Called by the deployment to generate certificates for HTTPS
 *
 * Request format:
 * {
 *   "installationKey": "abc-123",
 *   "scope": "installation" | "workmachine" | "workspace",  // Optional, defaults to "installation"
 *   "scopeIdentifier": "dev1",  // Required for workmachine/workspace scopes (wm-user or workspace name)
 *   "parentScopeIdentifier": "dev1"  // Required for workspace scope (wm-user)
 * }
 *
 * Examples:
 * - Installation cert: { "installationKey": "abc-123" }
 * - Workmachine cert: { "installationKey": "abc-123", "scope": "workmachine", "scopeIdentifier": "dev1" }
 * - Workspace cert: { "installationKey": "abc-123", "scope": "workspace", "scopeIdentifier": "workspace1", "parentScopeIdentifier": "dev1" }
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
        { status: 401 }
      )
    }

    const secretKey = authHeader.substring(7) // Remove "Bearer " prefix

    const body = await request.json()
    const {
      installationKey,
      scope = 'installation' as CertificateScope,
      scopeIdentifier,
      parentScopeIdentifier
    } = body

    if (!installationKey) {
      return NextResponse.json(
        { error: 'Installation key is required' },
        { status: 400 }
      )
    }

    // Validate scope-specific requirements
    if (scope === 'workmachine' && !scopeIdentifier) {
      return NextResponse.json(
        { error: 'scopeIdentifier (wm-user) is required for workmachine scope' },
        { status: 400 }
      )
    }

    if (scope === 'workspace' && (!scopeIdentifier || !parentScopeIdentifier)) {
      return NextResponse.json(
        { error: 'scopeIdentifier (workspace) and parentScopeIdentifier (wm-user) are required for workspace scope' },
        { status: 400 }
      )
    }

    // Look up user by installation key
    const user = await getUserByInstallationKey(installationKey)

    if (!user) {
      return NextResponse.json(
        { error: 'Invalid installation key' },
        { status: 404 }
      )
    }

    // Verify secret key matches
    if (user.secretKey !== secretKey) {
      return NextResponse.json(
        { error: 'Invalid secret key' },
        { status: 403 }
      )
    }

    // Check if user has subdomain assigned
    if (!user.subdomain) {
      return NextResponse.json(
        { error: 'User must have a subdomain assigned before generating certificates' },
        { status: 400 }
      )
    }

    console.log(`Generating ${scope} certificates for user: ${user.email}, subdomain: ${user.subdomain}`)
    if (scopeIdentifier) {
      console.log(`Scope identifier: ${scopeIdentifier}`)
    }
    if (parentScopeIdentifier) {
      console.log(`Parent scope identifier: ${parentScopeIdentifier}`)
    }

    // Generate hostnames for certificate
    const hostnames = generateHostnames(
      user.subdomain,
      CLOUDFLARE_DNS_DOMAIN,
      scope,
      scopeIdentifier,
      parentScopeIdentifier
    )
    console.log(`Certificate will cover hostnames:`, hostnames)

    // Generate certificate via Cloudflare Origin CA
    const cert = await generateCertificate(hostnames)

    if (!cert) {
      return NextResponse.json(
        { error: 'Failed to generate certificate' },
        { status: 500 }
      )
    }

    // Save certificate to database
    await saveCertificate({
      userEmail: user.email,
      cloudflareCertId: cert.id,
      certificate: cert.certificate,
      privateKey: cert.privateKey,
      hostnames: cert.hostnames,
      scope,
      scopeIdentifier: scopeIdentifier || null,
      parentScopeIdentifier: parentScopeIdentifier || null,
      validFrom: cert.validFrom,
      validUntil: cert.validUntil,
    })

    console.log(`Certificate saved for user: ${user.email}, scope: ${scope}`)

    const response = NextResponse.json({
      success: true,
      certificateId: cert.id,
      hostnames: cert.hostnames,
      scope,
      scopeIdentifier: scopeIdentifier || null,
      parentScopeIdentifier: parentScopeIdentifier || null,
      validFrom: cert.validFrom,
      validUntil: cert.validUntil,
      message: `Certificate generated successfully for ${scope} scope`
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Generate certificates error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
