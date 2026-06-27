import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { cachedInstallationAccess, cachedInstallationById } from '@/lib/console/cached-queries'
import { CompletionStatus } from '@/components/completion-status'

interface PageProps {
  params: Promise<{ id: string }>
}

export default async function InstallPage({ params }: PageProps) {
  const { id } = await params
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  try {
    await cachedInstallationAccess(id)
  } catch { redirect('/installations') }

  const installation = await cachedInstallationById(id)
  if (!installation) redirect('/installations')
  if (installation.deploymentReady) redirect(`/installations/${id}`)

  return (
    <CompletionStatus
      installationId={installation.id}
      cloudProvider={installation.cloudProvider}
    />
  )
}
