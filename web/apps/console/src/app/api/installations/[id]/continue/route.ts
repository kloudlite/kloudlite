import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationById } from '@/lib/console/supabase-storage-service'
import { SignJWT } from 'jose'
import { cookies } from 'next/headers'

/**
 * Continue API route - loads installation context and redirects to the appropriate step
 */
export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    return NextResponse.redirect(new URL('/installations/login', request.url))
  }

  // Fetch the installation
  const installation = await getInstallationById(id)

  if (!installation) {
    console.error('Installation not found:', id)
    return NextResponse.redirect(new URL('/installations', request.url))
  }

  // Verify user owns this installation
  if (installation.userId !== session.user.id) {
    return NextResponse.redirect(new URL('/installations', request.url))
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

  // Helper function to validate subdomain
  const isValidSubdomain = (subdomain: string | null | undefined): boolean => {
    if (!subdomain) return false
    if (subdomain === '0.0.0.0') return false
    if (subdomain.includes('0.0.0.0')) return false
    return true
  }

  // Determine next step based on installation status
  let redirectPath: string

  if (!installation.secretKey) {
    // Not installed yet - go to install step
    redirectPath = '/installations/new/install'
  } else if (!isValidSubdomain(installation.subdomain)) {
    // Installed but no valid domain - go to domain step
    redirectPath = '/installations/new/domain'
  } else if (!installation.deploymentReady) {
    // Has valid domain but not ready - go to complete step
    redirectPath = '/installations/new/complete'
  } else {
    // Fully set up - go back to installations list
    redirectPath = '/installations'
  }

  // Construct redirect URL using the request host headers
  const host = request.headers.get('host') || request.headers.get('x-forwarded-host')
  const protocol = request.headers.get('x-forwarded-proto') || 'https'
  const redirectUrl = host ? `${protocol}://${host}${redirectPath}` : new URL(redirectPath, request.url).toString()

  return NextResponse.redirect(redirectUrl)
}
