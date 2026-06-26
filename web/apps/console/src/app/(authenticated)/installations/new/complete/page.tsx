import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationByKey } from '@/lib/console/storage'
import { CompletionStatus } from '@/components/completion-status'

export default async function CompletePage() {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')
  if (!session.installationKey) redirect('/installations/new')

  const installation = await getInstallationByKey(session.installationKey)
  if (!installation) redirect('/installations/new')

  return (
    <CompletionStatus
      installationId={installation.id}
      cloudProvider={installation.cloudProvider}
    />
  )
}
