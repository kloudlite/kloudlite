import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationByKey } from '@/lib/console/storage'
import { InstallCommands } from '@/components/install-commands'

export default async function InstallPage() {
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  if (!session.installationKey) {
    redirect('/installations/new')
  }

  // Look up the installation ID from the key
  const installation = await getInstallationByKey(session.installationKey)
  const installationId = installation?.id ?? null

  return (
    <InstallCommands
      installationKey={session.installationKey}
      installationId={installationId}
    />
  )
}
