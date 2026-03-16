import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationByKey } from '@/lib/console/storage'
import { CompletionStatus } from '@/components/completion-status'

export default async function CompletePage() {
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  if (!session.installationKey) {
    redirect('/installations/new-byoc')
  }

  const installation = await getInstallationByKey(session.installationKey)

  if (!installation) {
    redirect('/installations/new-byoc')
  }

  const isValidSubdomain = installation.subdomain &&
    installation.subdomain !== '0.0.0.0' &&
    !installation.subdomain.includes('0.0.0.0')

  if (!isValidSubdomain) {
    // No valid subdomain yet, can't show completion page
    redirect('/installations/new/install')
  }

  const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'

  return (
    <CompletionStatus
      subdomain={installation.subdomain!}
      url={`https://${installation.subdomain}.${domain}`}
      installationId={installation.id}
      cloudProvider={installation.cloudProvider}
    />
  )
}
