import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationById, getMemberRole } from '@/lib/console/storage'
import { SignJWT } from 'jose'
import { cookies } from 'next/headers'

/**
 * Load installation context into session - used by the continue dialog
 */
export async function POST(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    return apiError('Unauthorized', 401)
  }

  // Fetch the installation
  const installation = await getInstallationById(id)

  if (!installation) {
    return apiError('Installation not found', 404)
  }

  // Check if user has access to this installation (owner or team member)
  const isOwner = installation.userId === session.user.id
  const userRole = await getMemberRole(id, session.user.id)

  if (!isOwner && !userRole) {
    return apiError('Access denied', 403)
  }

  // Update session cookie. Only owners get installationKey context for installer callbacks.
  const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
  const sessionPayload: Record<string, string> = {
    provider: session.provider,
    email: session.user.email,
    name: session.user.name,
    image: session.user.image || '',
    userId: session.user.id,
  }
  if (isOwner) {
    sessionPayload.installationKey = installation.installationKey
  }
  const token = await new SignJWT(sessionPayload)
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
    installationKey: isOwner ? installation.installationKey : undefined,
    subdomain: installation.subdomain,
    deploymentReady: installation.deploymentReady,
  })
}
