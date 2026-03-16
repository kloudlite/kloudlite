import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { requireInstallationAccess } from '@/lib/console/authorization'
import { getInstallationById } from '@/lib/console/storage'
import { SignJWT } from 'jose'
import { cookies } from 'next/headers'

function getPublicOrigin(request: Request): string {
  const proto = request.headers.get('x-forwarded-proto') || 'https'
  const host = request.headers.get('x-forwarded-host') || request.headers.get('host') || ''
  return `${proto}://${host}`
}

/**
 * Continue API route - loads installation context and redirects to the appropriate step
 */
export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const origin = getPublicOrigin(request)
  const session = await getRegistrationSession()

  if (!session?.user) {
    return NextResponse.redirect(new URL('/login', origin))
  }

  // Verify user has access via org membership
  try {
    await requireInstallationAccess(id)
  } catch {
    return NextResponse.redirect(new URL('/installations', origin))
  }

  // Fetch the installation
  const installation = await getInstallationById(id)

  if (!installation) {
    console.error('Installation not found:', id)
    return NextResponse.redirect(new URL('/installations', origin))
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

  // Determine next step based on installation status
  let redirectPath: string

  if (installation.deploymentReady) {
    redirectPath = `/installations/${id}`
  } else {
    redirectPath = `/installations/${id}/install`
  }

  return NextResponse.redirect(new URL(redirectPath, origin))
}
