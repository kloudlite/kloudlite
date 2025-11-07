import { NextRequest, NextResponse } from 'next/server'
import {
  getInstallationByKey,
  addOrUpdateIpRecord,
  removeIpRecord,
  markDeploymentReady,
  deleteEdgeCertificatesForDomainRequest,
  deleteEdgeCertificateForDomain,
  type IPRecord,
  type Installation,
} from '@/lib/console/supabase-storage-service'
import {
  CLOUDFLARE_DNS_DOMAIN,
  createDnsRecord,
  createCnameRecord,
  updateDnsRecord,
  deleteDnsRecord,
} from '@/lib/console/cloudflare-dns'
import {
  createOrReuseWildcardEdgeCertificate,
  deleteEdgeCertificate,
} from '@/lib/console/cloudflare-edge-certificates'

// Use Node.js runtime for Supabase
export const runtime = 'nodejs'

/**
 * Configure IP and DNS records for DomainRequest
 *
 * Request format:
 * {
 *   "installationKey": "abc-123",
 *   "ip": "1.2.3.4",
 *   "domainRequestName": "my-domain-request",
 *   "domains": ["api.example.com", "app.example.com"],
 *   "deleted": false  // Set to true to delete all records
 * }
 */

// Calculate difference between old and new domain lists
function calculateDomainDiff(oldDomains: string[], newDomains: string[]) {
  const oldSet = new Set(oldDomains)
  const newSet = new Set(newDomains)

  const toAdd = newDomains.filter(d => !oldSet.has(d))
  const toRemove = oldDomains.filter(d => !newSet.has(d))
  const unchanged = newDomains.filter(d => oldSet.has(d))

  return { toAdd, toRemove, unchanged }
}

// Handle deletion of domain request DNS records
async function handleDeletion(
  installation: Installation,
  domainRequestName: string,
): Promise<NextResponse> {
  const existingRecord = installation.ipRecords?.find(
    r => r.domainRequestName === domainRequestName
  )

  if (!existingRecord) {
    return NextResponse.json({
      success: true,
      message: 'Already deleted'
    })
  }

  console.log(`Deleting all DNS records for domain request: ${domainRequestName}`)

  // Delete SSH A record
  if (existingRecord.sshRecordId) {
    await deleteDnsRecord(existingRecord.sshRecordId)
    console.log(`Deleted SSH A record: ${existingRecord.sshRecordId}`)
  }

  // Delete all CNAME records
  if (existingRecord.routeRecordIds) {
    for (const recordId of existingRecord.routeRecordIds) {
      await deleteDnsRecord(recordId)
      console.log(`Deleted CNAME record: ${recordId}`)
    }
  }

  // Delete edge certificates from database
  const certPackIds = await deleteEdgeCertificatesForDomainRequest(
    installation.id,
    domainRequestName
  )

  // Delete edge certificates from Cloudflare
  for (const certPackId of certPackIds) {
    try {
      await deleteEdgeCertificate(certPackId)
      console.log(`Deleted edge certificate from Cloudflare: ${certPackId}`)
    } catch (err) {
      console.error(`Failed to delete edge certificate ${certPackId}:`, err)
    }
  }

  // Remove IP record from database
  await removeIpRecord(installation.id, domainRequestName)
  console.log(`Removed IP record for: ${domainRequestName}`)

  return NextResponse.json({
    success: true,
    deleted: true,
    domainRequestName,
  })
}

