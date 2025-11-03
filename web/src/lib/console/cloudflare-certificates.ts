/**
 * Cloudflare Origin CA Certificate Service
 *
 * Generates TLS certificates using Cloudflare's Origin CA API
 * These certificates are trusted by Cloudflare's edge for origin connections
 */

import { exec } from 'child_process'
import { promisify } from 'util'
import { writeFile, unlink } from 'fs/promises'
import { join } from 'path'
import { tmpdir } from 'os'

const execAsync = promisify(exec)

const CLOUDFLARE_ORIGIN_CA_KEY = process.env.CLOUDFLARE_ORIGIN_CA_KEY!
const CLOUDFLARE_ORIGIN_CA_API = 'https://api.cloudflare.com/client/v4/certificates'

interface CloudflareCertificateResponse {
  success: boolean
  errors: Array<{ code: number; message: string }>
  messages: string[]
  result: {
    id: string
    certificate: string
    hostnames: string[]
    expires_on: string
    request_type: string
    requested_validity: number
  }
}

export interface TLSCertificate {
  id: string
  certificate: string
  privateKey: string
  hostnames: string[]
  validFrom: string
  validUntil: string
}

/**
 * Generate a private key and CSR (Certificate Signing Request) using openssl
 */
async function generateCSR(hostnames: string[]): Promise<{ privateKey: string; csr: string }> {
  const keyFile = join(tmpdir(), `key-${Date.now()}.pem`)
  const csrFile = join(tmpdir(), `csr-${Date.now()}.pem`)
  const configFile = join(tmpdir(), `config-${Date.now()}.cnf`)

  try {
    // Create OpenSSL config with SANs
    const sanList = hostnames.map((h, i) => `DNS.${i + 1} = ${h}`).join('\n')
    const config = `[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = ${hostnames[0]}

[v3_req]
subjectAltName = @alt_names

[alt_names]
${sanList}`

    await writeFile(configFile, config)

    // Generate private key
    await execAsync(`openssl genrsa -out ${keyFile} 2048`)

    // Generate CSR
    await execAsync(`openssl req -new -key ${keyFile} -out ${csrFile} -config ${configFile}`)

    // Read the generated files
    const { readFile } = await import('fs/promises')
    const privateKey = await readFile(keyFile, 'utf-8')
    const csr = await readFile(csrFile, 'utf-8')

    // Clean up temp files
    await Promise.all([unlink(keyFile), unlink(csrFile), unlink(configFile)])

    return { privateKey, csr }
  } catch (error) {
    // Clean up on error
    try {
      await Promise.all([unlink(keyFile), unlink(csrFile), unlink(configFile)])
    } catch {}
    throw error
  }
}

/**
 * Generate a certificate using Cloudflare Origin CA API
 */
export async function generateCertificate(
  hostnames: string[],
  validityDays: number = 5475, // 15 years (max for Cloudflare Origin CA)
): Promise<TLSCertificate | null> {
  try {
    console.log(`Generating Cloudflare Origin CA certificate for hostnames:`, hostnames)

    // Generate private key and CSR
    const { privateKey, csr } = await generateCSR(hostnames)

    const response = await fetch(CLOUDFLARE_ORIGIN_CA_API, {
      method: 'POST',
      headers: {
        'X-Auth-User-Service-Key': CLOUDFLARE_ORIGIN_CA_KEY,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        csr,
        hostnames,
        requested_validity: validityDays,
        request_type: 'origin-rsa', // RSA 2048-bit key
      }),
    })

    if (!response.ok) {
      const error = await response.text()
      console.error(`Cloudflare certificate generation failed:`, error)
      return null
    }

    const result: CloudflareCertificateResponse = await response.json()

    if (!result.success) {
      console.error(`Cloudflare certificate generation failed:`, result.errors)
      return null
    }

    const cert: TLSCertificate = {
      id: result.result.id,
      certificate: result.result.certificate,
      privateKey, // Use our generated private key
      hostnames: result.result.hostnames,
      validFrom: new Date().toISOString(),
      validUntil: new Date(result.result.expires_on).toISOString(),
    }

    console.log(`Certificate generated successfully for hostnames:`, hostnames)
    console.log(`Certificate ID: ${cert.id}`)
    console.log(`Valid until: ${cert.validUntil}`)

    return cert
  } catch (error) {
    console.error(`Certificate generation error:`, error)
    return null
  }
}

/**
 * Revoke a certificate by ID
 */
export async function revokeCertificate(certificateId: string): Promise<boolean> {
  try {
    console.log(`Revoking certificate: ${certificateId}`)

    const response = await fetch(`${CLOUDFLARE_ORIGIN_CA_API}/${certificateId}`, {
      method: 'DELETE',
      headers: {
        'X-Auth-User-Service-Key': CLOUDFLARE_ORIGIN_CA_KEY,
      },
    })

    if (!response.ok && response.status !== 404) {
      const error = await response.text()
      console.error(`Certificate revocation failed:`, error)
      return false
    }

    console.log(`Certificate revoked successfully: ${certificateId}`)
    return true
  } catch (error) {
    console.error(`Certificate revocation error:`, error)
    return false
  }
}

