import { redirect, notFound } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { getWorkspaceByHash } from '@/app/actions/workspace.actions'
import { LocalTime } from '@/components/local-time'
import { WorkspaceMetrics } from '../../_components/workspace-metrics'
import { WorkspaceConnectOptions } from '../../_components/workspace-connect-options'
import { Activity, GitBranch, Globe, ExternalLink, Layers } from 'lucide-react'
import type { PageProps } from '@/types/shared'

export default async function OverviewPage({ params }: PageProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  const { id: hash } = await params

  const result = await getWorkspaceByHash(hash)

  if (!result.success || !result.data) {
    notFound()
  }

  const { workspace } = result.data
  const namespace = workspace.metadata?.namespace || 'default'
  const name = workspace.metadata?.name || ''
  const phase = workspace.status?.phase || 'Pending'

  return (
    <div className="space-y-6">
      {/* Real-time Metrics */}
      <WorkspaceMetrics workspaceName={name} namespace={namespace} />

      {/* Connect to Workspace */}
      <WorkspaceConnectOptions
        workspaceId={`${namespace}/${name}`}
        workspace={workspace as any}
      />

      {/* Two-column grid for secondary info */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Activity Status */}
        <div className="bg-card rounded-lg border">
          <div className="border-b p-4">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold">Activity</h3>
            </div>
          </div>
          <div className="p-4">
            <div className="grid gap-4 grid-cols-2">
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Status</p>
                <span
                  className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    phase === 'Running'
                      ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                      : phase === 'Creating' || phase === 'Pending'
                        ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                        : phase === 'Failed'
                          ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                          : phase === 'Stopped'
                            ? 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
                            : 'bg-secondary text-secondary-foreground'
                  }`}
                >
                  {phase}
                </span>
              </div>

              {phase === 'Running' && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground">Connections</p>
                  <p className="text-sm font-medium">{workspace.status?.activeConnections ?? 0}</p>
                </div>
              )}

              {workspace.status?.lastActivityTime && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground">Last Activity</p>
                  <LocalTime date={workspace.status.lastActivityTime} />
                </div>
              )}

              {workspace.status?.idleState === 'idle' && workspace.status?.idleSince && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground">Idle Since</p>
                  <LocalTime date={workspace.status.idleSince} />
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Exposed Routes OR Git Repository OR Connected Environment */}
        {workspace.status?.exposedRoutes &&
          Object.keys(workspace.status.exposedRoutes).length > 0 ? (
            <div className="bg-card rounded-lg border">
              <div className="border-b p-4">
                <div className="flex items-center gap-2">
                  <Globe className="h-4 w-4 text-muted-foreground" />
                  <h3 className="text-sm font-semibold">Exposed Routes</h3>
                </div>
              </div>
              <div className="p-4 space-y-2">
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
          ) : workspace.spec?.gitRepository ? (
            <div className="bg-card rounded-lg border">
              <div className="border-b p-4">
                <div className="flex items-center gap-2">
                  <GitBranch className="h-4 w-4 text-muted-foreground" />
                  <h3 className="text-sm font-semibold">Repository</h3>
                </div>
              </div>
              <div className="p-4">
                <p className="text-sm font-mono truncate">{workspace.spec.gitRepository.url}</p>
                {workspace.spec.gitRepository.branch && (
                  <p className="text-xs text-muted-foreground mt-1">
                    Branch: {workspace.spec.gitRepository.branch}
                  </p>
                )}
              </div>
            </div>
          ) : workspace.status?.connectedEnvironment ? (
            <div className="bg-card rounded-lg border">
              <div className="border-b p-4">
                <div className="flex items-center gap-2">
                  <Layers className="h-4 w-4 text-muted-foreground" />
                  <h3 className="text-sm font-semibold">Connected Environment</h3>
                </div>
              </div>
              <div className="p-4">
                <p className="text-sm font-medium">
                  {workspace.status.connectedEnvironment.name}
                </p>
                <p className="text-xs text-muted-foreground">
                  {workspace.status.connectedEnvironment.availableServices?.length || 0} services available
                </p>
              </div>
            </div>
          ) : null}
      </div>

      {/* Show remaining sections if not shown above */}
      {workspace.status?.exposedRoutes && Object.keys(workspace.status.exposedRoutes).length > 0 && (
        <>
          {/* Git Repository - show if exposed routes took the slot */}
          {workspace.spec?.gitRepository && (
            <div className="bg-card rounded-lg border">
              <div className="border-b p-4">
                <div className="flex items-center gap-2">
                  <GitBranch className="h-4 w-4 text-muted-foreground" />
                  <h3 className="text-sm font-semibold">Repository</h3>
                </div>
              </div>
              <div className="p-4">
                <p className="text-sm font-mono truncate">{workspace.spec.gitRepository.url}</p>
                {workspace.spec.gitRepository.branch && (
                  <p className="text-xs text-muted-foreground mt-1">
                    Branch: {workspace.spec.gitRepository.branch}
                  </p>
                )}
              </div>
            </div>
          )}

          {/* Connected Environment - show if exposed routes took the slot */}
          {workspace.status?.connectedEnvironment && (
            <div className="bg-card rounded-lg border">
              <div className="border-b p-4">
                <div className="flex items-center gap-2">
                  <Layers className="h-4 w-4 text-muted-foreground" />
                  <h3 className="text-sm font-semibold">Connected Environment</h3>
                </div>
              </div>
              <div className="p-4">
                <p className="text-sm font-medium">
                  {workspace.status.connectedEnvironment.name}
                </p>
                <p className="text-xs text-muted-foreground">
                  {workspace.status.connectedEnvironment.availableServices?.length || 0} services available
                </p>
              </div>
            </div>
          )}
        </>
      )}

      {/* Show connected environment if git repo took the second slot */}
      {!workspace.status?.exposedRoutes?.length && workspace.spec?.gitRepository && workspace.status?.connectedEnvironment && (
        <div className="bg-card rounded-lg border">
          <div className="border-b p-4">
            <div className="flex items-center gap-2">
              <Layers className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold">Connected Environment</h3>
            </div>
          </div>
          <div className="p-4">
            <p className="text-sm font-medium">
              {workspace.status.connectedEnvironment.name}
            </p>
            <p className="text-xs text-muted-foreground">
              {workspace.status.connectedEnvironment.availableServices?.length || 0} services available
            </p>
          </div>
        </div>
      )}
    </div>
  )
}
