import { SignJWT } from 'jose'

type Fetch = typeof fetch

interface PlatformApiClientOptions {
  baseUrl?: string
  token?: string
  fetch?: Fetch
}

interface ListResponse<T> {
  items?: T[]
}

export interface KubernetesResource<TSpec = Record<string, unknown>> {
  apiVersion?: string
  kind?: string
  metadata: {
    name: string
    namespace?: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
  }
  spec: TSpec
  status?: Record<string, unknown>
}

export function createPlatformApiClient(options: PlatformApiClientOptions = {}) {
  const baseUrl = (options.baseUrl || process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'https://localhost:9443').replace(/\/$/, '')
  const fetchImpl = options.fetch || fetch
  const token = options.token || process.env.PLATFORM_API_TOKEN

  async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      Accept: 'application/json',
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...(init.headers as Record<string, string> | undefined),
    }
    if (token) headers.Authorization = `Bearer ${token}`

    const response = await fetchImpl(`${baseUrl}${path}`, {
      ...init,
      headers,
    })

    if (!response.ok) {
      const body = await response.text().catch(() => '')
      throw new Error(body || `platform-api request failed with ${response.status}`)
    }

    if (response.status === 204) return undefined as T
    return (await response.json()) as T
  }

  return {
    async listClusterResources<T>(resource: string): Promise<T[]> {
      const response = await request<ListResponse<T>>(`/api/v1/resources/${resource}`, { method: 'GET' })
      return response.items || []
    },
    getClusterResource<T>(resource: string, name: string): Promise<T> {
      return request<T>(`/api/v1/resources/${resource}/${encodeURIComponent(name)}`, { method: 'GET' })
    },
    createClusterResource<T>(resource: string, object: unknown): Promise<T> {
      return request<T>(`/api/v1/resources/${resource}`, {
        method: 'POST',
        body: JSON.stringify(object),
      })
    },
    patchClusterResource<T>(resource: string, name: string, patch: unknown): Promise<T> {
      return request<T>(`/api/v1/resources/${resource}/${encodeURIComponent(name)}`, {
        method: 'PATCH',
        body: JSON.stringify(patch),
      })
    },
    deleteClusterResource(resource: string, name: string): Promise<void> {
      return request<void>(`/api/v1/resources/${resource}/${encodeURIComponent(name)}`, { method: 'DELETE' })
    },
  }
}

export async function createPlatformApiToken() {
  const secret = process.env.JWT_SECRET
  if (!secret) return undefined
  return new SignJWT({ sub: 'dashboard-admin', aud: 'platform-api' })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('10m')
    .sign(new TextEncoder().encode(secret))
}

export async function getPlatformApiClient() {
  return createPlatformApiClient({ token: await createPlatformApiToken() })
}
