import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getInstallationById,
  deleteInstallation,
  deleteIpRecords,
  deleteDomainReservation,
} from '@/lib/console/supabase-storage-service'
import { deleteDnsRecords } from '@/lib/console/cloudflare-dns'

/**
 * Delete installation API route
 * Cleans up DNS records from Cloudflare before deleting from database
 */
export async function DELETE(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  // Fetch the installation
  const installation = await getInstallationById(id)

  if (!installation) {
    return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
  }

  // Verify user owns this installation
  if (installation.userId !== session.user.id) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
  }

  try {
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
    return NextResponse.json({ error: 'Failed to delete installation' }, { status: 500 })
  }
}
