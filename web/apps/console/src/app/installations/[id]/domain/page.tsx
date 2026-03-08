import { redirect, notFound } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { cachedInstallationAccess, cachedInstallationById } from '@/lib/console/cached-queries'
import { checkInstallationDomainStatus } from '@/lib/console/storage'
import { DomainReselectionForm } from '@/components/domain-reselection-form'

interface PageProps {
  params: Promise<{ id: string }>
}

export default async function ReselectDomainPage({ params }: PageProps) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  try {
    await cachedInstallationAccess(id)
  } catch {
    redirect('/installations')
  }

  const installation = await cachedInstallationById(id)

  if (!installation) {
    notFound()
  }

  // If installation is already deployed, domain is locked — no reselection needed
  if (installation.deploymentReady) {
    redirect(`/installations/${id}`)
  }

  // If no subdomain yet, no reselection needed
  if (!installation.subdomain) {
    redirect(`/installations/${id}`)
  }

  // Check if the domain has expired and been claimed by another user
  const domainStatus = await checkInstallationDomainStatus(id, installation.subdomain)

  if (!domainStatus.isExpired || !domainStatus.isClaimedByOther) {
    // Domain is still valid, redirect back
    redirect(`/installations/${id}`)
  }

  return (
    <DomainReselectionForm
      installationId={id}
      installationInfo={{
        name: installation.name || 'Unnamed Installation',
        description: installation.description || null,
        oldSubdomain: installation.subdomain,
        claimedByEmail: domainStatus.claimedByEmail,
        claimedByName: domainStatus.claimedByName,
      }}
    />
  )
}
