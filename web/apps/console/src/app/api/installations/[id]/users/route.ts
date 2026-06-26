import { NextRequest, NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { cachedInstallationAccess } from '@/lib/console/cached-queries'
import { getTenantClient } from '@/lib/console/tenant-client'

export const runtime = 'nodejs'

export async function GET(_req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const session = await getRegistrationSession()
  if (!session?.user) return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })

  try {
    await cachedInstallationAccess(id)
    const client = await getTenantClient(id)
    const users = await client.listUsers()
    return NextResponse.json({ users })
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Unknown error'
    return NextResponse.json({ error: message }, { status: 500 })
  }
}

export async function POST(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const session = await getRegistrationSession()
  if (!session?.user) return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })

  try {
    await cachedInstallationAccess(id)
    const { email, name, role } = await req.json()
    if (!email) return NextResponse.json({ error: 'Email is required' }, { status: 400 })

    const client = await getTenantClient(id)
    const user = await client.createUser(email.replace(/[@.]/g, '-'), {
      email,
      displayName: name || email.split('@')[0],
      roles: [role || 'member'],
      active: true,
    })
    return NextResponse.json({ user })
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Unknown error'
    return NextResponse.json({ error: message }, { status: 500 })
  }
}

export async function DELETE(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const session = await getRegistrationSession()
  if (!session?.user) return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })

  try {
    await cachedInstallationAccess(id)
    const { name } = await req.json()
    if (!name) return NextResponse.json({ error: 'User name is required' }, { status: 400 })

    const client = await getTenantClient(id)
    await client.deleteUser(name)
    return NextResponse.json({ success: true })
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Unknown error'
    return NextResponse.json({ error: message }, { status: 500 })
  }
}
