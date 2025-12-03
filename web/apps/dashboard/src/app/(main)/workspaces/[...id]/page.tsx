import { redirect, notFound } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { Breadcrumb } from '@/components/breadcrumb'
import { Package, XCircle, Loader2, ArrowRight, Globe, ExternalLink, Activity } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { WorkspaceConnectOptions } from '../_components/workspace-connect-options'
import { WorkspaceActions } from '../_components/workspace-actions'
import { PackagesSheet } from '../_components/packages-sheet'
import { WorkspaceMetrics } from '../_components/workspace-metrics'
import { workspaceService } from '@/lib/services/workspace.service'

interface PageProps {
  params: Promise<{
    id: string[]
  }>
}

export default async function WorkspaceDetailPage({ params }: PageProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  // Parse ID array to extract namespace and name
  // Format: ["namespace", "name"] or ["name"]
  const { id } = await params
  const namespace = id.length === 2 ? id[0] : 'default'
  const name = id.length === 2 ? id[1] : id[0]

  // Fetch workspace data
  let workspace

  try {
    workspace = await workspaceService.get(name, namespace)
  } catch (err) {
    console.error('Failed to fetch workspace:', err)
    notFound()
  }

  if (!workspace) {
    notFound()
  }

  const displayName = `${workspace.spec.ownedBy || 'unknown'}/${workspace.spec.displayName || workspace.metadata.name}`

  const breadcrumbItems = [
    { label: 'Workspaces', href: '/workspaces' },
    { label: displayName },
  ]

  // Use runtime phase for status display
  const phase = workspace.status?.phase || 'Pending'
  const statusColor =
    phase === 'Running'
      ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
      : phase === 'Creating' || phase === 'Pending'
        ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
        : phase === 'Failed'
          ? 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
          : phase === 'Terminating'
            ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
            : phase === 'Stopped'
              ? 'bg-secondary text-secondary-foreground'
              : 'bg-secondary text-secondary-foreground'

  const packageCount = workspace.spec.packages?.length || 0
  const installedCount = workspace.status?.installedPackages?.length || 0
  const failedCount = workspace.status?.failedPackages?.length || 0
  const pendingCount = packageCount - installedCount - failedCount

  return (
    <>
      {/* Workspace Header with Info */}
      <div className="bg-card border-b">
        <div className="mx-auto max-w-7xl px-6">
          {/* Breadcrumb */}
          <div className="py-4">
            <Breadcrumb items={breadcrumbItems} />
          </div>

          {/* Workspace Header */}
          <div className="pb-6">
            <div className="flex items-start justify-between">
              <div>
                <h1 className="text-2xl font-semibold">{displayName}</h1>
                {workspace.spec.description && (
                  <p className="text-muted-foreground mt-1.5 text-sm">
                    {workspace.spec.description}
                  </p>
                )}
                <div className="text-muted-foreground mt-3 flex items-center gap-4 text-sm">
                  <span
                    className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${statusColor}`}
                  >
                    {phase}
                  </span>
                  {workspace.status?.message && (
                    <>
                      <span>•</span>
                      <span className="text-xs">{workspace.status.message}</span>
                    </>
                  )}
                </div>
              </div>
              <WorkspaceActions workspace={workspace} />
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="mx-auto max-w-7xl px-6 py-8">
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Connection Options - Takes 2/3 width */}
          <div className="lg:col-span-2">
            <WorkspaceConnectOptions
              workspaceId={`${workspace.metadata.namespace}/${workspace.metadata.name}`}
              workspace={workspace}
            />
          </div>

          {/* Workspace Details - Takes 1/3 width */}
          <div className="space-y-6">
            {/* Packages */}
            <div className="bg-card rounded-lg border">
              <div className="border-b p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div className="bg-primary/10 rounded-lg p-2">
                      <Package className="text-primary h-4 w-4" />
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold">Packages</h3>
                      <p className="text-muted-foreground text-xs">
                        {packageCount} {packageCount === 1 ? 'package' : 'packages'} configured
                      </p>
                    </div>
                  </div>
                </div>
              </div>
              <div className="space-y-3 p-4">
                {/* Installing Packages Status */}
                {pendingCount > 0 && (
                  <div className="flex items-center gap-2 rounded-md border border-yellow-200 bg-yellow-50 px-3 py-2 dark:border-yellow-900/50 dark:bg-yellow-950/20">
                    <Loader2 className="h-4 w-4 flex-shrink-0 animate-spin text-yellow-600 dark:text-yellow-500" />
                    <span className="text-sm text-yellow-700 dark:text-yellow-400">
                      Installing {pendingCount} package{pendingCount > 1 ? 's' : ''}
                    </span>
                  </div>
                )}

                {failedCount > 0 && (
                  <div className="flex items-center gap-2 rounded-md border border-red-200 bg-red-50 px-3 py-2 dark:border-red-900/50 dark:bg-red-950/20">
                    <XCircle className="h-4 w-4 flex-shrink-0 text-red-600 dark:text-red-500" />
                    <span className="text-sm text-red-700 dark:text-red-400">
                      {failedCount} package{failedCount > 1 ? 's' : ''} failed
                    </span>
                  </div>
                )}

                {/* Manage Packages Button */}
                <PackagesSheet
                  workspace={workspace}
                  trigger={
                    <Button variant="outline" className="w-full justify-between">
                      <span className="flex items-center gap-2">
                        <Package className="h-4 w-4" />
                        {packageCount > 0 ? 'Manage Packages' : 'Add Packages'}
                      </span>
                      <ArrowRight className="h-4 w-4" />
                    </Button>
                  }
                />
              </div>
            </div>

            {/* Real-time Metrics */}
            <WorkspaceMetrics
              workspaceName={workspace.metadata.name}
              namespace={workspace.metadata.namespace}
            />

            {/* Exposed Routes */}
            {workspace.status?.exposedRoutes && Object.keys(workspace.status.exposedRoutes).length > 0 && (
              <div className="bg-card rounded-lg border">
                <div className="border-b p-4">
                  <div className="flex items-center gap-2">
                    <div className="bg-primary/10 rounded-lg p-2">
                      <Globe className="text-primary h-4 w-4" />
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold">Exposed Routes</h3>
                      <p className="text-muted-foreground text-xs">
                        {Object.keys(workspace.status.exposedRoutes).length} HTTP{' '}
                        {Object.keys(workspace.status.exposedRoutes).length === 1 ? 'port' : 'ports'} exposed
                      </p>
                    </div>
                  </div>
                </div>
                <div className="space-y-2 p-4">
                  {Object.entries(workspace.status.exposedRoutes).map(([port, url]) => (
                    <div
                      key={port}
                      className="flex items-center justify-between rounded-md border bg-muted/30 px-3 py-2"
                    >
                      <div className="flex items-center gap-2 min-w-0">
                        <span className="text-muted-foreground text-xs font-medium">:{port}</span>
                        <span className="text-muted-foreground">→</span>
                        <a
                          href={url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-xs text-primary hover:underline truncate"
                        >
                          {url.replace('https://', '')}
                        </a>
                      </div>
                      <a
                        href={url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-muted-foreground hover:text-primary flex-shrink-0"
                      >
                        <ExternalLink className="h-3.5 w-3.5" />
                      </a>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Activity & Info */}
            <div className="bg-card rounded-lg border p-6">
              <h3 className="mb-4 text-sm font-medium">Information</h3>
              <div className="space-y-3">
                {workspace.metadata.creationTimestamp && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Created</span>
                    <span>
                      {new Date(workspace.metadata.creationTimestamp).toLocaleDateString('en-US', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric',
                      })}
                    </span>
                  </div>
                )}
                {workspace.status?.lastActivity && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Last Activity</span>
                    <span>{workspace.status.lastActivity}</span>
                  </div>
                )}
                {workspace.status?.idleState && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Activity Status</span>
                    <span
                      className={`inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium ${
                        workspace.status.idleState === 'active'
                          ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                          : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                      }`}
                    >
                      <Activity className="h-3 w-3" />
                      {workspace.status.idleState === 'active' ? 'Active' : 'Idle'}
                    </span>
                  </div>
                )}
                {workspace.status?.activeConnections !== undefined && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Active Connections</span>
                    <span>{workspace.status.activeConnections}</span>
                  </div>
                )}
                {workspace.status?.idleState === 'idle' && workspace.status?.idleSince && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Idle Since</span>
                    <span>
                      {new Date(workspace.status.idleSince).toLocaleString('en-US', {
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </span>
                  </div>
                )}
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Owner</span>
                  <span>{workspace.spec.ownedBy || 'unknown'}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
