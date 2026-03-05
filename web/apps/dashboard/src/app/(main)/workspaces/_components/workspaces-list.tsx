'use client'

import { useState, memo } from 'react'
import Link from 'next/link'
import { Activity } from 'lucide-react'
import { Badge, type BadgeProps } from '@kloudlite/ui'
import type { Workspace } from '@kloudlite/types'
import { WorkspaceRowActions } from './workspace-row-actions'
import { CreateWorkspaceSheet } from './create-workspace-sheet'
import { formatWorkspaceName } from '@kloudlite/lib'
import { useResourceWatch } from '@/lib/hooks/use-resource-watch'

interface WorkspacesListProps {
  workspaces: Workspace[]
  currentUser: string
  namespace?: string
  workMachineRunning?: boolean
  pinnedWorkspaceIds?: string[]
}

export const WorkspacesList = memo(function WorkspacesList({
  workspaces,
  currentUser,
  namespace: _namespace = 'default',
  workMachineRunning = false,
  pinnedWorkspaceIds = [],
}: WorkspacesListProps) {
  useResourceWatch('workspaces')

  const pinnedSet = new Set(pinnedWorkspaceIds)
  const [scopeFilter, setScope] = useState<'all' | 'mine'>('all')
  const [statusFilter, setStatus] = useState<'all' | 'active' | 'suspended' | 'archived'>('all')

  let filteredWorkspaces = workspaces

  // Apply scope filter (available to all users)
  if (scopeFilter === 'mine') {
    filteredWorkspaces = filteredWorkspaces.filter((ws) => ws.spec.ownedBy === currentUser)
  }

  // Apply status filter
  if (statusFilter !== 'all') {
    filteredWorkspaces = filteredWorkspaces.filter((ws) => ws.spec.status === statusFilter)
  }

  return (
    <div className="space-y-6">
      {/* Filter and Actions */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          {/* Scope Filter */}
          <div className="bg-muted flex items-center gap-1 rounded-lg p-1">
            <button
              onClick={() => setScope('all')}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                scopeFilter === 'all'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setScope('mine')}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                scopeFilter === 'mine'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Mine
            </button>
          </div>

          {/* Status Filter */}
          <div className="bg-muted flex items-center gap-1 rounded-lg p-1">
            <button
              onClick={() => setStatus('all')}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                statusFilter === 'all'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setStatus('active')}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                statusFilter === 'active'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Active
            </button>
            <button
              onClick={() => setStatus('suspended')}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                statusFilter === 'suspended'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Suspended
            </button>
            <button
              onClick={() => setStatus('archived')}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                statusFilter === 'archived'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Archived
            </button>
          </div>

          <div className="h-6 w-px bg-border" />

          <span className="text-muted-foreground text-sm font-medium">
            {filteredWorkspaces.length}{' '}
            {filteredWorkspaces.length === 1 ? 'workspace' : 'workspaces'}
          </span>
        </div>
        <CreateWorkspaceSheet workMachineRunning={workMachineRunning} />
      </div>

      {/* Table */}
      <div className="bg-card overflow-hidden rounded-xl border">
        <table className="min-w-full">
          <thead className="bg-muted/30 border-b">
            <tr>
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wider uppercase">
                Name
              </th>
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wider uppercase">
                Owner
              </th>
              <th className="text-muted-foreground w-32 px-6 py-3.5 text-left text-xs font-semibold tracking-wider uppercase">
                Status
              </th>
              <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wider uppercase">
                Environment
              </th>
              <th className="text-muted-foreground w-32 px-6 py-3.5 text-left text-xs font-semibold tracking-wider uppercase">
                Created
              </th>
              <th className="text-muted-foreground w-20 px-6 py-3.5 text-right text-xs font-semibold tracking-wider uppercase">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/50">
            {filteredWorkspaces.map((workspace) => {
              // Use runtime phase for display instead of desired spec.status
              const phase = workspace.status?.phase || 'Pending'
              const phaseVariant: BadgeProps['variant'] =
                phase === 'Running'
                  ? 'success'
                  : phase === 'Creating' || phase === 'Pending'
                    ? 'info'
                    : phase === 'Failed'
                      ? 'destructive'
                      : phase === 'Terminating'
                        ? 'warning'
                        : 'secondary'

              return (
                <tr
                  key={workspace.metadata.uid || workspace.metadata.name}
                  className="transition-colors hover:bg-muted/30"
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link
                      href={`/workspaces/${workspace.status?.hash || workspace.metadata.name}`}
                      className="hover:text-primary text-sm font-semibold"
                    >
                      {formatWorkspaceName(workspace.spec.ownedBy, workspace.metadata.name)}
                    </Link>
                  </td>
                  <td className="px-6 py-4 text-sm whitespace-nowrap">
                    {workspace.spec.ownedBy || 'unknown'}
                  </td>
                  <td className="w-32 px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      <Badge variant={phaseVariant} className="min-w-[70px] justify-center">
                        {phase}
                      </Badge>
                      {phase === 'Running' && workspace.status?.idleState === 'idle' && (
                        <Badge
                          variant="warning"
                          className="gap-1"
                          title={workspace.status?.idleSince ? `Idle since ${new Date(workspace.status.idleSince).toLocaleString()}` : 'Idle'}
                        >
                          <Activity className="h-3 w-3" />
                          Idle
                        </Badge>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-sm whitespace-nowrap">
                    {workspace.status?.connectedEnvironment ? (
                      <span className="text-foreground font-medium">
                        {workspace.status.connectedEnvironment.name}
                      </span>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </td>
                  <td className="text-muted-foreground w-32 px-6 py-4 text-sm whitespace-nowrap">
                    {workspace.metadata.creationTimestamp
                      ? new Date(workspace.metadata.creationTimestamp).toLocaleDateString('en-US', {
                          year: 'numeric',
                          month: 'short',
                          day: 'numeric',
                        })
                      : '-'}
                  </td>
                  <td className="w-20 px-6 py-4 text-right text-sm whitespace-nowrap">
                    <WorkspaceRowActions
                      workspace={workspace}
                      workMachineRunning={workMachineRunning}
                      isPinned={pinnedSet.has(`${workspace.metadata.namespace}/${workspace.metadata.name}`)}
                    />
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {filteredWorkspaces.length === 0 && (
        <div className="bg-card rounded-xl border py-16 text-center">
          <p className="text-muted-foreground text-sm">
            {scopeFilter === 'all' && statusFilter === 'active'
              ? 'No active workspaces found'
              : scopeFilter === 'all' && statusFilter !== 'all'
                ? `No ${statusFilter} workspaces found`
                : scopeFilter === 'all'
                  ? 'No workspaces found'
                  : scopeFilter === 'mine' && statusFilter !== 'all'
                    ? `You don't have any ${statusFilter} workspaces`
                    : "You don't have any workspaces yet"}
          </p>
        </div>
      )}
    </div>
  )
})
