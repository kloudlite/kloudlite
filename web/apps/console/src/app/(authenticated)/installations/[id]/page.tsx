import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { checkInstallationDomainStatus } from '@/lib/console/storage'
import { cachedInstallationAccess, cachedInstallationById } from '@/lib/console/cached-queries'
import { getInstallationStatus, hasActiveJob } from '@/lib/installation-status'
import { DeleteInstallationButton } from '@/components/delete-installation-button'
import { InstallationDetailsCard } from '@/components/installation-details-card'
import { InstallationJobProgress } from '@/components/installation-job-progress'
import { SuperAdminLoginCard } from '@/components/superadmin-login-card'
import { UninstallScriptCard } from '@/components/uninstall-script-card'
import { AlertTriangle } from 'lucide-react'

interface PageProps {
  params: Promise<{ id: string }>
}

export default async function InstallationSettingsPage({ params }: PageProps) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/login')
  }

  let userRole: string
  try {
    const { role } = await cachedInstallationAccess(id)
    userRole = role
  } catch {
    redirect('/installations')
  }

  const installation = await cachedInstallationById(id)

  if (!installation) {
    redirect('/installations')
  }

  // Check if domain has expired and been claimed by another user
  // Only check if installation has a subdomain but is not yet deployed
  if (installation.subdomain && !installation.deploymentReady) {
    const domainStatus = await checkInstallationDomainStatus(id, installation.subdomain)
    if (domainStatus.isExpired && domainStatus.isClaimedByOther) {
      // Redirect to domain re-selection page
      redirect(`/installations/${id}/domain`)
    }
  }

  // Check if installation has an active job
  const activeJob = hasActiveJob(installation)

  // Determine installation status
  const isUninstalling = installation.deployJobOperation === 'uninstall' && installation.deployJobStatus !== 'failed'

  const installationStatus = getInstallationStatus(installation)
  return (
    <div className="space-y-6">
      {/* Job Progress Banner */}
      {(activeJob || isUninstalling) && (
        <InstallationJobProgress
          installationId={installation.id}
          initialActive={true}
        />
      )}

      {/* Installation Key — inline */}
      <InstallationDetailsCard installation={installation} />

      {/* Super Admin Access */}
      {installation.secretKey && installation.subdomain && (
        <div className="border border-foreground/10 rounded-lg p-6 bg-background">
          <SuperAdminLoginCard
            installationId={installation.id}
            isActive={installationStatus.status === 'ACTIVE'}
          />
        </div>
      )}

      {/* Danger Zone — only for owner */}
      {userRole === 'owner' && !activeJob && !isUninstalling && (
        <div className="border border-destructive/20 rounded-lg p-6 bg-destructive/[0.03]">
          <div className="mb-6">
            <div className="flex items-center gap-2 mb-1">
              <AlertTriangle className="text-destructive h-5 w-5" />
              <h2 className="text-destructive text-lg font-semibold">Danger Zone</h2>
            </div>
            <p className="text-muted-foreground text-sm">Irreversible actions that affect your installation</p>
          </div>

          {installation.secretKey && installation.cloudProvider !== 'oci' && (
            <>
              <div className="border border-destructive/20 bg-destructive/[0.03] rounded-lg p-4 mb-6">
                <UninstallScriptCard
                  installationKey={installation.installationKey}
                  provider={installation.cloudProvider}
                  region={installation.cloudLocation}
                />
              </div>

              <div className="border-t border-destructive/20 my-6" />
            </>
          )}

          <div className="space-y-4">
            <div>
              <p className="text-foreground text-sm font-semibold">Delete Installation</p>
              <p className="text-muted-foreground mt-1 text-sm">
                {installation.cloudProvider === 'oci'
                  ? 'This will tear down all managed infrastructure and permanently delete this installation.'
                  : 'Permanently delete this installation record. This action cannot be undone.'}
              </p>
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
