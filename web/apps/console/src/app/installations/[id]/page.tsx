import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationById, checkInstallationDomainStatus, getMemberRole } from '@/lib/console/storage'
import { DeleteInstallationButton } from '@/components/delete-installation-button'
import { InstallationDetailsCard } from '@/components/installation-details-card'
import { SuperAdminLoginCard } from '@/components/superadmin-login-card'
import { UninstallScriptCard } from '@/components/uninstall-script-card'
import { ManagedUninstallButton } from '@/components/managed-uninstall-button'
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

  const installation = await getInstallationById(id)

  if (!installation) {
    redirect('/installations')
  }

  // Check if user has access to this installation (team member)
  // First check installation_members table, then fallback to checking if user is the owner
  let userRole = await getMemberRole(id, session.user.id)

  // If not found in members table, check if user is the installation owner (legacy support)
  if (!userRole && installation.userId === session.user.id) {
    userRole = 'owner'
  }

  if (!userRole) {
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

  // Determine installation status
  const getStatus = () => {
    if (!installation.secretKey) {
      return {
        label: 'NOT INSTALLED',
        color: 'bg-foreground/[0.06] text-foreground border border-foreground/10',
        description: 'Installation has not been deployed yet',
      }
    }
    if (!installation.subdomain) {
      return {
        label: 'PENDING',
        color: 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border border-yellow-500/20',
        description: 'Domain configuration is pending',
      }
    }
    if (!installation.deploymentReady) {
      return {
        label: 'CONFIGURING',
        color: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/20',
        description: 'Installation is being configured',
      }
    }
    return {
      label: 'ACTIVE',
      color: 'bg-green-500/10 text-green-700 dark:text-green-400 border border-green-500/20',
      description: 'Installation is active and running',
    }
  }

  const status = getStatus()
  const domain = process.env.CLOUDFLARE_DNS_DOMAIN || 'khost.dev'
  const installationUrl = installation.subdomain
    ? `https://${installation.subdomain}.${domain}`
    : null

  return (
    <div className="space-y-6">
      {/* Status & Details Card */}
      <div className="border border-foreground/10 rounded-lg p-6 bg-background">
        <InstallationDetailsCard
          installation={installation}
          status={status}
          domain={domain}
          installationUrl={installationUrl}
        />
      </div>

      {/* Super Admin Login Card */}
      {installation.secretKey && installation.subdomain && (
        <div className="border border-foreground/10 rounded-lg p-6 bg-background">
          <SuperAdminLoginCard
            installationId={installation.id}
            isActive={status.label === 'ACTIVE'}
          />
        </div>
      )}

      {/* Danger Zone - Only for Owner */}
      {userRole === 'owner' && (
        <div className="border border-red-500/20 rounded-lg p-6 bg-red-500/5">
          <div className="mb-6">
            <div className="flex items-center gap-2 mb-1">
              <AlertTriangle className="text-red-600 dark:text-red-400 h-5 w-5" />
              <h2 className="text-red-600 dark:text-red-400 text-xl font-semibold">Danger Zone</h2>
            </div>
            <p className="text-muted-foreground text-sm">Irreversible actions that affect your installation</p>
          </div>

          {installation.secretKey && installation.cloudProvider === 'oci' && (
            <>
              <div className="space-y-4 mb-6">
                <div>
                  <p className="text-foreground text-sm font-semibold">Uninstall Kloudlite Cloud</p>
                  <p className="text-muted-foreground mt-1 text-sm">
                    This will tear down all managed infrastructure and delete this installation.
                  </p>
                </div>
                <ManagedUninstallButton
                  installationId={installation.id}
                  installationName={installation.name}
                />
              </div>

              <div className="border-t border-red-500/20 my-6" />
            </>
          )}

          {installation.secretKey && installation.cloudProvider !== 'oci' && (
            <>
              <div className="border border-red-500/20 bg-red-500/5 rounded-lg p-4 mb-6">
                <UninstallScriptCard
                  installationKey={installation.installationKey}
                  provider={installation.cloudProvider}
                  region={installation.cloudLocation}
                />
              </div>

              <div className="border-t border-red-500/20 my-6" />

              <div className="border border-amber-500/20 bg-amber-500/10 rounded-lg p-4 mb-6">
                <p className="mb-2 text-sm font-semibold text-amber-900 dark:text-amber-200">
                  Warning: Destructive Action
                </p>
                <p className="text-sm text-amber-900 dark:text-amber-200">
                  Force deleting this installation will immediately remove it from our system.
                  For a cleaner uninstallation, run the uninstall script above first,
                  then delete the record here.
                </p>
              </div>
            </>
          )}

          <div className="space-y-4">
            <div>
              <p className="text-foreground text-sm font-semibold">Force Delete Installation</p>
              <p className="text-muted-foreground mt-1 text-sm">
                {installation.secretKey
                  ? 'Forcefully delete this installation record. Run the uninstall script above first if you want to clean up cloud resources.'
                  : 'Permanently delete this installation record. This action cannot be undone.'}
              </p>
            </div>
            <DeleteInstallationButton
              installationId={installation.id}
              installationName={installation.name}
              hasSecretKey={!!installation.secretKey}
              variant="button"
            />
          </div>
        </div>
      )}
    </div>
  )
}
