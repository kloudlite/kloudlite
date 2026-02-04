import { redirect, notFound } from 'next/navigation'
import Link from 'next/link'
import { getSession } from '@/lib/get-session'
import { WorkspaceNav } from '../_components/workspace-nav'
import { WorkspaceActions } from '../_components/workspace-actions'
import { WorkspaceStatusIndicator } from '@/components/workspace-status-indicator'
import { SnapshotsSheet } from '../_components/snapshots-sheet'
import { getWorkspaceByHash } from '@/app/actions/workspace.actions'
import { ArrowLeft, Camera } from 'lucide-react'
import { Button } from '@kloudlite/ui'

interface LayoutProps {
  children: React.ReactNode
  params: Promise<{
    id: string
  }>
}

function formatTimeAgo(timestamp?: string): string {
  if (!timestamp) return 'Unknown'

  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / (1000 * 60))
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins} min${diffMins > 1 ? 's' : ''} ago`
  if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
  if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
  if (diffDays < 30) {
    const weeks = Math.floor(diffDays / 7)
    return `${weeks} week${weeks > 1 ? 's' : ''} ago`
  }
  const months = Math.floor(diffDays / 30)
  return `${months} month${months > 1 ? 's' : ''} ago`
}

export default async function WorkspaceLayout({ children, params }: LayoutProps) {
  const layoutStart = performance.now()

  const sessionStart = performance.now()
  const session = await getSession()
  console.log(`[PERF] getSession: ${(performance.now() - sessionStart).toFixed(2)}ms`)

  if (!session) {
    redirect('/auth/signin')
  }

  // id is now the workspace hash
  const { id: hash } = await params

  // Fetch workspace data using server action
  const apiStart = performance.now()
  const result = await getWorkspaceByHash(hash)
  console.log(`[PERF] getWorkspaceByHash: ${(performance.now() - apiStart).toFixed(2)}ms`)

  if (!result.success || !result.data) {
    notFound()
  }

  const { workspace, workMachineRunning } = result.data
  console.log(`[PERF] Layout total (before render): ${(performance.now() - layoutStart).toFixed(2)}ms`)

  const workspaceData = {
    hash,
    name: workspace.metadata?.name || '',
    namespace: workspace.metadata?.namespace || 'default',
    displayName: `${workspace.spec?.ownedBy || 'unknown'}/${workspace.spec?.displayName || workspace.metadata?.name}`,
    owner: workspace.spec?.ownedBy || 'unknown',
    phase: workspace.status?.phase || 'Pending',
    created: formatTimeAgo(workspace.metadata?.creationTimestamp),
  }

  return (
    <>
      {/* Back button */}
      <div className="mb-8">
        <Link
          href="/workspaces"
          className="group inline-flex items-center gap-2 text-muted-foreground hover:text-primary transition-colors duration-300 text-sm"
        >
          <ArrowLeft className="h-4 w-4 transition-transform duration-300 group-hover:-translate-x-1" />
          <span className="relative">
            Back to Workspaces
            <span className="absolute -bottom-0.5 left-0 right-0 h-0.5 bg-primary scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
          </span>
        </Link>
      </div>

      {/* Workspace Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between gap-4 mb-2">
          <h1 className="text-2xl font-semibold tracking-tight truncate">{workspaceData.displayName}</h1>
          <div className="flex-shrink-0 flex items-center gap-2">
            <SnapshotsSheet
              workspace={workspace as any}
              workMachineRunning={workMachineRunning}
              trigger={
                <Button variant="outline" size="sm">
                  <Camera className="h-4 w-4 mr-2" />
                  Snapshots
                </Button>
              }
            />
            <WorkspaceActions workspace={workspace as any} workMachineRunning={workMachineRunning} />
          </div>
        </div>
        <div className="flex items-center gap-3 text-sm text-muted-foreground">
          <span>{workspaceData.owner}</span>
          <span>•</span>
          <span>{workspaceData.created}</span>
          <span>•</span>
          <WorkspaceStatusIndicator
            namespace={workspaceData.namespace}
            workspaceName={workspaceData.name}
            initialPhase={workspaceData.phase}
          />
        </div>
      </div>

      {/* Navigation */}
      <WorkspaceNav workspaceId={hash} />

      {/* Page Content */}
      <div className="flex-1">{children}</div>
    </>
  )
}
