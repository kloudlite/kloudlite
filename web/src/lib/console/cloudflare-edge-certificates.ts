/**
 * CloudFlare Edge Certificate Management
 *
 * Manages SSL/TLS certificates for browser-to-CloudFlare connections
 * Uses CloudFlare Advanced Certificate Manager for wildcard subdomains
 */

const CLOUDFLARE_API_TOKEN = process.env.CLOUDFLARE_API_TOKEN!
const CLOUDFLARE_ZONE_ID = process.env.CLOUDFLARE_ZONE_ID!

const CERT_API_BASE = `https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/ssl/certificate_packs`

interface CloudflareCertificatePackResponse {
  success: boolean
  errors: Array<{ code: number; message: string }>
  messages: string[]
  result: {
    id: string
    type: string
    hosts: string[]
    status: string
    validation_method: string
    validity_days: number
    certificate_authority: string
    cloudflare_branding: boolean
  }
}

interface EdgeCertificateInfo {
  id: string
  hosts: string[]
  status: string
  validityDays: number
  certificateAuthority: string
}

/**
 * Order a new Advanced Certificate for wildcard subdomains
 * @param hosts - Array of hostnames to cover (e.g., ["subdomain.khost.dev", "*.subdomain.khost.dev"])
 * @returns Certificate pack ID or null if failed
 */
export async function orderEdgeCertificate(hosts: string[]): Promise<string | null> {
  try {
    console.log(`Ordering CloudFlare Edge Certificate for hosts:`, hosts)

    const response = await fetch(`${CERT_API_BASE}/order`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        type: 'advanced',
        hosts,
        validation_method: 'http',
        validity_days: 90,
        certificate_authority: 'lets_encrypt',
        cloudflare_branding: false,
      }),
    })

    if (!response.ok) {
      const error = await response.text()
      console.error(`Edge certificate order failed:`, error)
      return null
    }

    const result: CloudflareCertificatePackResponse = await response.json()

    if (!result.success) {
      console.error(`Edge certificate order failed:`, result.errors)
      return null
    }

    console.log(`Edge certificate ordered successfully: ${result.result.id}`)
    console.log(`Status: ${result.result.status}`)
    console.log(`Hosts: ${result.result.hosts.join(', ')}`)

    return result.result.id
  } catch (error) {
    console.error(`Edge certificate order error:`, error)
    return null
  }
}

/**
 * Get certificate pack information
 * @param certificateId - Certificate pack ID
 * @returns Certificate information or null if not found
 */
export async function getEdgeCertificate(certificateId: string): Promise<EdgeCertificateInfo | null> {
  try {
    const response = await fetch(`${CERT_API_BASE}/${certificateId}`, {
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
      },
    })

    if (!response.ok) {
      return null
    }

    const result: CloudflareCertificatePackResponse = await response.json()

    if (!result.success) {
      return null
    }

    return {
      id: result.result.id,
      hosts: result.result.hosts,
      status: result.result.status,
      validityDays: result.result.validity_days,
      certificateAuthority: result.result.certificate_authority,
    }
  } catch (error) {
    console.error(`Get edge certificate error:`, error)
    return null
  }
}

/**
 * Delete a certificate pack
 * @param certificateId - Certificate pack ID
 * @returns Success status
 */
