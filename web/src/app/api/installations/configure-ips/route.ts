import { NextRequest, NextResponse } from 'next/server'
import {
  getInstallationByKey,
  addOrUpdateIpRecord,
  markDeploymentReady,
  saveEdgeCertificate,
} from '@/lib/console/supabase-storage-service'
import type { IPRecord } from '@/lib/console/supabase-storage-service'
import {
  CLOUDFLARE_DNS_DOMAIN,
  createDomainRequestDnsRecords,
  updateDnsRecord,
  deleteDnsRecord,
} from '@/lib/console/cloudflare-dns'
import { createDomainRequestEdgeCertificates } from '@/lib/console/cloudflare-edge-certificates'

// Use Node.js runtime for Supabase (uses Node.js APIs)
export const runtime = 'nodejs'

/**
 * Configure IP for DomainRequest
 * Called by the DomainRequest controller after HAProxy is ready
 *
 * Request format:
 * {
 *   "installationKey": "abc-123",
 *   "ip": "1.2.3.4",
 *   "domainRequestName": "my-domain-request",
 *   "domainRoutes": [
 *     { "domain": "api.example.com" },
 *     { "domain": "app.example.com" }
 *   ]
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
    const { installationKey, ip, domainRequestName, domainRoutes } = body

    if (!installationKey) {
      return NextResponse.json({ error: 'Installation key is required' }, { status: 400 })
    }

    if (!ip || typeof ip !== 'string') {
      return NextResponse.json({ error: 'IP address is required' }, { status: 400 })
    }

    if (!domainRequestName || typeof domainRequestName !== 'string') {
      return NextResponse.json({ error: 'domainRequestName is required' }, { status: 400 })
    }

    // domainRoutes is optional (can be empty array or undefined)
    const routes = domainRoutes || []

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

    // Get existing IP record for this domain request
    const existingRecord = installation.ipRecords?.find(
      (r) => r.domainRequestName === domainRequestName,
    )

    let sshRecordId: string | null = null
    let routeRecordIds: string[] = []
    let edgeCertificateIds: string[] = []
    let dnsCreated = false

    const sshDomain = `ssh.${domainRequestName}.${installation.subdomain}.${CLOUDFLARE_DNS_DOMAIN}`

    // Create or update DNS records
    try {
      if (existingRecord) {
        // Check if IP changed or routes changed
        const ipChanged = existingRecord.ip !== ip
        const routesChanged = JSON.stringify(existingRecord.domainRoutes || []) !== JSON.stringify(routes)

        if (ipChanged && existingRecord.sshRecordId) {
          // Update SSH A record with new IP
          console.log(`Updating SSH A record for ${sshDomain} with new IP: ${ip}`)
          await updateDnsRecord(existingRecord.sshRecordId, sshDomain, ip, false)
          sshRecordId = existingRecord.sshRecordId
        } else {
          sshRecordId = existingRecord.sshRecordId || null
        }

        if (routesChanged) {
          // Delete old route CNAME records
          if (existingRecord.routeRecordIds && existingRecord.routeRecordIds.length > 0) {
            console.log(`Deleting old route CNAME records`)
            for (const recordId of existingRecord.routeRecordIds) {
              await deleteDnsRecord(recordId)
            }
          }

          // Create new route CNAME records
          if (routes.length > 0) {
            console.log(`Creating ${routes.length} new route CNAME records`)
            const result = await createDomainRequestDnsRecords(
              domainRequestName,
              installation.subdomain,
              ip,
              routes,
            )

            // Use existing SSH record or the new one
            sshRecordId = sshRecordId || result.sshRecordId
            routeRecordIds = result.routeRecordIds

            // Create edge certificates for new routes
            if (routeRecordIds.length > 0) {
              console.log(`Creating edge certificates for ${routes.length} route domains`)
              edgeCertificateIds = await createDomainRequestEdgeCertificates(routes)

              // Store edge certificates
              for (let i = 0; i < edgeCertificateIds.length; i++) {
                const certId = edgeCertificateIds[i]
                const domain = routes[i].domain
                await saveEdgeCertificate({
                  installationId: installation.id,
                  cloudflareCertPackId: certId,
                  hostnames: [domain],
                  domainRequestName,
                  status: 'pending',
                })
              }
            }
          }
        } else {
          // Routes didn't change, keep existing records
          routeRecordIds = existingRecord.routeRecordIds || []
        }

        dnsCreated = sshRecordId !== null
      } else {
        // Create new DNS records for new domain request
        console.log(`Creating DNS records for new domain request: ${domainRequestName}`)

        const result = await createDomainRequestDnsRecords(
          domainRequestName,
          installation.subdomain,
          ip,
          routes,
        )

        sshRecordId = result.sshRecordId
        routeRecordIds = result.routeRecordIds
        dnsCreated = sshRecordId !== null

        console.log(`Created SSH A record and ${routeRecordIds.length} route CNAME records`)

        // Create edge certificates for route domains
        if (dnsCreated && routes.length > 0) {
          console.log(`Creating edge certificates for ${routes.length} route domains`)
          edgeCertificateIds = await createDomainRequestEdgeCertificates(routes)

          // Store edge certificates
          for (let i = 0; i < edgeCertificateIds.length; i++) {
            const certId = edgeCertificateIds[i]
            const domain = routes[i].domain
            await saveEdgeCertificate({
              installationId: installation.id,
              cloudflareCertPackId: certId,
              hostnames: [domain],
              domainRequestName,
              status: 'pending',
            })
          }
        }
      }
    } catch (dnsError) {
      console.error('DNS record creation/update failed:', dnsError)
      // Continue even if DNS fails - use existing DNS record IDs if available
      if (existingRecord) {
        sshRecordId = existingRecord.sshRecordId || null
        routeRecordIds = existingRecord.routeRecordIds || []
      }
    }

    // Create new IP record with DNS record IDs and domain routes
    const newRecord: IPRecord = {
      domainRequestName,
      ip,
      configuredAt: new Date().toISOString(),
      sshRecordId,
      routeRecordIds,
      domainRoutes: routes,
    }

    // Atomically add or update IP record
    const totalRecords = await addOrUpdateIpRecord(installation.id, newRecord)

    // Atomically mark deployment as ready
    await markDeploymentReady(installation.id, true)

    const response = NextResponse.json({
      success: true,
      domainRequestName,
      ip,
      sshDomain,
      totalRecords,
      subdomain: `${installation.subdomain}.${CLOUDFLARE_DNS_DOMAIN}`,
      sshRecordCreated: sshRecordId !== null,
      routeRecordsCreated: routeRecordIds.length,
      edgeCertificatesCreated: edgeCertificateIds.length,
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
