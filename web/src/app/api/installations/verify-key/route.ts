import { NextRequest, NextResponse } from 'next/server'
import {
  getInstallationByKey,
  markInstallationComplete,
  updateHealthCheck,
  updateInstallation,
} from '@/lib/console/supabase-storage-service'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'
/**
 * Verify installation key (POST method)
 * Used by installation script to verify the key and get user info
 * Also used by deployment to poll for configuration (every 10 minutes)
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

    // Generate secret key on first verification (if not exists)
    let updatedInstallation = installation
    if (!installation.secretKey) {
      console.log('First deployment verification for installation:', installation.id)
      console.log('Generating secret key for installation key:', installationKey)

      const secretKey = crypto.randomUUID()

      // Atomically mark installation complete and set secret key
      updatedInstallation = await markInstallationComplete(installation.id, secretKey)

      console.log('Secret key generated and installation marked as complete')
    } else if (!installation.pollerActive) {
      // Secret key exists but pollerActive is false
      // This means SubdomainPoller has started polling (2nd+ call)
      console.log('SubdomainPoller started polling for installation:', installation.id)

      // Mark poller as active
      updatedInstallation = await updateInstallation(installation.id, { pollerActive: true })

      console.log('Poller marked as active')
    }

    // Atomically update last health check timestamp (deployment is polling)
    updatedInstallation = await updateHealthCheck(installation.id)

    // Return only operational information needed by deployment
    const response = NextResponse.json({
      success: true,
      secretKey: updatedInstallation.secretKey,
      subdomain: updatedInstallation.subdomain,
      deploymentReady: updatedInstallation.deploymentReady || false,
      ipRecords: updatedInstallation.ipRecords || [],
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Verification error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
