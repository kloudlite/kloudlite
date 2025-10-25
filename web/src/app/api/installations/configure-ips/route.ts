import { NextRequest, NextResponse } from 'next/server'
import {
  getInstallationByKey,
  addOrUpdateIpRecord,
  markDeploymentReady,
} from '@/lib/registration/supabase-storage-service'
import type { IPRecord } from '@/lib/registration/supabase-storage-service'
import {
  createInstallationDnsRecords,
  createWorkmachineDnsRecords,
  updateDnsRecords,
} from '@/lib/registration/cloudflare-dns'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'
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
        { status: 401 },
      )
    }

    const secretKey = authHeader.substring(7) // Remove "Bearer " prefix

    const body = await request.json()
    const { installationKey, type, ip, workMachineName } = body

    if (!installationKey) {
      return NextResponse.json({ error: 'Installation key is required' }, { status: 400 })
    }

    if (!type || (type !== 'installation' && type !== 'workmachine')) {
      return NextResponse.json(
        { error: 'Type must be "installation" or "workmachine"' },
        { status: 400 },
      )
    }

    if (!ip || typeof ip !== 'string') {
      return NextResponse.json({ error: 'IP address is required' }, { status: 400 })
    }

    if (type === 'workmachine' && !workMachineName) {
      return NextResponse.json(
        { error: 'workMachineName is required for workmachine type' },
        { status: 400 },
      )
    }

    // Look up installation by key
    const installation = await getInstallationByKey(installationKey)

    if (!installation) {
      return NextResponse.json({ error: 'Invalid installation key' }, { status: 404 })
    }

    // Verify secret key matches
    if (installation.secretKey !== secretKey) {
      return NextResponse.json({ error: 'Invalid secret key' }, { status: 403 })
    }

    // Check if installation has subdomain assigned (required for DNS creation)
    if (!installation.subdomain) {
      return NextResponse.json(
        { error: 'Installation must have a subdomain assigned before configuring IPs' },
        { status: 400 },
      )
    }

    // Get existing IP record to check if we need to update DNS
    const existingRecord = installation.ipRecords?.find((r) => {
      if (r.type === 'installation' && type === 'installation') {
        return true
      }
      if (
        r.type === 'workmachine' &&
        type === 'workmachine' &&
        r.workMachineName === workMachineName
      ) {
        return true
      }
      return false
    })

    let dnsRecordIds: string[] = []
    let dnsCreated = false

    // Create or update DNS records
    try {
      if (existingRecord) {
        // Update existing record
        // If IP changed and we have DNS record IDs, update them
        if (
          existingRecord.ip !== ip &&
          existingRecord.dnsRecordIds &&
          existingRecord.dnsRecordIds.length > 0
        ) {
          const domainName =
            type === 'installation'
              ? `${installation.subdomain}`
              : `${workMachineName}.${installation.subdomain}`

          await updateDnsRecords(existingRecord.dnsRecordIds, domainName, ip)
          dnsRecordIds = existingRecord.dnsRecordIds
          dnsCreated = true
          console.log(`Updated DNS records for ${type}: ${domainName}`)
        } else if (!existingRecord.dnsRecordIds || existingRecord.dnsRecordIds.length === 0) {
          // No DNS records exist, create them
          if (type === 'installation') {
            dnsRecordIds = await createInstallationDnsRecords(installation.subdomain, ip)
            dnsCreated = dnsRecordIds.length > 0
            console.log(`Created ${dnsRecordIds.length} DNS records for installation`)
          } else if (type === 'workmachine') {
            dnsRecordIds = await createWorkmachineDnsRecords(
              workMachineName!,
              installation.subdomain,
              ip,
            )
            dnsCreated = dnsRecordIds.length > 0
            console.log(
              `Created ${dnsRecordIds.length} DNS records for workmachine: ${workMachineName}`,
            )
          }
        } else {
          // IP didn't change, keep existing DNS record IDs
          dnsRecordIds = existingRecord.dnsRecordIds
          dnsCreated = true
        }
      } else {
        // Create new DNS records for new IP record
        if (type === 'installation') {
          dnsRecordIds = await createInstallationDnsRecords(installation.subdomain, ip)
          dnsCreated = dnsRecordIds.length > 0
          console.log(`Created ${dnsRecordIds.length} DNS records for new installation`)
        } else if (type === 'workmachine') {
          dnsRecordIds = await createWorkmachineDnsRecords(
            workMachineName!,
            installation.subdomain,
            ip,
          )
          dnsCreated = dnsRecordIds.length > 0
          console.log(
            `Created ${dnsRecordIds.length} DNS records for new workmachine: ${workMachineName}`,
          )
        }
      }
    } catch (dnsError) {
      console.error('DNS record creation/update failed:', dnsError)
      // Continue even if DNS fails - use existing DNS record IDs if available
      if (existingRecord?.dnsRecordIds) {
        dnsRecordIds = existingRecord.dnsRecordIds
      }
    }

    // Create new IP record with DNS record IDs
    const newRecord: IPRecord = {
      type,
      ip,
      configuredAt: new Date().toISOString(),
      dnsRecordIds,
    }

    if (type === 'workmachine') {
      newRecord.workMachineName = workMachineName
    }

    // Atomically add or update IP record
    const totalRecords = await addOrUpdateIpRecord(installation.id, newRecord)

    // Atomically mark deployment as ready
    await markDeploymentReady(installation.id, true)

    const response = NextResponse.json({
      success: true,
      type,
      ip,
      workMachineName: type === 'workmachine' ? workMachineName : undefined,
      totalRecords,
      subdomain: installation.subdomain,
      dnsRecordsCreated: dnsRecordIds.length,
      dnsSuccess: dnsCreated,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Configure IP error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
