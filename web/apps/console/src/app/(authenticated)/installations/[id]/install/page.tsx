import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { cachedInstallationAccess, cachedInstallationById } from '@/lib/console/cached-queries'
import { getCreditAccount } from '@/lib/console/storage/credits'
import { InstallCommands } from '@/components/install-commands'
import { CompletionStatus } from '@/components/completion-status'
import { CreditTopupPrompt } from '@/components/credit-topup-prompt'

interface PageProps {
  params: Promise<{ id: string }>
}

export default async function InstallPage({ params }: PageProps) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  let orgId: string
  try {
    const context = await cachedInstallationAccess(id)
    orgId = context.orgId
  } catch {
    redirect('/installations')
  }

  const installation = await cachedInstallationById(id)

  if (!installation) {
    redirect('/installations')
  }

  if (installation.deploymentReady) {
    redirect(`/installations/${id}`)
  }

  const isKloudliteCloud = installation.cloudProvider === 'oci'
  const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'

  if (isKloudliteCloud) {
    // Check credits before showing deploy progress
    const creditAccount = await getCreditAccount(orgId)
    const balance = creditAccount?.balance ?? 0

    if (balance <= 0) {
      return <CreditTopupPrompt orgId={orgId} installationId={id} />
    }

    const installationUrl = installation.subdomain
      ? `https://${installation.subdomain}.${domain}`
      : ''

    return (
      <CompletionStatus
        installationId={installation.id}
        subdomain={installation.subdomain || ''}
        url={installationUrl}
        cloudProvider={installation.cloudProvider}
      />
    )
  }

  return (
    <InstallCommands
      installationKey={installation.installationKey}
      installationId={installation.id}
    />
  )
}
