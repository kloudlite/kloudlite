import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { cachedInstallationAccess, cachedInstallationById } from '@/lib/console/cached-queries'
import { getCreditAccount } from '@/lib/console/storage/credits'
import { CompletionStatus } from '@/components/completion-status'
import { CreditTopupPrompt } from '@/components/credit-topup-prompt'

interface PageProps {
  params: Promise<{ id: string }>
}

export default async function InstallPage({ params }: PageProps) {
  const { id } = await params
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  let orgId: string
  try {
    const context = await cachedInstallationAccess(id)
    orgId = context.orgId
  } catch { redirect('/installations') }

  const installation = await cachedInstallationById(id)
  if (!installation) redirect('/installations')
  if (installation.deploymentReady) redirect(`/installations/${id}`)

  if (installation.cloudProvider === 'oci') {
    const creditAccount = await getCreditAccount(orgId)
    const balance = creditAccount?.balance ?? 0
    if (balance <= 0) return <CreditTopupPrompt installationId={id} />

    return (
      <CompletionStatus
        installationId={installation.id}
        cloudProvider={installation.cloudProvider}
      />
    )
  }

  redirect(`/installations/${id}`)
}
