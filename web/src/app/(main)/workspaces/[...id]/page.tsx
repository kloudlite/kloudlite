import { redirect, notFound } from 'next/navigation'
import { auth } from '@/lib/auth'
import { Breadcrumb } from '@/components/breadcrumb'
import { Package, CheckCircle2, XCircle, Loader2, AlertCircle, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { WorkspaceConnectOptions } from '../_components/workspace-connect-options'
import { WorkspaceActions } from '../_components/workspace-actions'
import { PackagesSheet } from '../_components/packages-sheet'
import { WorkspaceMetrics } from '../_components/workspace-metrics'
import { workspaceService } from '@/lib/services/workspace.service'

interface PageProps {
  params: {
    id: string[]
  }
}

export default async function WorkspaceDetailPage({ params }: PageProps) {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  // Parse ID array to extract namespace and name
  // Format: ["namespace", "name"] or ["name"]
  const namespace = params.id.length === 2 ? params.id[0] : 'default'
  const name = params.id.length === 2 ? params.id[1] : params.id[0]

  // Fetch workspace data
  let workspace
  let error = null

  try {
    workspace = await workspaceService.get(name, namespace)
  } catch (err) {
    console.error('Failed to fetch workspace:', err)
    error = err instanceof Error ? err.message : 'Failed to fetch workspace'
    notFound()
  }

  if (!workspace) {
    notFound()
  }

  const breadcrumbItems = [
    { label: 'Workspaces', href: '/workspaces' },
    { label: workspace.spec.displayName || workspace.metadata.name }
  ]

  const statusColor = workspace.spec.status === 'active'
    ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
    : workspace.spec.status === 'suspended'
    ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
    : 'bg-secondary text-secondary-foreground'

  const packageCount = workspace.spec.packages?.length || 0
  const installedCount = workspace.status?.installedPackages?.length || 0
  const failedCount = workspace.status?.failedPackages?.length || 0
  const pendingCount = packageCount - installedCount - failedCount

  // Create a map of installed packages for quick lookup
  const installedPackagesMap = new Map(
    workspace.status?.installedPackages?.map(pkg => [pkg.name, pkg]) || []
  )
  const failedPackagesSet = new Set(workspace.status?.failedPackages || [])

  // Determine package status for each configured package
  const packageStatuses = workspace.spec.packages?.map(pkg => {
    const isInstalled = installedPackagesMap.has(pkg.name)
    const isFailed = failedPackagesSet.has(pkg.name)
    const isPending = !isInstalled && !isFailed

    return {
      ...pkg,
      status: isInstalled ? 'installed' : isFailed ? 'failed' : 'pending',
      installedInfo: installedPackagesMap.get(pkg.name)
    }
  }) || []

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
                <h1 className="text-2xl font-semibold">{workspace.spec.displayName || workspace.metadata.name}</h1>
                {workspace.spec.description && (
                  <p className="text-sm text-muted-foreground mt-1.5">{workspace.spec.description}</p>
                )}
                <div className="mt-3 flex items-center gap-4 text-sm text-muted-foreground">
                  <span>Owner: {workspace.spec.owner.split('@')[0]}</span>
                  <span>•</span>
                  <span>Namespace: {workspace.metadata.namespace}</span>
                  <span>•</span>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColor}`}>
                    {workspace.spec.status || 'active'}
                  </span>
                  {workspace.status?.phase && (
                    <>
                      <span>•</span>
                      <span>Phase: {workspace.status.phase}</span>
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
              <div className="p-4 border-b">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div className="p-2 rounded-lg bg-primary/10">
                      <Package className="h-4 w-4 text-primary" />
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold">Packages</h3>
                      <p className="text-xs text-muted-foreground">
                        {packageCount} {packageCount === 1 ? 'package' : 'packages'} configured
                      </p>
                    </div>
                  </div>
                </div>
              </div>
              <div className="p-4 space-y-3">
                {/* Installing Packages Status */}
                {pendingCount > 0 && (
                  <div className="flex items-center gap-2 px-3 py-2 rounded-md bg-yellow-50 dark:bg-yellow-950/20 border border-yellow-200 dark:border-yellow-900/50">
                    <Loader2 className="h-4 w-4 text-yellow-600 dark:text-yellow-500 animate-spin flex-shrink-0" />
                    <span className="text-sm text-yellow-700 dark:text-yellow-400">
                      Installing {pendingCount} package{pendingCount > 1 ? 's' : ''}
                    </span>
                  </div>
                )}

                {failedCount > 0 && (
                  <div className="flex items-center gap-2 px-3 py-2 rounded-md bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-900/50">
                    <XCircle className="h-4 w-4 text-red-600 dark:text-red-500 flex-shrink-0" />
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

            {/* Activity & Info */}
            <div className="bg-card rounded-lg border p-6">
              <h3 className="text-sm font-medium mb-4">Information</h3>
              <div className="space-y-3">
                {workspace.metadata.creationTimestamp && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Created</span>
                    <span>{new Date(workspace.metadata.creationTimestamp).toLocaleDateString('en-US', {
                      year: 'numeric',
                      month: 'short',
                      day: 'numeric'
                    })}</span>
                  </div>
                )}
                {workspace.status?.lastActivity && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Last Activity</span>
                    <span>{workspace.status.lastActivity}</span>
                  </div>
                )}
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Owner</span>
                  <span>{workspace.spec.owner.split('@')[0]}</span>
                </div>
                {workspace.status?.podName && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Pod</span>
                    <span className="font-mono text-xs">{workspace.status.podName}</span>
                  </div>
                )}
                {workspace.status?.podIP && (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Pod IP</span>
                    <span className="font-mono text-xs">{workspace.status.podIP}</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}