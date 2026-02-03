import { redirect, notFound } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { Globe, ExternalLink } from 'lucide-react'
import { WorkspaceConnectOptions } from '../../_components/workspace-connect-options'
import { WorkspaceMetrics } from '../../_components/workspace-metrics'
import { CodeAnalysisCard } from '../../_components/code-analysis-card'
import { getWorkspaceByHash } from '@/app/actions/workspace.actions'

interface PageProps {
  params: Promise<{
    id: string
  }>
}

export default async function ConnectPage({ params }: PageProps) {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  const { id: hash } = await params

  const result = await getWorkspaceByHash(hash)

  if (!result.success || !result.data) {
    notFound()
  }

  const { workspace, workMachineRunning } = result.data
  const namespace = workspace.metadata?.namespace || 'default'
  const name = workspace.metadata?.name || ''

  return (
    <div className="space-y-6">
      {/* Connection Options */}
      <WorkspaceConnectOptions
        workspaceId={`${namespace}/${name}`}
        workspace={workspace as any}
      />

      {/* Code Analysis - Only visible when WorkMachine is running */}
      {workMachineRunning && (
        <CodeAnalysisCard
          workspaceName={name}
          namespace={namespace}
        />
      )}

      {/* Two column layout for metrics and routes */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Real-time Metrics */}
        <WorkspaceMetrics
          workspaceName={name}
          namespace={namespace}
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
      </div>
    </div>
  )
}