export async function POST(request: NextRequest) {
  try {
    // Validate authorization
    const authHeader = request.headers.get('authorization')
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return NextResponse.json(
        { error: 'Missing or invalid authorization header' },
        { status: 401 },
      )
    }

    const secretKey = authHeader.substring(7)
    const body = await request.json()
    const { installationKey, ip, domainRequestName, domains, deleted } = body

    // Validate required fields
    if (!installationKey) {
      return NextResponse.json({ error: 'installationKey is required' }, { status: 400 })
    }

    if (!domainRequestName) {
      return NextResponse.json({ error: 'domainRequestName is required' }, { status: 400 })
    }

    // Get installation
    const installation = await getInstallationByKey(installationKey)
    if (!installation) {
      return NextResponse.json({ error: 'Invalid installation key' }, { status: 404 })
    }

    // Verify secret key
    if (installation.secretKey !== secretKey) {
      return NextResponse.json({ error: 'Invalid secret key' }, { status: 403 })
    }

    // Handle deletion
    if (deleted) {
      return await handleDeletion(installation, domainRequestName)
    }

    // For creation/update, require IP and subdomain
    if (!ip) {
      return NextResponse.json({ error: 'IP address is required' }, { status: 400 })
    }

    if (!installation.subdomain) {
      return NextResponse.json(
        { error: 'Installation must have a subdomain assigned' },
        { status: 400 },
      )
    }

    const domainList: string[] = domains || []
    const sshDomain = `ssh.${domainRequestName}.${installation.subdomain}.${CLOUDFLARE_DNS_DOMAIN}`

    // Get existing record
    const existingRecord = installation.ipRecords?.find(
      r => r.domainRequestName === domainRequestName
    )

    let sshRecordId: string | null = null
    const routeRecordMap: Record<string, string> = {}
    let dnsSuccess = false

    try {
      // Create or update SSH A record
      if (existingRecord) {
        if (existingRecord.ip !== ip && existingRecord.sshRecordId) {
          // Update existing A record with new IP
          console.log(`Updating SSH A record ${existingRecord.sshRecordId}: ${sshDomain} → ${ip}`)
          await updateDnsRecord(existingRecord.sshRecordId, sshDomain, ip, false)
          sshRecordId = existingRecord.sshRecordId
        } else {
          sshRecordId = existingRecord.sshRecordId || null
        }
      } else {
        // Create new SSH A record
        console.log(`Creating SSH A record: ${sshDomain} → ${ip}`)
        sshRecordId = await createDnsRecord(sshDomain, ip, false)
      }

      dnsSuccess = sshRecordId !== null

      // Handle domain route changes (differential updates)
      if (existingRecord) {
        const oldDomains = (existingRecord.domainRoutes || []).map(r => r.domain)
        const { toAdd, toRemove } = calculateDomainDiff(oldDomains, domainList)

        console.log(`[DEBUG] Domain changes: +${toAdd.length} -${toRemove.length}`)
        console.log(`[DEBUG] Old domains:`, oldDomains)
        console.log(`[DEBUG] New domains:`, domainList)
        console.log(`[DEBUG] To add:`, toAdd)
        console.log(`[DEBUG] To remove:`, toRemove)

        // Copy existing valid record mappings
        console.log(`[DEBUG] Existing routeRecordMap:`, existingRecord.routeRecordMap)
        if (existingRecord.routeRecordMap) {
          Object.assign(routeRecordMap, existingRecord.routeRecordMap)
          console.log(`[DEBUG] Copied routeRecordMap:`, routeRecordMap)
        } else {
          console.log(`[DEBUG] No existing routeRecordMap found`)
        }

        // Delete removed domains
        for (const domain of toRemove) {
          const recordId = routeRecordMap[domain]
          console.log(`[DEBUG] Processing deletion for domain: ${domain}, recordId: ${recordId}`)
          if (recordId) {
            await deleteDnsRecord(recordId)
            console.log(`[DEBUG] Deleted CNAME DNS record for removed domain: ${domain} (ID: ${recordId})`)
            delete routeRecordMap[domain]
          } else {
            console.log(`[DEBUG] No DNS record ID found in routeRecordMap for domain: ${domain}`)
          }

          // Delete edge certificate
          const certPackId = await deleteEdgeCertificateForDomain(
            installation.id,
            domainRequestName,
            domain
          )
          console.log(`[DEBUG] Edge certificate lookup for ${domain}: ${certPackId}`)
          if (certPackId) {
            try {
              await deleteEdgeCertificate(certPackId)
              console.log(`[DEBUG] Deleted edge certificate from Cloudflare for ${domain}: ${certPackId}`)
            } catch (err) {
              console.error(`[DEBUG] Failed to delete edge certificate for ${domain}:`, err)
            }
          } else {
            console.log(`[DEBUG] No edge certificate found for domain: ${domain}`)
          }
        }

        // Create or reuse wildcard certificate for new domains
        if (toAdd.length > 0) {
          const wildcardCertId = await createOrReuseWildcardEdgeCertificate(
            installation.id,
            toAdd.map(domain => ({ domain })),
            domainRequestName
          )
          if (wildcardCertId) {
            console.log(`Using wildcard certificate ${wildcardCertId} for ${toAdd.length} new domains`)
          }
        }

        // Create new domains
        for (const domain of toAdd) {
          const cnameRecordId = await createCnameRecord(domain, sshDomain, true)
          if (cnameRecordId) {
            routeRecordMap[domain] = cnameRecordId
            console.log(`Created CNAME: ${domain} → ${sshDomain}`)
          }
        }
      } else {
        // New domain request - create all CNAMEs
        console.log(`Creating ${domainList.length} new CNAME records`)

        // Create or reuse wildcard certificate for all domains
        if (domainList.length > 0) {
          const wildcardCertId = await createOrReuseWildcardEdgeCertificate(
            installation.id,
            domainList.map(domain => ({ domain })),
            domainRequestName
          )
          if (wildcardCertId) {
            console.log(`Using wildcard certificate ${wildcardCertId} for ${domainList.length} domains`)
          }
        }

        for (const domain of domainList) {
          const cnameRecordId = await createCnameRecord(domain, sshDomain, true)
          if (cnameRecordId) {
            routeRecordMap[domain] = cnameRecordId
            console.log(`Created CNAME: ${domain} → ${sshDomain}`)
          }
        }
      }
    } catch (dnsError) {
      console.error('DNS record operation failed:', dnsError)
      // Continue with partial state if some records were created
    }

    // Store updated state
    const newRecord: IPRecord = {
      domainRequestName,
      ip,
      configuredAt: new Date().toISOString(),
      sshRecordId,
      routeRecordIds: Object.values(routeRecordMap),
      routeRecordMap,
      domainRoutes: domainList.map(domain => ({ domain })),
    }

    const totalRecords = await addOrUpdateIpRecord(installation.id, newRecord)
    await markDeploymentReady(installation.id, true)

    const response = NextResponse.json({
      success: true,
      domainRequestName,
      ip,
      sshDomain,
      subdomain: `${installation.subdomain}.${CLOUDFLARE_DNS_DOMAIN}`,
      sshRecordCreated: sshRecordId !== null,
      routeRecordsCreated: Object.keys(routeRecordMap).length,
      totalRecords,
      dnsSuccess,
    })

    // Disable caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (error) {
    console.error('Configure IP error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
