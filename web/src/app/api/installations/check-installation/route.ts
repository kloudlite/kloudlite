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
    const pollerActive = !!installation.pollerActive

    const response = NextResponse.json({
      verified,
      pollerActive,
      hasCompletedInstallation: installation.hasCompletedInstallation,
      message: pollerActive
        ? 'Deployment is active and polling for configuration'
        : verified
          ? 'Installation verified, waiting for deployment to start polling'
          : 'Installation not verified. Please ensure the deployment has contacted the server.',
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
