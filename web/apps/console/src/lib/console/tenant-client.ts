import { SignJWT } from 'jose'
import { cachedInstallationById } from '@/lib/console/cached-queries'
import { getRegistrationSession } from '@/lib/console-auth'

const JWT_SECRET = new TextEncoder().encode(process.env.NEXTAUTH_SECRET || 'your-secret-key')

async function getAuthToken(userId: string, email: string, role: string): Promise<string> {
  return await new SignJWT({ sub: userId, email, roles: [role] })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('1h')
    .sign(JWT_SECRET)
}

export async function getTenantClient(installationId: string) {
  const session = await getRegistrationSession()
  if (!session?.user) throw new Error('Not authenticated')

  const installation = await cachedInstallationById(installationId)
  if (!installation) throw new Error('Installation not found')

  const apiUrl = installation.apiServerUrl
  if (!apiUrl) throw new Error('Installation has no API server URL')

  const token = await getAuthToken(session.user.id, session.user.email, 'admin')

  return {
    apiUrl,
    token,
    async request<T>(method: string, path: string, body?: unknown): Promise<T> {
      const res = await fetch(`${apiUrl}/api/v1${path}`, {
        method,
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: body ? JSON.stringify(body) : undefined,
      })
      if (!res.ok) {
        const err = await res.text().catch(() => 'Unknown error')
        throw new Error(`Tenant API error (${res.status}): ${err}`)
      }
      return res.json()
    },
    async listUsers() {
      return this.request<any[]>('GET', '/resources/users')
    },
    async createUser(name: string, spec: Record<string, unknown>) {
      return this.request<any>('POST', '/resources/users', {
        metadata: { name },
        spec,
      })
    },
    async deleteUser(name: string) {
      return this.request<any>('DELETE', `/resources/users/${name}`)
    },
  }
}
