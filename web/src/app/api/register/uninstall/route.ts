import { NextRequest, NextResponse } from 'next/server'
import { getUserByInstallationKey, deleteIpRecords, deleteDomainReservation, resetUserInstallation, deleteCertificates } from '@/lib/registration/supabase-storage-service'
import { deleteDnsRecords } from '@/lib/registration/cloudflare-dns'
import { revokeCertificate } from '@/lib/registration/cloudflare-certificates'

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
      return NextResponse.json(
        { error: 'Missing or invalid authorization header' },
        { status: 401 }
      )
    }

    const secretKey = authHeader.substring(7) // Remove "Bearer " prefix

    const body = await request.json()
    const { installationKey } = body

    if (!installationKey) {
      return NextResponse.json(
        { error: 'Installation key is required' },
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

    console.log(`Starting uninstall for user: ${user.email}`)

    // Step 1: Get all DNS record IDs and delete IP records
    const dnsRecordIds = await deleteIpRecords(user.email)
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

    // Step 3: Delete TLS certificates
    let certRevokeCount = 0
    try {
      const certIds = await deleteCertificates(user.email)
      console.log(`Found ${certIds.length} certificates to revoke`)

      for (const certId of certIds) {
        const revoked = await revokeCertificate(certId)
        if (revoked) {
          certRevokeCount++
        }
      }
      console.log(`Revoked ${certRevokeCount} certificates from Cloudflare`)
    } catch (error) {
      console.error('Failed to delete/revoke certificates:', error)
      // Continue anyway - other cleanup is done
    }

    // Step 4: Delete domain reservation
    try {
      await deleteDomainReservation(user.email)
      console.log(`Deleted domain reservation for: ${user.subdomain}`)
    } catch (error) {
      console.error('Failed to delete domain reservation:', error)
      // Continue anyway - IP records and DNS are already deleted
    }

    // Step 5: Reset user installation
    await resetUserInstallation(user.email)
    console.log(`Reset installation for user: ${user.email}`)

    const response = NextResponse.json({
      success: true,
      message: 'Installation uninstalled successfully',
      email: user.email,
      subdomain: user.subdomain,
      dnsRecordsDeleted: dnsDeleteCount,
      ipRecordsDeleted: dnsRecordIds.length,
      certificatesRevoked: certRevokeCount,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Uninstall error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
