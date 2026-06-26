import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { cachedInstallationAccess, cachedInstallationById } from '@/lib/console/cached-queries'
import { TeamManagementClient } from '@/components/console/team-client'

interface PageProps {
  params: Promise<{ id: string }>
}

export default async function TeamManagementPage({ params }: PageProps) {
  const { id } = await params
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  let userRole: string
  try {
    const { role } = await cachedInstallationAccess(id)
    userRole = role
  } catch {
    redirect('/installations')
  }

  const installation = await cachedInstallationById(id)
  if (!installation) redirect('/installations')

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-foreground text-2xl font-semibold">Team</h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Manage users who can access this installation
        </p>
      </div>

      <TeamManagementClient
        installationId={id}
        apiServerUrl={installation.apiServerUrl}
        isOwner={userRole === 'owner'}
      />
    </div>
  )
}