export async function deleteEdgeCertificate(certificateId: string): Promise<boolean> {
  try {
    console.log(`Deleting edge certificate: ${certificateId}`)

    const response = await fetch(`${CERT_API_BASE}/${certificateId}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
      },
    })

    if (!response.ok && response.status !== 404) {
      const error = await response.text()
      console.error(`Edge certificate deletion failed:`, error)
      return false
    }

    console.log(`Edge certificate deleted successfully: ${certificateId}`)
    return true
  } catch (error) {
    console.error(`Edge certificate deletion error:`, error)
    return false
  }
}

/**
 * Create edge certificate for installation subdomain
 * @param subdomain - User's subdomain (e.g., "karthik")
 * @param baseDomain - Base domain (e.g., "khost.dev")
 * @returns Certificate pack ID or null if failed
 */
export async function createInstallationEdgeCertificate(
  subdomain: string,
  baseDomain: string,
): Promise<string | null> {
  const hosts = [
    `${subdomain}.${baseDomain}`,
    `*.${subdomain}.${baseDomain}`,
  ]

  return orderEdgeCertificate(hosts)
}

/**
 * Create edge certificate for workmachine subdomain
 * @param workMachineName - Workmachine name (e.g., "node1")
 * @param subdomain - User's subdomain (e.g., "karthik")
 * @param baseDomain - Base domain (e.g., "khost.dev")
 * @returns Certificate pack ID or null if failed
 */
export async function createWorkmachineEdgeCertificate(
  workMachineName: string,
  subdomain: string,
  baseDomain: string,
): Promise<string | null> {
  const hosts = [
    `${workMachineName}.${subdomain}.${baseDomain}`,
    `*.${workMachineName}.${subdomain}.${baseDomain}`,
  ]

  return orderEdgeCertificate(hosts)
}

/**
 * Extract subdomain pattern from a domain
 * For example: "x.karthik.khost.dev" -> "karthik.khost.dev"
 */
function extractSubdomainPattern(domain: string): string | null {
  const parts = domain.split('.')
  if (parts.length < 3) {
    return null // Need at least 3 parts: service.subdomain.domain
  }
  // Remove the first part (service name) to get the subdomain pattern
  return parts.slice(1).join('.')
}

/**
 * Create or reuse wildcard edge certificate for domain routes
 * Checks if a wildcard certificate already exists before creating a new one
 *
 * @param installationId - Installation ID
 * @param domainRoutes - Array of domain routes to create certificates for
 * @param domainRequestName - Domain request name for tracking
 * @returns Certificate pack ID (either existing or newly created)
 */
export async function createOrReuseWildcardEdgeCertificate(
  installationId: string,
  domainRoutes: Array<{ domain: string }>,
  domainRequestName: string,
): Promise<string | null> {
  if (domainRoutes.length === 0) {
    return null
  }

  // Extract subdomain pattern from the first domain
  // All domains should have the same subdomain pattern
  const subdomainPattern = extractSubdomainPattern(domainRoutes[0].domain)
  if (!subdomainPattern) {
    console.error(`Invalid domain format: ${domainRoutes[0].domain}`)
    return null
  }

  const wildcardPattern = `*.${subdomainPattern}`
  console.log(`Checking for wildcard certificate: ${wildcardPattern}`)

  // Import storage service dynamically to avoid circular dependencies
  const { findWildcardEdgeCertificate, saveEdgeCertificate } = await import(
    './supabase-storage-service'
  )

  // Check if wildcard certificate already exists
  const existingCert = await findWildcardEdgeCertificate(installationId, subdomainPattern)
  if (existingCert) {
    console.log(
      `Reusing existing wildcard certificate for ${wildcardPattern}: ${existingCert.cloudflareCertPackId}`,
    )
    return existingCert.cloudflareCertPackId
  }

  // No wildcard certificate exists, create a new one
  console.log(`No wildcard certificate found, creating new one for ${wildcardPattern}`)
  const certId = await orderEdgeCertificate([wildcardPattern])

  if (certId) {
    // Save the wildcard certificate to database
    await saveEdgeCertificate({
      installationId,
      cloudflareCertPackId: certId,
      hostnames: [wildcardPattern],
      domainRequestName,
      status: 'pending',
    })
    console.log(`Created and saved wildcard edge certificate for ${wildcardPattern}: ${certId}`)
    return certId
  }

  console.error(`Failed to create wildcard edge certificate for ${wildcardPattern}`)
  return null
}

/**
 * Create edge certificates for DomainRequest route domains
 * Creates edge certificates for domains that are proxied through CloudFlare
 *
 * Note: SSH domain (ssh.{name}.{subdomain}.{domain}) doesn't need an edge certificate
 * because it's not proxied through CloudFlare (direct A record)
 *
 * @param domainRoutes - Array of domain routes to create certificates for
 * @returns Array of certificate pack IDs
 */
export async function createDomainRequestEdgeCertificates(
  domainRoutes: Array<{ domain: string }>,
): Promise<string[]> {
  const certificateIds: string[] = []

  // Create individual edge certificate for each route domain
  // These domains are proxied via CNAME and need edge certificates for TLS termination
  for (const route of domainRoutes) {
    console.log(`Creating edge certificate for route domain: ${route.domain}`)

    const certId = await orderEdgeCertificate([route.domain])

    if (certId) {
      certificateIds.push(certId)
      console.log(`Edge certificate created for ${route.domain}: ${certId}`)
    } else {
      console.error(`Failed to create edge certificate for ${route.domain}`)
      // Continue with other domains even if one fails
    }
  }

  if (certificateIds.length > 0) {
    console.log(`Created ${certificateIds.length} edge certificates for domain routes`)
  } else {
    console.log('No edge certificates created (no domain routes or all failed)')
  }

  return certificateIds
}
