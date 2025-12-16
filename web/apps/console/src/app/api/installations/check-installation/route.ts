import { NextRequest, NextResponse } from 'next/server'
import { getInstallationByKey } from '@/lib/console/supabase-storage-service'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'

/**
 * Check if installation has been verified by deployment
 * Used by UI to verify installation is complete before allowing domain selection
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { installationKey } = body

    if (!installationKey) {
      return NextResponse.json({ error: 'Installation key is required' }, { status: 400 })
    }

    // Look up installation by key
    const installation = await getInstallationByKey(installationKey)

    if (!installation) {
      return NextResponse.json({ error: 'Invalid installation key' }, { status: 404 })
    }

    // Check if installation has been verified (has secret key)
    const verified = !!installation.secretKey

    // Check if DNS is configured:
    // 1. deploymentReady flag is set, OR
    // 2. has IP records with DNS record IDs
    const dnsConfigured = installation.deploymentReady || installation.ipRecords?.some(
      (record) => record.sshRecordId || (record.routeRecordIds && record.routeRecordIds.length > 0)
    ) ?? false

    const response = NextResponse.json({
      verified,
      dnsConfigured,
      hasCompletedInstallation: installation.hasCompletedInstallation,
      message: !verified
        ? 'Waiting for deployment to contact the server...'
        : !dnsConfigured
          ? 'Waiting for DNS configuration...'
          : 'Installation complete',
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Check installation error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
