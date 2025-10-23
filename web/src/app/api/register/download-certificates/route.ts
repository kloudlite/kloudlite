import { NextRequest, NextResponse } from 'next/server'
import {
  getUserByInstallationKey,
  getLatestCertificate,
  type CertificateScope,
} from '@/lib/registration/supabase-storage-service'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'
/**
 * Download TLS certificates (certificate + private key)
 * Called by the deployment to download certificates
 *
 * Query params:
 *   installationKey: Installation key
 *   format: 'json' | 'pem' | 'bundle' (default: 'json')
 *   scope: 'installation' | 'workmachine' | 'workspace' (optional, filters by scope)
 *   scopeIdentifier: wm-user or workspace name (optional, required with scope=workmachine/workspace)
 *   parentScopeIdentifier: wm-user (optional, required with scope=workspace)
 *
 * Examples:
 *   - Get installation cert: /api/register/download-certificates?installationKey=abc-123&format=json
 *   - Get workmachine cert: /api/register/download-certificates?installationKey=abc-123&scope=workmachine&scopeIdentifier=dev1
 *   - Get workspace cert: /api/register/download-certificates?installationKey=abc-123&scope=workspace&scopeIdentifier=workspace1&parentScopeIdentifier=dev1
 *
 * Requires Authorization: Bearer <secret-key>
 */
export async function GET(request: NextRequest) {
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

    const searchParams = request.nextUrl.searchParams
    const installationKey = searchParams.get('installationKey')
    const format = searchParams.get('format') || 'json'
    const scope = searchParams.get('scope') as CertificateScope | null
    const scopeIdentifier = searchParams.get('scopeIdentifier')
    const parentScopeIdentifier = searchParams.get('parentScopeIdentifier')

    if (!installationKey) {
      return NextResponse.json({ error: 'Installation key is required' }, { status: 400 })
    }

    // Validate scope-specific requirements
    if (scope === 'workmachine' && !scopeIdentifier) {
      return NextResponse.json(
        { error: 'scopeIdentifier (wm-user) is required for workmachine scope' },
        { status: 400 },
      )
    }

    if (scope === 'workspace' && (!scopeIdentifier || !parentScopeIdentifier)) {
      return NextResponse.json(
        {
          error:
            'scopeIdentifier (workspace) and parentScopeIdentifier (wm-user) are required for workspace scope',
        },
        { status: 400 },
      )
    }

    // Look up user by installation key
    const user = await getUserByInstallationKey(installationKey)

    if (!user) {
      return NextResponse.json({ error: 'Invalid installation key' }, { status: 404 })
    }

    // Verify secret key matches
    if (user.secretKey !== secretKey) {
      return NextResponse.json({ error: 'Invalid secret key' }, { status: 403 })
    }

    // Get latest certificate (optionally filtered by scope)
    const cert = await getLatestCertificate(
      user.email,
      scope || undefined,
      scopeIdentifier || undefined,
      parentScopeIdentifier || undefined,
    )

    if (!cert) {
      const scopeDesc = scope ? ` for ${scope} scope` : ''
      return NextResponse.json(
        { error: `No certificate found for this installation${scopeDesc}` },
        { status: 404 },
      )
    }

    const scopeLog = scope ? `, scope: ${scope}` : ''
    console.log(`Downloading certificate for user: ${user.email}${scopeLog}, format: ${format}`)

    // Return in requested format
    switch (format) {
      case 'pem': {
        // Return certificate and key as separate PEM files in a tar/zip would be ideal
        // For now, return as JSON with PEM content
        return NextResponse.json({
          certificate: cert.certificate,
          privateKey: cert.privateKey,
          hostnames: cert.hostnames,
          scope: cert.scope,
          scopeIdentifier: cert.scopeIdentifier,
          parentScopeIdentifier: cert.parentScopeIdentifier,
          validFrom: cert.validFrom,
          validUntil: cert.validUntil,
        })
      }

      case 'bundle': {
        // Return certificate and key as a single bundle
        const bundle = `${cert.certificate}\n${cert.privateKey}`
        return new NextResponse(bundle, {
          headers: {
            'Content-Type': 'text/plain',
            'Content-Disposition': `attachment; filename="${user.subdomain}-tls-bundle.pem"`,
            'Cache-Control': 'no-store, no-cache, must-revalidate, max-age=0',
            Pragma: 'no-cache',
            Expires: '0',
          },
        })
      }

      case 'json':
      default: {
        const response = NextResponse.json({
          success: true,
          certificate: cert.certificate,
          privateKey: cert.privateKey,
          hostnames: cert.hostnames,
          scope: cert.scope,
          scopeIdentifier: cert.scopeIdentifier,
          parentScopeIdentifier: cert.parentScopeIdentifier,
          validFrom: cert.validFrom,
          validUntil: cert.validUntil,
          cloudflareCertId: cert.cloudflareCertId,
        })

        // Disable all caching
        response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
        response.headers.set('Pragma', 'no-cache')
        response.headers.set('Expires', '0')

        return response
      }
    }
  } catch (error) {
    console.error('Download certificates error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
