/**
 * Cloudflare KV REST API Client
 *
 * This service uses Cloudflare's REST API to interact with KV storage
 * from Next.js server-side code.
 */

const CLOUDFLARE_ACCOUNT_ID = process.env.CLOUDFLARE_ACCOUNT_ID!
const CLOUDFLARE_API_TOKEN = process.env.CLOUDFLARE_API_TOKEN!
const KV_NAMESPACE_ID = process.env.CLOUDFLARE_KV_NAMESPACE_ID!

const KV_API_BASE = `https://api.cloudflare.com/client/v4/accounts/${CLOUDFLARE_ACCOUNT_ID}/storage/kv/namespaces/${KV_NAMESPACE_ID}`

interface CloudflareKVResponse<T> {
  success: boolean
  errors: Array<{ code: number; message: string }>
  messages: string[]
  result: T
}

/**
 * Get a value from KV
 */
export async function kvGet<T>(key: string): Promise<T | null> {
  try {
    const response = await fetch(`${KV_API_BASE}/values/${key}`, {
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
      },
      cache: 'no-store', // Force fresh data, no Next.js caching
      next: { revalidate: 0 }, // Disable Next.js data cache
    })

    if (response.status === 404) {
      return null
    }

    if (!response.ok) {
      throw new Error(`KV GET failed: ${response.statusText}`)
    }

    const text = await response.text()

    // Try to parse as JSON, otherwise return as string
    try {
      return JSON.parse(text) as T
    } catch {
      return text as T
    }
  } catch (error) {
    console.error('KV GET error:', error)
    throw error
  }
}

type JsonValue = string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue }

/**
 * Put a value into KV
 */
export async function kvPut(key: string, value: JsonValue, expirationTtl?: number): Promise<void> {
  try {
    const formData = new FormData()

    // Convert value to string if it's an object
    const stringValue = typeof value === 'string' ? value : JSON.stringify(value)
    formData.append('value', stringValue)
    formData.append('metadata', JSON.stringify({}))

    if (expirationTtl) {
      formData.append('expiration_ttl', expirationTtl.toString())
    }

    const response = await fetch(`${KV_API_BASE}/values/${key}`, {
      method: 'PUT',
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
      },
      body: formData,
    })

    if (!response.ok) {
      const error = await response.text()
      throw new Error(`KV PUT failed: ${error}`)
    }

    const result: CloudflareKVResponse<null> = await response.json()

    if (!result.success) {
      throw new Error(`KV PUT failed: ${result.errors[0]?.message || 'Unknown error'}`)
    }
  } catch (error) {
    console.error('KV PUT error:', error)
    throw error
  }
}

/**
 * Delete a value from KV
 */
export async function kvDelete(key: string): Promise<void> {
  try {
    const response = await fetch(`${KV_API_BASE}/values/${key}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
      },
    })

    if (!response.ok && response.status !== 404) {
      throw new Error(`KV DELETE failed: ${response.statusText}`)
    }
  } catch (error) {
    console.error('KV DELETE error:', error)
    throw error
  }
}

/**
 * List keys in KV (with optional prefix)
 */
export async function kvList(prefix?: string): Promise<string[]> {
  try {
    const url = new URL(`${KV_API_BASE}/keys`)
    if (prefix) {
      url.searchParams.set('prefix', prefix)
    }

    const response = await fetch(url.toString(), {
      headers: {
        Authorization: `Bearer ${CLOUDFLARE_API_TOKEN}`,
      },
    })

    if (!response.ok) {
      throw new Error(`KV LIST failed: ${response.statusText}`)
    }

    const result: CloudflareKVResponse<Array<{ name: string }>> = await response.json()

    if (!result.success) {
      throw new Error(`KV LIST failed: ${result.errors[0]?.message || 'Unknown error'}`)
    }

    return result.result.map((item) => item.name)
  } catch (error) {
    console.error('KV LIST error:', error)
    throw error
  }
}
