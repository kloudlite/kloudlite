import { NextRequest, NextResponse } from 'next/server'
import { getInstallationByKey } from '@/lib/console/supabase-storage-service'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'

/**
 * Get origin certificate for an installation
 * Called by the DomainRequest controller to download the installation's origin certificate
 *
 * Request format (Query parameters):
 * {
 *   "installationKey": "abc-123"
 * }
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

    // Get installationKey from query parameters
    const { searchParams } = new URL(request.url)
    const installationKey = searchParams.get('installationKey')

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

    // Check if origin certificate exists
    if (!installation.originCertificate || !installation.originPrivateKey) {
      return NextResponse.json(
        { error: 'Origin certificate not found for this installation' },
        { status: 404 },
      )
    }

    console.log(
      `Serving origin certificate for installation: ${installation.id}, cert ID: ${installation.originCertId}`,
    )

    const response = NextResponse.json({
      success: true,
      certificate: installation.originCertificate,
      privateKey: installation.originPrivateKey,
      certificateId: installation.originCertId,
      validFrom: installation.originCertValidFrom,
      validUntil: installation.originCertValidUntil,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Get origin certificate error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
