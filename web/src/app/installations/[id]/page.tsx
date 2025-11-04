import { redirect } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { getRegistrationSession } from '@/lib/console-auth'
import { getInstallationById } from '@/lib/console/supabase-storage-service'
import { DeleteInstallationButton } from '@/components/delete-installation-button'
import { InstallationDetailsCard } from '@/components/installation-details-card'
import { InstallationsHeader } from '@/components/installations-header'
import { SuperAdminLoginCard } from '@/components/superadmin-login-card'
import { Button } from '@/components/ui/button'
import Link from 'next/link'
import { ArrowLeft, AlertTriangle } from 'lucide-react'

interface PageProps {
  params: Promise<{ id: string }>
}

export default async function InstallationSettingsPage({ params }: PageProps) {
  const { id } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    redirect('/installations/login')
  }

  const installation = await getInstallationById(id)

  if (!installation) {
    redirect('/installations')
  }

  if (installation.userId !== session.user.id) {
    redirect('/installations')
  }

  // Determine installation status
  const getStatus = () => {
    if (!installation.secretKey) {
      return {
        label: 'Not Installed',
        color: 'bg-gray-500/10 text-gray-600',
        description: 'Installation has not been deployed yet',
      }
    }
    if (!installation.subdomain) {
      return {
        label: 'Pending Domain',
        color: 'bg-yellow-500/10 text-yellow-600',
        description: 'Domain configuration is pending',
      }
    }
    if (!installation.deploymentReady) {
      return {
        label: 'Configuring',
        color: 'bg-blue-500/10 text-blue-600',
        description: 'Installation is being configured',
      }
    }
    return {
      label: 'Active',
      color: 'bg-green-500/10 text-green-600',
      description: 'Installation is active and running',
    }
  }

  const status = getStatus()
  const domain = process.env.CLOUDFLARE_DNS_DOMAIN || 'khost.dev'
  const installationUrl = installation.subdomain
    ? `https://${installation.subdomain}.${domain}`
    : null

  return (
    <div className="bg-background min-h-screen">
      <InstallationsHeader user={session.user} />

      {/* Content */}
      <main className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        {/* Back Button and Title */}
        <div className="mb-6">
          <Button asChild variant="ghost" size="sm" className="mb-4">
            <Link href="/installations">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back to Installations
            </Link>
          </Button>
          <h1 className="text-2xl font-semibold">{installation.name}</h1>
          {installation.description && (
            <p className="text-muted-foreground mt-1.5 text-sm">{installation.description}</p>
          )}
        </div>

        <div className="space-y-6">
          {/* Status & Details Card */}
          <InstallationDetailsCard
            installation={installation}
            status={status}
            domain={domain}
            installationUrl={installationUrl}
          />

          {/* Super Admin Login Card */}
          {installation.secretKey && installation.subdomain && (
            <SuperAdminLoginCard
              installationId={installation.id}
              isActive={status.label === 'Active'}
            />
          )}

          {/* Danger Zone */}
          <Card className="border-destructive">
            <CardHeader>
              <div className="flex items-center gap-2">
                <AlertTriangle className="text-destructive h-5 w-5" />
                <CardTitle className="text-destructive">Danger Zone</CardTitle>
              </div>
              <CardDescription>Irreversible actions that affect your installation</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {installation.secretKey && (
                <div className="rounded-lg border border-amber-200 bg-amber-50 p-4 dark:border-amber-900 dark:bg-amber-950">
                  <p className="mb-2 text-sm font-semibold text-amber-900 dark:text-amber-200">
                    Warning: Destructive Action
                  </p>
                  <p className="text-sm text-amber-900 dark:text-amber-200">
                    Force deleting this installation will immediately remove it from our system and
                    attempt to uninstall Kloudlite from your cluster. For a cleaner uninstallation,
                    it&apos;s recommended to uninstall from your installation&apos;s dashboard
                    settings first, then delete the record here.
                  </p>
                </div>
              )}

              <div>
                <p className="text-foreground text-sm font-semibold">Force Delete Installation</p>
                <p className="text-muted-foreground mt-1 text-sm">
                  {installation.secretKey
                    ? 'Forcefully delete this installation and uninstall Kloudlite from your cluster. This action cannot be undone.'
                    : 'Permanently delete this installation record. This action cannot be undone.'}
                </p>
              </div>
              <DeleteInstallationButton
                installationId={installation.id}
                installationName={installation.name}
                hasSecretKey={!!installation.secretKey}
                variant="button"
              />
            </CardContent>
          </Card>
        </div>
      </main>
    </div>
  )
}
