import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getErrorMessage } from '@/lib/errors'
import { requireInstallationOwner } from '@/lib/console/authorization'
import { getInstallationById, reReserveSubdomain, getUserById } from '@/lib/console/storage'

/**
 * Re-reserve a new subdomain for an installation whose previous domain expired
 * and was claimed by another user
 */
export async function POST(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  // Verify owner access via org membership
  let accessContext
  try {
    accessContext = await requireInstallationOwner(id)
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Unauthorized'
    if (message.includes('No session')) return apiError('Unauthorized', 401)
    if (message.includes('Not found')) return apiError('Installation not found', 404)
    return apiError('Forbidden', 403)
  }

  // Fetch the installation details
  const installation = await getInstallationById(id)

  if (!installation) {
    return apiError('Installation not found', 404)
  }

  // Cannot change domain if installation is already deployed
  if (installation.deploymentReady) {
    return apiError('Cannot change domain for a deployed installation', 400)
  }

  // Get the new subdomain from the request body
  const body = await request.json()
  const { subdomain } = body

  if (!subdomain) {
    return apiError('Subdomain is required', 400)
  }

  // Validate subdomain format
  const subdomainRegex = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/
  if (!subdomainRegex.test(subdomain)) {
    return apiError('Invalid subdomain format', 400)
  }

  if (subdomain.length < 3 || subdomain.length > 63) {
    return apiError('Subdomain must be between 3 and 63 characters', 400)
  }

  try {
    // Get user details for the reservation record
    const user = await getUserById(accessContext.userId)

    const reservation = await reReserveSubdomain(
      id,
      subdomain,
      accessContext.userId,
      user?.email || '',
      user?.name || '',
    )

    const domain = process.env.CLOUDFLARE_DNS_DOMAIN || 'khost.dev'

    return NextResponse.json({
      success: true,
      subdomain: reservation.subdomain,
      url: `https://${reservation.subdomain}.${domain}`,
    })
  } catch (error) {
    console.error('Error re-reserving subdomain:', error)
    return apiError(getErrorMessage(error, 'Failed to reserve subdomain'), 500)
  }
}
