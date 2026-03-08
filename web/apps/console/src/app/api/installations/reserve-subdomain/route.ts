import { NextRequest, NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { cookies } from 'next/headers'
import { jwtVerify } from 'jose'
import {
  reserveSubdomain,
  getOrgInstallations,
  createInstallation,
  isOrgMember,
} from '@/lib/console/storage'

/**
 * Reserve subdomain for an organization's installation
 */
export async function POST(request: NextRequest) {
  try {
    // Verify registration session
    const cookieStore = await cookies()
    const token = cookieStore.get('registration_session')?.value

    if (!token) {
      return apiError('Not authenticated', 401)
    }

    const secret = new TextEncoder().encode(process.env.NEXTAUTH_SECRET)
    const { payload } = await jwtVerify(token, secret)

    const userEmail = payload.email as string
    const userName = payload.name as string
    const userId = payload.userId as string
    const sessionInstallationKey = payload.installationKey as string | undefined

    if (!userEmail || !userId) {
      return apiError('Invalid session', 401)
    }

    // Get subdomain and orgId from request body
    const body = await request.json()
    const { subdomain, orgId } = body

    if (!subdomain) {
      return apiError('Subdomain is required', 400)
    }

    if (!orgId) {
      return apiError('Organization ID is required', 400)
    }

    // Verify user is a member of the organization
    const isMember = await isOrgMember(orgId, userId)
    if (!isMember) {
      return apiError('Not a member of this organization', 403)
    }

    // Determine which installation to use
    let installationId: string
    const installations = await getOrgInstallations(orgId)

    if (sessionInstallationKey) {
      // Use the installation from session
      const installation = installations.find((i) => i.installationKey === sessionInstallationKey)
      if (!installation) {
        return apiError('Installation not found', 404)
      }
      installationId = installation.id
    } else {
      // Create a new installation if none exists
      const incompleteInstallation = installations.find((i) => !i.setupCompleted)
      if (incompleteInstallation) {
        installationId = incompleteInstallation.id
      } else {
        // Create new installation
        const generatedKey = crypto.randomUUID()
        const newInstallation = await createInstallation(
          orgId,
          'My Installation',
          undefined,
          generatedKey,
        )
        installationId = newInstallation.id
      }
    }

    // Reserve subdomain for the installation
    const reservation = await reserveSubdomain(
      subdomain,
      installationId,
      userId,
      userEmail,
      userName,
    )

    const response = NextResponse.json({
      success: true,
      subdomain: reservation.subdomain,
      url: `https://${reservation.subdomain}.${process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}`,
    })

    // Disable all caching
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')

    return response
  } catch (err) {
    console.error('Reserve subdomain error:', err)

    const error = err instanceof Error ? err : new Error('Unknown error')
    if (
      error.message === 'Subdomain is not available' ||
      error.message === 'Subdomain is already reserved'
    ) {
      return apiError('Subdomain is not available', 409)
    }

    return apiError('Internal server error', 500)
  }
}
