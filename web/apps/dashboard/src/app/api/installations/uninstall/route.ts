import { NextRequest, NextResponse } from 'next/server'
import {
  getInstallationByKey,
  deleteIpRecords,
  deleteDomainReservation,
  resetInstallation,
  deleteCertificates,
  deleteEdgeCertificates,
} from '@/lib/console/supabase-storage-service'
import { deleteDnsRecords } from '@/lib/console/cloudflare-dns'
import { revokeCertificate } from '@/lib/console/cloudflare-certificates'
import { deleteEdgeCertificate } from '@/lib/console/cloudflare-edge-certificates'

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
      return NextResponse.json(
        { error: 'Missing or invalid authorization header' },
        { status: 401 },
      )
    }

    const secretKey = authHeader.substring(7) // Remove "Bearer " prefix

    const body = await request.json()
    const { installationKey } = body

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

    console.log(`Starting uninstall for installation: ${installation.id}`)

    // Step 1: Get all DNS record IDs and delete IP records
    const dnsRecordIds = await deleteIpRecords(installation.id)
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

    // Step 3: Delete and revoke all TLS certificates from tls_certificates table
    // This includes origin certificates (scope='installation'), workmachine certificates, and workspace certificates
    let certRevokeCount = 0
    try {
      const certIds = await deleteCertificates(installation.id)
      console.log(`Found ${certIds.length} certificates to revoke from tls_certificates table`)

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

    // Step 3b: Delete edge certificates
    let edgeCertDeleteCount = 0
    let edgeCertFailCount = 0
    try {
      const edgeCertPackIds = await deleteEdgeCertificates(installation.id)
      console.log(`Found ${edgeCertPackIds.length} edge certificates to delete`)

      for (const certPackId of edgeCertPackIds) {
        console.log(`Attempting to delete edge certificate: ${certPackId}`)
        const deleted = await deleteEdgeCertificate(certPackId)
        if (deleted) {
          edgeCertDeleteCount++
          console.log(`Successfully deleted edge certificate: ${certPackId}`)
        } else {
          edgeCertFailCount++
          console.error(`Failed to delete edge certificate: ${certPackId}`)
        }
      }
      console.log(`Deleted ${edgeCertDeleteCount} edge certificates from Cloudflare (${edgeCertFailCount} failed)`)
    } catch (error) {
      console.error('Failed to delete edge certificates:', error)
      // Continue anyway
    }

    // Step 4: Delete domain reservation
    try {
      await deleteDomainReservation(installation.id)
      console.log(`Deleted domain reservation for: ${installation.subdomain}`)
    } catch (error) {
      console.error('Failed to delete domain reservation:', error)
      // Continue anyway - IP records and DNS are already deleted
    }

    // Step 5: Reset installation
    await resetInstallation(installation.id)
    console.log(`Reset installation: ${installation.id}`)

    const response = NextResponse.json({
      success: true,
      message: 'Installation uninstalled successfully',
      installationId: installation.id,
      subdomain: installation.subdomain,
      dnsRecordsDeleted: dnsDeleteCount,
      ipRecordsDeleted: dnsRecordIds.length,
      certificatesRevoked: certRevokeCount,
      edgeCertificatesDeleted: edgeCertDeleteCount,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Uninstall error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
