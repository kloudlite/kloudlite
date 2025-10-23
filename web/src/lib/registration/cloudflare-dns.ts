/**
 * Cloudflare DNS REST API Client
 *
 * Manages DNS A records for Kloudlite installations and workmachines
 */

const CLOUDFLARE_API_TOKEN = process.env.CLOUDFLARE_API_TOKEN!
const CLOUDFLARE_ZONE_ID = process.env.CLOUDFLARE_ZONE_ID!
const CLOUDFLARE_DNS_DOMAIN = process.env.CLOUDFLARE_DNS_DOMAIN!

const DNS_API_BASE = `https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records`

interface CloudflareDnsResponse<T> {
  success: boolean
  errors: Array<{ code: number; message: string }>
  messages: string[]
  result: T
}

interface DnsRecord {
  id: string
  type: string
  name: string
  content: string
  proxied: boolean
  ttl: number
}

/**
 * Create a DNS A record
 * @param name - Full domain name (e.g., "test.khost.dev" or "*.user1.test.khost.dev")
 * @param ip - IP address
 * @param proxied - Whether to proxy through Cloudflare (default: false)
 * @returns DNS record ID
 */
export async function createDnsRecord(
  name: string,
  ip: string,
  proxied: boolean = false,
): Promise<string | null> {
  try {
    console.log(`Creating DNS A record: ${name} → ${ip}`)

    const response = await fetch(DNS_API_BASE, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        type: 'A',
        name,
        content: ip,
        ttl: 120, // 2 minutes for faster propagation
        proxied,
      }),
    })

    if (!response.ok) {
      const error = await response.text()
      console.error(`DNS CREATE failed for ${name}:`, error)
      return null
    }

    const result: CloudflareDnsResponse<DnsRecord> = await response.json()

    if (!result.success) {
      console.error(`DNS CREATE failed for ${name}:`, result.errors)
      return null
    }

    console.log(`DNS A record created successfully: ${name} → ${ip} (ID: ${result.result.id})`)
    return result.result.id
  } catch (error) {
    console.error(`DNS CREATE error for ${name}:`, error)
    return null
  }
}

/**
 * Update an existing DNS A record
 * @param recordId - DNS record ID
 * @param name - Full domain name
 * @param ip - New IP address
 * @param proxied - Whether to proxy through Cloudflare
 * @returns Success status
 */
export async function updateDnsRecord(
  recordId: string,
  name: string,
  ip: string,
  proxied: boolean = false,
): Promise<boolean> {
  try {
    console.log(`Updating DNS A record ${recordId}: ${name} → ${ip}`)

    const response = await fetch(`${DNS_API_BASE}/${recordId}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        type: 'A',
        name,
        content: ip,
        ttl: 120,
        proxied,
      }),
    })

    if (!response.ok) {
      const error = await response.text()
      console.error(`DNS UPDATE failed for ${recordId}:`, error)
      return false
    }

    const result: CloudflareDnsResponse<DnsRecord> = await response.json()

    if (!result.success) {
      console.error(`DNS UPDATE failed for ${recordId}:`, result.errors)
      return false
    }

    console.log(`DNS A record updated successfully: ${name} → ${ip}`)
    return true
  } catch (error) {
    console.error(`DNS UPDATE error for ${recordId}:`, error)
    return false
  }
}

/**
 * Delete a DNS A record
 * @param recordId - DNS record ID
 * @returns Success status
 */
export async function deleteDnsRecord(recordId: string): Promise<boolean> {
  try {
    console.log(`Deleting DNS A record: ${recordId}`)

    const response = await fetch(`${DNS_API_BASE}/${recordId}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
      },
    })

    if (!response.ok && response.status !== 404) {
      const error = await response.text()
      console.error(`DNS DELETE failed for ${recordId}:`, error)
      return false
    }

    console.log(`DNS A record deleted successfully: ${recordId}`)
    return true
  } catch (error) {
    console.error(`DNS DELETE error for ${recordId}:`, error)
    return false
  }
}

/**
 * Get a DNS record by name
 * @param name - Full domain name
 * @returns DNS record or null if not found
 */
export async function getDnsRecord(name: string): Promise<DnsRecord | null> {
  try {
    const url = new URL(DNS_API_BASE)
    url.searchParams.set('name', name)
    url.searchParams.set('type', 'A')

    const response = await fetch(url.toString(), {
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
      },
    })

    if (!response.ok) {
      return null
    }

    const result: CloudflareDnsResponse<DnsRecord[]> = await response.json()

    if (!result.success || result.result.length === 0) {
      return null
    }

    return result.result[0]
  } catch (error) {
    console.error(`DNS GET error for ${name}:`, error)
    return null
  }
}

/**
 * Create DNS A records for installation
 * @param subdomain - User's subdomain (e.g., "test")
 * @param ip - Installation IP address
 * @returns Array of created DNS record IDs
 */
export async function createInstallationDnsRecords(
  subdomain: string,
  ip: string,
): Promise<string[]> {
  const recordIds: string[] = []

  // Create: {subdomain}.{domain} → IP
  const fullDomain = `${subdomain}.${CLOUDFLARE_DNS_DOMAIN}`
  const recordId = await createDnsRecord(fullDomain, ip)

  if (recordId) {
    recordIds.push(recordId)
  }

  return recordIds
}

/**
 * Create DNS A records for workmachine
 * @param workMachineName - Workmachine name (e.g., "user1")
 * @param subdomain - User's subdomain (e.g., "test")
 * @param ip - Workmachine IP address
 * @returns Array of created DNS record IDs
 */
export async function createWorkmachineDnsRecords(
  workMachineName: string,
  subdomain: string,
  ip: string,
): Promise<string[]> {
  const recordIds: string[] = []

  // Create: {workMachineName}.{subdomain}.{domain} → IP
  const exactDomain = `${workMachineName}.${subdomain}.${CLOUDFLARE_DNS_DOMAIN}`
  const exactRecordId = await createDnsRecord(exactDomain, ip)

  if (exactRecordId) {
    recordIds.push(exactRecordId)
  }

  // Create: *.{workMachineName}.{subdomain}.{domain} → IP
  const wildcardDomain = `*.${workMachineName}.${subdomain}.${CLOUDFLARE_DNS_DOMAIN}`
  const wildcardRecordId = await createDnsRecord(wildcardDomain, ip)

  if (wildcardRecordId) {
    recordIds.push(wildcardRecordId)
  }

  return recordIds
}

/**
 * Update DNS A records with new IP
 * @param recordIds - Existing DNS record IDs
 * @param name - Domain name (for logging)
 * @param ip - New IP address
 * @returns Success status
 */
export async function updateDnsRecords(
  recordIds: string[],
  name: string,
  ip: string,
): Promise<boolean> {
  let allSucceeded = true

  for (const recordId of recordIds) {
    const success = await updateDnsRecord(recordId, name, ip)
    if (!success) {
      allSucceeded = false
    }
  }

  return allSucceeded
}

/**
 * Delete multiple DNS A records
 * @param recordIds - DNS record IDs to delete
 * @returns Success status
 */
export async function deleteDnsRecords(recordIds: string[]): Promise<boolean> {
  let allSucceeded = true

  for (const recordId of recordIds) {
    const success = await deleteDnsRecord(recordId)
    if (!success) {
      allSucceeded = false
    }
  }

  return allSucceeded
}
