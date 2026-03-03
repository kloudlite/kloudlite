import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import { createInstallation, cleanupExpiredInstallations, updateInstallation } from '@/lib/console/storage'
import { SignJWT } from 'jose'
import crypto from 'crypto'

export const runtime = 'nodejs'

/**
 * Create a new installation with name and description
 * This is called when user submits the name/description form
 */
export async function POST(request: Request) {
  try {
    const session = await getRegistrationSession()

    if (!session?.user) {
      return apiError('Not authenticated', 401)
    }

    // Cleanup any expired installations for this user before creating a new one
    const cleanedUp = await cleanupExpiredInstallations(session.user.id)
    if (cleanedUp > 0) {
      console.log(`Cleaned up ${cleanedUp} expired installation(s) for user ${session.user.id}`)
    }

    const body = await request.json()
    const { name, description, subdomain, hostingType } = body

    if (!name || typeof name !== 'string' || name.trim().length === 0) {
      return apiError('Installation name is required', 400)
    }

    if (!subdomain || typeof subdomain !== 'string' || subdomain.trim().length === 0) {
      return apiError('Subdomain is required', 400)
    }

    // Validate subdomain format
    const subdomainRegex = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/
    if (!subdomainRegex.test(subdomain.trim())) {
      return apiError('Invalid subdomain format', 400)
    }

    // Generate a new installation key
    const installationKey = crypto.randomUUID()

    // Create the installation with subdomain
    const installation = await createInstallation(
      session.user.id,
      name.trim(),
      description?.trim() || undefined,
      installationKey,
      subdomain.trim(),
    )

    // If Kloudlite Cloud hosting, set cloud provider to OCI
    if (hostingType === 'kloudlite') {
      await updateInstallation(installation.id, { cloudProvider: 'oci' })
    }

    // Update the session cookie with the installation key, subdomain, and hosting type
    const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
    const token = await new SignJWT({
      provider: session.provider,
      email: session.user.email,
      name: session.user.name,
      image: session.user.image,
      installationKey: installation.installationKey,
      subdomain: subdomain.trim(),
      userId: session.user.id,
      hostingType: hostingType || 'byoc',
    })
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('30d')
      .sign(secret)

    const response = NextResponse.json({
      success: true,
      installationKey: installation.installationKey,
      installationId: installation.id,
    })

    // Update the session cookie
    response.cookies.set('registration_session', token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 30 * 24 * 60 * 60, // 30 days
    })

    return response
  } catch (error) {
    console.error('Error creating installation:', error)
    return apiError('Internal server error', 500)
  }
}
