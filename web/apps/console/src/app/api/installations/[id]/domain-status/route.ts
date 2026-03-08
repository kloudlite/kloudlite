import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { requireInstallationAccess } from '@/lib/console/authorization'
import { getInstallationById, checkInstallationDomainStatus } from '@/lib/console/storage'

/**
 * Check if an installation's domain has expired and been claimed by another user
 * Returns the installation info and whether domain re-selection is needed
 */
export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  // Verify access via org membership
  try {
    await requireInstallationAccess(id)
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

  // If installation is already deployed (deploymentReady=true), domain is locked
  if (installation.deploymentReady) {
    return NextResponse.json({
      needsReselection: false,
      name: installation.name,
      description: installation.description,
      subdomain: installation.subdomain,
    })
  }

  // If no subdomain yet, no reselection needed (they haven't picked one yet)
  if (!installation.subdomain) {
    return NextResponse.json({
      needsReselection: false,
      name: installation.name,
      description: installation.description,
      subdomain: null,
    })
  }

  // Check if the domain has expired and been claimed by another user
  const domainStatus = await checkInstallationDomainStatus(id, installation.subdomain)

  if (domainStatus.isExpired && domainStatus.isClaimedByOther) {
    return NextResponse.json({
      needsReselection: true,
      name: installation.name,
      description: installation.description,
      oldSubdomain: installation.subdomain,
      claimedByEmail: domainStatus.claimedByEmail,
      claimedByName: domainStatus.claimedByName,
    })
  }

  // Domain is still valid (not expired, or expired but not claimed)
  return NextResponse.json({
    needsReselection: false,
    name: installation.name,
    description: installation.description,
    subdomain: installation.subdomain,
    isExpired: domainStatus.isExpired,
  })
}
