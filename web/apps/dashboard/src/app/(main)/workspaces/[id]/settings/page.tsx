import { redirect, notFound } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { getWorkspaceByHash } from '@/app/actions/workspace.actions'
import { LocalTime } from '@/components/local-time'
import { Activity, Info, Clock, User, HardDrive, GitBranch } from 'lucide-react'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function SettingsPage({ params }: PageProps) {
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
  const phase = workspace.status?.phase || 'Pending'

  return (
    <div className="space-y-6">
      {/* General Information */}
      <div className="bg-card rounded-lg border">
        <div className="border-b p-4">
          <div className="flex items-center gap-2">
            <Info className="h-4 w-4 text-muted-foreground" />
            <h3 className="text-sm font-semibold">General Information</h3>
          </div>
        </div>
        <div className="p-4 space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Display Name</p>
              <p className="text-sm font-medium">{workspace.spec?.displayName || workspace.metadata?.name}</p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Workspace Name</p>
              <p className="text-sm font-mono">{workspace.metadata?.name}</p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Namespace</p>
              <p className="text-sm font-mono">{workspace.metadata?.namespace}</p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Hash</p>
              <p className="text-sm font-mono">{workspace.status?.hash || hash}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Owner & Time Information */}
      <div className="grid gap-6 lg:grid-cols-2">
        <div className="bg-card rounded-lg border">
          <div className="border-b p-4">
            <div className="flex items-center gap-2">
              <User className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold">Owner</h3>
            </div>
          </div>
          <div className="p-4 space-y-4">
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Owned By</p>
              <p className="text-sm font-medium">{workspace.spec?.ownedBy || 'unknown'}</p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Work Machine</p>
              <p className="text-sm font-mono">{workspace.spec?.workmachine || 'default'}</p>
            </div>
          </div>
        </div>

        <div className="bg-card rounded-lg border">
          <div className="border-b p-4">
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold">Timestamps</h3>
            </div>
          </div>
          <div className="p-4 space-y-4">
            {workspace.metadata?.creationTimestamp && (
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Created</p>
                <p className="text-sm">
                  {new Date(workspace.metadata.creationTimestamp).toLocaleDateString('en-US', {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </p>
              </div>
            )}
            {workspace.status?.startTime && (
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Last Started</p>
                <LocalTime date={workspace.status.startTime} />
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Activity Status */}
      <div className="bg-card rounded-lg border">
        <div className="border-b p-4">
          <div className="flex items-center gap-2">
            <Activity className="h-4 w-4 text-muted-foreground" />
            <h3 className="text-sm font-semibold">Activity Status</h3>
          </div>
        </div>
        <div className="p-4">
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Phase</p>
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

            {workspace.status?.idleState && (
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Idle State</p>
                <span
                  className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${
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

            {phase === 'Running' && (
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Active Connections</p>
                <p className="text-sm font-medium">{workspace.status?.activeConnections ?? 0}</p>
              </div>
            )}

            {workspace.status?.idleState === 'idle' && workspace.status?.idleSince && (
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Idle Since</p>
                <LocalTime date={workspace.status.idleSince} />
              </div>
            )}

            {workspace.status?.lastActivityTime && (
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Last Activity</p>
                <LocalTime date={workspace.status.lastActivityTime} />
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Additional Information */}
      <div className="grid gap-6 lg:grid-cols-2">

        {/* Git Repository */}
        {workspace.spec?.gitRepository && (
          <div className="bg-card rounded-lg border">
            <div className="border-b p-4">
              <div className="flex items-center gap-2">
                <GitBranch className="h-4 w-4 text-muted-foreground" />
                <h3 className="text-sm font-semibold">Git Repository</h3>
              </div>
            </div>
            <div className="p-4 space-y-4">
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Repository URL</p>
                <p className="text-sm font-mono break-all">{workspace.spec.gitRepository.url}</p>
              </div>
              {workspace.spec.gitRepository.branch && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground">Branch</p>
                  <p className="text-sm font-medium">{workspace.spec.gitRepository.branch}</p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Pod Information */}
        {(workspace.status?.podName || workspace.status?.podIP || workspace.status?.nodeName) && (
          <div className="bg-card rounded-lg border">
            <div className="border-b p-4">
              <div className="flex items-center gap-2">
                <HardDrive className="h-4 w-4 text-muted-foreground" />
                <h3 className="text-sm font-semibold">Pod Information</h3>
              </div>
            </div>
            <div className="p-4 space-y-4">
              {workspace.status?.podName && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground">Pod Name</p>
                  <p className="text-sm font-mono">{workspace.status.podName}</p>
                </div>
              )}
              {workspace.status?.podIP && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground">Pod IP</p>
                  <p className="text-sm font-mono">{workspace.status.podIP}</p>
                </div>
              )}
              {workspace.status?.nodeName && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground">Node</p>
                  <p className="text-sm font-mono">{workspace.status.nodeName}</p>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Connected Environment */}
      {workspace.status?.connectedEnvironment && (
        <div className="bg-card rounded-lg border">
          <div className="border-b p-4">
            <h3 className="text-sm font-semibold">Connected Environment</h3>
          </div>
          <div className="p-4">
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Name</p>
                <p className="text-sm font-medium">{workspace.status.connectedEnvironment.name}</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Target Namespace</p>
                <p className="text-sm font-mono">{workspace.status.connectedEnvironment.targetNamespace}</p>
              </div>
              {workspace.status.connectedEnvironment.availableServices && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground">Available Services</p>
                  <p className="text-sm">{workspace.status.connectedEnvironment.availableServices.length} services</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Last Restored Snapshot */}
      {workspace.status?.lastRestoredSnapshot && (
        <div className="bg-card rounded-lg border">
          <div className="border-b p-4">
            <h3 className="text-sm font-semibold">Last Restored Snapshot</h3>
          </div>
          <div className="p-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Snapshot Name</p>
                <p className="text-sm font-medium">{workspace.status.lastRestoredSnapshot.name}</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Restored At</p>
                <LocalTime date={workspace.status.lastRestoredSnapshot.restoredAt} />
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
