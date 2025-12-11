import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationById, checkInstallationDomainStatus } from '@/lib/console/supabase-storage-service'

/**
 * Check if an installation's domain has expired and been claimed by another user
 * Returns the installation info and whether domain re-selection is needed
 */
export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
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

  // Verify user owns this installation
  if (installation.userId !== session.user.id) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
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
