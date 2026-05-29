import { randomUUID } from 'crypto'
import * as https from 'https'
import * as http from 'http'

const API_BASE = process.env.KLOUDLITE_API_URL || 'https://localhost:9443'
const JWT_SECRET = process.env.KLOUDLITE_JWT_SECRET || 'dev-jwt-secret-change-me'

// Simple JWT generation matching the Go backend (HS256)
function generateToken(username: string): string {
  const header = { typ: 'JWT', alg: 'HS256' }
  const now = Math.floor(Date.now() / 1000)
  const claims = {
    sub: username,
    email: `${username}@kloudlite.io`,
    name: username,
    username,
    isActive: true,
    iat: now,
    exp: now + 86400
  }

  const b64 = (obj: unknown) =>
    Buffer.from(JSON.stringify(obj))
      .toString('base64url')
      .replace(/=+$/, '')

  const headerB64 = b64(header)
  const claimsB64 = b64(claims)
  const signature = Buffer.from(
    require('crypto').createHmac('sha256', JWT_SECRET)
      .update(`${headerB64}.${claimsB64}`)
      .digest()
  ).toString('base64url').replace(/=+$/, '')

  return `${headerB64}.${claimsB64}.${signature}`
}

function apiRequest(
  method: string,
  path: string,
  body?: unknown
): Promise<unknown> {
  return new Promise((resolve, reject) => {
    const url = new URL(path, API_BASE)
    const token = generateToken('karthik')
    const payload = body ? JSON.stringify(body) : undefined

    const options = {
      hostname: url.hostname,
      port: url.port || 443,
      path: url.pathname + url.search,
      method,
      headers: {
        Authorization: `Bearer ${token}`,
        ...(payload ? { 'Content-Type': 'application/json' } : {}),
      } as Record<string, string>,
      rejectUnauthorized: false,
    }

    if (payload) {
      options.headers['Content-Length'] = String(Buffer.byteLength(payload))
    }

    const lib = API_BASE.startsWith('https') ? https : http
    const req = lib.request(options, (res) => {
      const chunks: Buffer[] = []
      res.on('data', (c: Buffer) => chunks.push(c))
      res.on('end', () => {
        const raw = Buffer.concat(chunks).toString()
        try {
          const parsed = JSON.parse(raw)
          if (res.statusCode && res.statusCode >= 400) {
            reject(new Error(parsed.error || `HTTP ${res.statusCode}`))
          } else {
            resolve(parsed)
          }
        } catch {
          if (res.statusCode && res.statusCode >= 400) {
            reject(new Error(`HTTP ${res.statusCode}`))
          } else {
            resolve(raw)
          }
        }
      })
    })

    req.on('error', reject)
    if (payload) req.write(payload)
    req.end()
  })
}

// Resource CRUD helpers
function resourcePath(namespace: string | null, resource: string, name?: string): string {
  const base = namespace
    ? `/api/v1/namespaces/${namespace}/resources/${resource}`
    : `/api/v1/resources/${resource}`
  return name ? `${base}/${name}` : base
}

export const platformAPI = {
  // Health check
  health: () => apiRequest('GET', '/health') as Promise<{ status: string }>,

  // Info
  info: () => apiRequest('GET', '/api/v1/info') as Promise<Record<string, unknown>>,

  // Environments
  listEnvironments: (namespace: string) =>
    apiRequest('GET', resourcePath(namespace, 'environments')) as Promise<{ items: Record<string, unknown>[] }>,

  getEnvironment: (namespace: string, name: string) =>
    apiRequest('GET', resourcePath(namespace, 'environments', name)) as Promise<Record<string, unknown>>,

  createEnvironment: (namespace: string, name: string, spec: Record<string, unknown>) =>
    apiRequest('POST', resourcePath(namespace, 'environments'), {
      apiVersion: 'environments.kloudlite.io/v1',
      kind: 'Environment',
      metadata: { name, namespace },
      spec,
    }) as Promise<Record<string, unknown>>,

  patchEnvironment: (namespace: string, name: string, patch: Record<string, unknown>) =>
    apiRequest('PATCH', resourcePath(namespace, 'environments', name), patch) as Promise<Record<string, unknown>>,

  deleteEnvironment: (namespace: string, name: string) =>
    apiRequest('DELETE', resourcePath(namespace, 'environments', name)) as Promise<void>,

  // Workspaces
  listWorkspaces: (namespace: string) =>
    apiRequest('GET', resourcePath(namespace, 'workspaces')) as Promise<{ items: Record<string, unknown>[] }>,

  getWorkspace: (namespace: string, name: string) =>
    apiRequest('GET', resourcePath(namespace, 'workspaces', name)) as Promise<Record<string, unknown>>,

  createWorkspace: (namespace: string, name: string, spec: Record<string, unknown>) =>
    apiRequest('POST', resourcePath(namespace, 'workspaces'), {
      apiVersion: 'workspaces.kloudlite.io/v1',
      kind: 'Workspace',
      metadata: { name, namespace },
      spec,
    }) as Promise<Record<string, unknown>>,

  deleteWorkspace: (namespace: string, name: string) =>
    apiRequest('DELETE', resourcePath(namespace, 'workspaces', name)) as Promise<void>,

  // Generic resource list (for any resource type)
  listResources: (namespace: string | null, resource: string) =>
    apiRequest('GET', resourcePath(namespace, resource)) as Promise<{ items: Record<string, unknown>[] }>,

  // WorkMachines
  listWorkMachines: () =>
    apiRequest('GET', resourcePath(null, 'workmachines')) as Promise<{ items: Record<string, unknown>[] }>,
}

export function generateAuthToken(username: string): string {
  return generateToken(username)
}
