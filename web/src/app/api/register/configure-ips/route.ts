import { NextRequest, NextResponse } from 'next/server'
import { getUserByInstallationKey, saveUserRegistration } from '@/lib/registration/storage-service'
import type { IPRecord } from '@/lib/registration/storage-service'

/**
 * Configure IP for deployment
 * Called by the installed deployment to send individual IP configurations
 *
 * Request format:
 * {
 *   "installationKey": "abc-123",
 *   "type": "installation" | "workmachine",
 *   "ip": "1.2.3.4",
 *   "workMachineName": "user1" (optional, for workmachine type only)
 * }
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
    const { installationKey, type, ip, workMachineName } = body

    if (!installationKey) {
      return NextResponse.json(
        { error: 'Installation key is required' },
        { status: 400 }
      )
    }

    if (!type || (type !== 'installation' && type !== 'workmachine')) {
      return NextResponse.json(
        { error: 'Type must be "installation" or "workmachine"' },
        { status: 400 }
      )
    }

    if (!ip || typeof ip !== 'string') {
      return NextResponse.json(
        { error: 'IP address is required' },
        { status: 400 }
      )
    }

    if (type === 'workmachine' && !workMachineName) {
      return NextResponse.json(
        { error: 'workMachineName is required for workmachine type' },
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

    // Initialize ipRecords array if it doesn't exist
    if (!user.ipRecords) {
      user.ipRecords = []
    }

    // Create new IP record
    const newRecord: IPRecord = {
      type,
      ip,
      configuredAt: new Date().toISOString(),
    }

    if (type === 'workmachine') {
      newRecord.workMachineName = workMachineName
    }

    // Check if this type/workMachineName combination already exists, update if so, otherwise add
    const existingIndex = user.ipRecords.findIndex(r => {
      if (r.type === 'installation' && type === 'installation') {
        return true
      }
      if (r.type === 'workmachine' && type === 'workmachine' && r.workMachineName === workMachineName) {
        return true
      }
      return false
    })

    if (existingIndex >= 0) {
      user.ipRecords[existingIndex] = newRecord
    } else {
      user.ipRecords.push(newRecord)
    }

    // Mark deployment as ready
    // Note: hasCompletedInstallation is set when secret key is generated (in verify-key API)
    user.deploymentReady = true

    // Save updated registration
    await saveUserRegistration(user)

    const response = NextResponse.json({
      success: true,
      type,
      ip,
      workMachineName: type === 'workmachine' ? workMachineName : undefined,
      totalRecords: user.ipRecords.length,
      subdomain: user.subdomain,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Configure IP error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
