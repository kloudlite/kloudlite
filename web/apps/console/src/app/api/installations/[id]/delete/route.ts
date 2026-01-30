import { NextResponse } from 'next/server'
import { requireOwnerPermission } from '@/lib/console/authorization'
import {
  getInstallationById,
  deleteInstallation,
  deleteIpRecords,
  deleteDomainReservation,
} from '@/lib/console/storage'
import { deleteDnsRecords } from '@/lib/console/cloudflare-dns'

/**
 * Delete installation API route
 * Cleans up DNS records from Cloudflare before deleting from database
 */
export async function DELETE(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    // Only owner can delete installation
    await requireOwnerPermission(id)

    // Fetch the installation for DNS cleanup
    const installation = await getInstallationById(id)

    if (!installation) {
      return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
    }
    console.log(`Deleting installation: ${id}`)

    // Step 1: Get all DNS record IDs and delete IP records from database
    const dnsRecordIds = await deleteIpRecords(id)
    console.log(`Found ${dnsRecordIds.length} DNS records to delete`)

    // Step 2: Delete DNS records from Cloudflare
    if (dnsRecordIds.length > 0) {
      const success = await deleteDnsRecords(dnsRecordIds)
      if (success) {
        console.log(`Deleted ${dnsRecordIds.length} DNS records from Cloudflare`)
      } else {
        console.error('Some DNS records failed to delete from Cloudflare')
      }
    }

    // Step 3: Delete domain reservation
    try {
      await deleteDomainReservation(id)
      console.log(`Deleted domain reservation for: ${installation.subdomain}`)
    } catch (error) {
      console.error('Failed to delete domain reservation:', error)
      // Continue anyway
    }

    // Step 4: Delete the installation from database
    await deleteInstallation(id)
    console.log(`Installation deleted: ${id}`)

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('Error deleting installation:', error)
    const message = error instanceof Error ? error.message : 'Failed to delete installation'
    const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : 500
    return NextResponse.json({ error: message }, { status })
  }
}
