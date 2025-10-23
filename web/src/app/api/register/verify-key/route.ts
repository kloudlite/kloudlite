import { NextRequest, NextResponse } from 'next/server'
import {
  getUserByInstallationKey,
  markInstallationComplete,
  updateHealthCheck,
} from '@/lib/registration/supabase-storage-service'

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

    // Look up user by installation key
    const user = await getUserByInstallationKey(installationKey)

    if (!user) {
      return NextResponse.json({ error: 'Invalid installation key' }, { status: 404 })
    }

    // Generate secret key on first verification (if not exists)
    if (!user.secretKey) {
      console.log('First deployment verification for:', user.email)
      console.log('Generating secret key for installation key:', installationKey)

      const secretKey = crypto.randomUUID()

      // Atomically mark installation complete and set secret key
      const updatedUser = await markInstallationComplete(user.email, secretKey)
      user.secretKey = updatedUser.secretKey
      user.hasCompletedInstallation = updatedUser.hasCompletedInstallation

      console.log('Secret key generated and installation marked as complete')
    }

    // Atomically update last health check timestamp (deployment is polling)
    const updatedUser = await updateHealthCheck(user.email)
    user.lastHealthCheck = updatedUser.lastHealthCheck

    // Return only operational information needed by deployment
    const response = NextResponse.json({
      success: true,
      secretKey: user.secretKey,
      subdomain: user.subdomain,
      deploymentReady: user.deploymentReady || false,
      ipRecords: user.ipRecords || [],
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
