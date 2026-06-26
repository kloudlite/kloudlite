import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { cachedInstallationAccess, cachedInstallationById } from '@/lib/console/cached-queries'
import { hasActiveJob } from '@/lib/installation-status'
import { DeleteInstallationButton } from '@/components/delete-installation-button'
import { InstallationDetailsCard } from '@/components/installation-details-card'
import { InstallationJobProgress } from '@/components/installation-job-progress'
import { AlertTriangle } from 'lucide-react'

interface PageProps {
  params: Promise<{ id: string }>
}

export default async function InstallationSettingsPage({ params }: PageProps) {
  const { id } = await params
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  let userRole: string
  try {
    const { role } = await cachedInstallationAccess(id)
    userRole = role
  } catch { redirect('/installations') }

  const installation = await cachedInstallationById(id)
  if (!installation) redirect('/installations')

  const activeJob = hasActiveJob(installation)
  const isUninstalling = installation.deployJobOperation === 'uninstall' && installation.deployJobStatus !== 'failed'

  return (
    <div className="space-y-6">
      {(activeJob || isUninstalling) && (
        <InstallationJobProgress installationId={installation.id} initialActive={true} />
      )}

      <InstallationDetailsCard installation={installation} />

      {/* Team Management */}
      <div className="border border-foreground/10 rounded-lg p-6 bg-background">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-foreground text-lg font-semibold">Team</h2>
            <p className="text-muted-foreground text-sm mt-1">Manage user access to this installation</p>
          </div>
          <a
            href={`/installations/${installation.id}/team`}
            className="inline-flex items-center gap-1.5 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            Manage Team
          </a>
        </div>
      </div>

      {/* Danger Zone */}
      {userRole === 'owner' && !activeJob && !isUninstalling && (
        <div className="border border-destructive/20 rounded-lg p-6 bg-destructive/[0.03]">
          <div className="mb-6">
            <div className="flex items-center gap-2 mb-1">
              <AlertTriangle className="text-destructive h-5 w-5" />
              <h2 className="text-destructive text-lg font-semibold">Danger Zone</h2>
            </div>
            <p className="text-muted-foreground text-sm">Irreversible actions that affect your installation</p>
          </div>
          <div className="space-y-4">
            <div>
              <p className="text-foreground text-sm font-semibold">Delete Installation</p>
              <p className="text-muted-foreground mt-1 text-sm">Permanently delete this installation record. This action cannot be undone.</p>
            </div>
            <DeleteInstallationButton
              installationId={installation.id}
              installationName={installation.name}
              hasSecretKey={!!installation.secretKey}
              cloudProvider={installation.cloudProvider}
              variant="button"
            />
          </div>
        </div>
      )}
    </div>
  )
}
