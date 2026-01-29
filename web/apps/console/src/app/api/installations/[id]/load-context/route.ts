import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationById, getMemberRole } from '@/lib/console/supabase-storage-service'
import { SignJWT } from 'jose'
import { cookies } from 'next/headers'

/**
 * Load installation context into session - used by the continue dialog
 */
export async function POST(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  // Fetch the installation
  const installation = await getInstallationById(id)

  if (!installation) {
    return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
  }

  // Check if user has access to this installation (owner or team member)
  const isOwner = installation.userId === session.user.id
  const userRole = await getMemberRole(id, session.user.id)

  console.log('load-context access check:', {
    installationUserId: installation.userId,
    sessionUserId: session.user.id,
    isOwner,
    userRole,
  })

  if (!isOwner && !userRole) {
    return NextResponse.json({ error: 'Access denied' }, { status: 403 })
  }

  // Update session cookie with this installation's key
  const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
  const token = await new SignJWT({
    provider: session.provider,
    email: session.user.email,
    name: session.user.name,
    image: session.user.image,
    installationKey: installation.installationKey,
    userId: session.user.id,
  })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('30d')
    .sign(secret)

  const cookieStore = await cookies()
  cookieStore.set('registration_session', token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 30 * 24 * 60 * 60, // 30 days
  })

  return NextResponse.json({
    success: true,
    installationKey: installation.installationKey,
    subdomain: installation.subdomain,
    deploymentReady: installation.deploymentReady,
  })
}
