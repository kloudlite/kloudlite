import { NextRequest, NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import {
  getInstallationByKey,
  deleteDnsConfigurations,
  deleteDomainReservation,
  resetInstallation,
} from '@/lib/console/storage'
import { deleteDnsRecords } from '@/lib/console/cloudflare-dns'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'
/**
 * Uninstall deployment
 * Called by the deployment or user to completely remove installation and cleanup DNS
 *
 * Request format:
 * {
 *   "installationKey": "abc-123"
 * }
 *
 * Requires Authorization: Bearer <secret-key>
 */
export async function POST(request: NextRequest) {
  try {
    // Extract and validate bearer token
    const authHeader = request.headers.get('authorization')
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return apiError('Missing or invalid authorization header', 401)
    }

    const secretKey = authHeader.substring(7) // Remove "Bearer " prefix

    const body = await request.json()
    const { installationKey } = body

    if (!installationKey) {
      return apiError('Installation key is required', 400)
    }

    // Look up installation by installation key
    const installation = await getInstallationByKey(installationKey)

    if (!installation) {
      return apiError('Invalid installation key', 404)
    }

    // Verify secret key matches
    if (installation.secretKey !== secretKey) {
      return apiError('Invalid secret key', 403)
    }

    console.log(`Starting uninstall for installation: ${installation.id}`)

    // Step 1: Get all DNS record IDs and delete IP records
    const dnsRecordIds = await deleteDnsConfigurations(installation.id)
    console.log(`Found ${dnsRecordIds.length} DNS records to delete`)

    // Step 2: Delete DNS records from Cloudflare
    let dnsDeleteCount = 0
    if (dnsRecordIds.length > 0) {
      const dnsDeleteSuccess = await deleteDnsRecords(dnsRecordIds)
      if (dnsDeleteSuccess) {
        dnsDeleteCount = dnsRecordIds.length
        console.log(`Successfully deleted ${dnsDeleteCount} DNS records from Cloudflare`)
      } else {
        console.error('Some DNS records failed to delete from Cloudflare')
      }
    }

    // Step 3: Delete domain reservation
    try {
      await deleteDomainReservation(installation.id)
      console.log(`Deleted domain reservation for: ${installation.subdomain}`)
    } catch (error) {
      console.error('Failed to delete domain reservation:', error)
      // Continue anyway - IP records and DNS are already deleted
    }

    // Step 4: Reset installation
    await resetInstallation(installation.id)
    console.log(`Reset installation: ${installation.id}`)

    const response = NextResponse.json({
      success: true,
      message: 'Installation uninstalled successfully',
      installationId: installation.id,
      subdomain: installation.subdomain,
      dnsRecordsDeleted: dnsDeleteCount,
      dnsConfigurationsDeleted: dnsRecordIds.length,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Uninstall error:', error)
    return apiError('Internal server error', 500)
  }
}
