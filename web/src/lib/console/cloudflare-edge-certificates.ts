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
