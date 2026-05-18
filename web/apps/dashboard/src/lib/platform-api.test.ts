import { describe, expect, test, vi } from 'vitest'
import { createPlatformApiClient } from './platform-api'

describe('platform-api dashboard client', () => {
  test('lists cluster resources from platform-api resource endpoints', async () => {
    const fetch = vi.fn(async () => new Response(JSON.stringify({ items: [{ metadata: { name: 'alice' } }] }), { status: 200 }))
    const client = createPlatformApiClient({ baseUrl: 'https://api-server.kloudlite.svc.cluster.local', fetch })

    const users = await client.listClusterResources('users')

    expect(users).toEqual([{ metadata: { name: 'alice' } }])
    expect(fetch).toHaveBeenCalledWith('https://api-server.kloudlite.svc.cluster.local/api/v1/resources/users', expect.objectContaining({ method: 'GET' }))
  })

  test('creates cluster resources through platform-api', async () => {
    const created = { metadata: { name: 'alice' }, spec: { email: 'alice@example.com' } }
    const fetch = vi.fn(async () => new Response(JSON.stringify(created), { status: 201 }))
    const client = createPlatformApiClient({ baseUrl: 'https://api-server.kloudlite.svc.cluster.local', token: 'signed-token', fetch })

    await client.createClusterResource('users', created)

    expect(fetch).toHaveBeenCalledWith(
      'https://api-server.kloudlite.svc.cluster.local/api/v1/resources/users',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ Authorization: 'Bearer signed-token' }),
        body: JSON.stringify(created),
      }),
    )
  })
})