/**
 * Get certificate information by ID
 */
export async function getCertificate(certificateId: string): Promise<TLSCertificate | null> {
  try {
    const response = await fetch(`${CLOUDFLARE_ORIGIN_CA_API}/${certificateId}`, {
      headers: {
        'X-Auth-User-Service-Key': CLOUDFLARE_ORIGIN_CA_KEY,
      },
    })

    if (!response.ok) {
      return null
    }

    const result: CloudflareCertificateResponse = await response.json()

    if (!result.success) {
      return null
    }

    return {
      id: result.result.id,
      certificate: result.result.certificate,
      privateKey: '', // Private key not available when retrieving existing certificates
      hostnames: result.result.hostnames,
      validFrom: new Date().toISOString(),
      validUntil: new Date(result.result.expires_on).toISOString(),
    }
  } catch (error) {
    console.error(`Get certificate error:`, error)
    return null
  }
}

/**
 * List all origin certificates
 */
export async function listAllCertificates(): Promise<Array<{ id: string; hostnames: string[] }>> {
  try {
    console.log('Fetching all origin certificates from CloudFlare')

    const response = await fetch(CLOUDFLARE_ORIGIN_CA_API, {
      method: 'GET',
      headers: {
        'X-Auth-User-Service-Key': CLOUDFLARE_ORIGIN_CA_KEY,
      },
    })

    if (!response.ok) {
      const error = await response.text()
      console.error('Failed to list certificates:', error)
      return []
    }

    const result = await response.json()

    if (!result.success) {
      console.error('Failed to list certificates:', result.errors)
      return []
    }

    return result.result.map((cert: any) => ({
      id: cert.id,
      hostnames: cert.hostnames,
    }))
  } catch (error) {
    console.error('List certificates error:', error)
    return []
  }
}

export type CertificateScope = 'installation' | 'workmachine' | 'workspace'

/**
 * Generate hostnames for certificates at different domain levels
 *
 * @param subdomain - User's subdomain (e.g., "hello-wrold")
 * @param baseDomain - Base domain (e.g., "khost.dev")
 * @param scope - Certificate scope level
 * @param scopeIdentifier - Identifier for the scope (wm-user for workmachine, workspace name for workspace)
 * @param parentScopeIdentifier - Parent scope identifier (only for workspace scope - the wm-user)
 *
 * Examples:
 * - scope="installation": ["hello-wrold.khost.dev", "*.hello-wrold.khost.dev"]
 *   Covers: hello-wrold.khost.dev, dev1.hello-wrold.khost.dev
 *
 * - scope="workmachine", scopeIdentifier="dev1": ["dev1.hello-wrold.khost.dev", "*.dev1.hello-wrold.khost.dev"]
 *   Covers: dev1.hello-wrold.khost.dev, workspace1.dev1.hello-wrold.khost.dev
 *
 * - scope="workspace", scopeIdentifier="workspace1", parentScopeIdentifier="dev1":
 *   ["workspace1.dev1.hello-wrold.khost.dev", "*.workspace1.dev1.hello-wrold.khost.dev"]
 *   Covers: workspace1.dev1.hello-wrold.khost.dev, vscode.workspace1.dev1.hello-wrold.khost.dev
 */
export function generateHostnames(
  subdomain: string,
  baseDomain: string,
  scope: CertificateScope = 'installation',
  scopeIdentifier?: string,
  parentScopeIdentifier?: string,
): string[] {
  switch (scope) {
    case 'installation':
      return [`${subdomain}.${baseDomain}`, `*.${subdomain}.${baseDomain}`]

    case 'workmachine':
      if (!scopeIdentifier) {
        throw new Error('scopeIdentifier (wm-user) is required for workmachine scope')
      }
      return [
        `${scopeIdentifier}.${subdomain}.${baseDomain}`,
        `*.${scopeIdentifier}.${subdomain}.${baseDomain}`,
      ]

    case 'workspace':
      if (!scopeIdentifier || !parentScopeIdentifier) {
        throw new Error(
          'scopeIdentifier (workspace) and parentScopeIdentifier (wm-user) are required for workspace scope',
        )
      }
      return [
        `${scopeIdentifier}.${parentScopeIdentifier}.${subdomain}.${baseDomain}`,
        `*.${scopeIdentifier}.${parentScopeIdentifier}.${subdomain}.${baseDomain}`,
      ]

    default:
      throw new Error(`Invalid certificate scope: ${scope}`)
  }
}
