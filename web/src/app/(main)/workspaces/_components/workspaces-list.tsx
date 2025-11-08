'use client'

import { useState } from 'react'
import Link from 'next/link'
import { ExternalLink } from 'lucide-react'
import type { Workspace } from '@/types/workspace'
import { WorkspaceRowActions } from './workspace-row-actions'
import { CreateWorkspaceSheet } from './create-workspace-sheet'
import { formatResourceName } from '@/lib/utils'

interface WorkspacesListProps {
  workspaces: Workspace[]
  currentUser: string
  isAdmin?: boolean
  namespace?: string
}

export function WorkspacesList({
  workspaces,
  currentUser,
  isAdmin = false,
  namespace = 'default',
}: WorkspacesListProps) {
  const [scopeFilter, setScope] = useState<'all' | 'mine'>('all')
  const [statusFilter, setStatus] = useState<'all' | 'active' | 'suspended' | 'archived'>('all')

  let filteredWorkspaces = workspaces

  // Apply scope filter (only for admins)
  if (isAdmin && scopeFilter === 'mine') {
    filteredWorkspaces = filteredWorkspaces.filter((ws) => ws.spec.ownedBy === currentUser)
  }

  // Apply status filter
  if (statusFilter !== 'all') {
    filteredWorkspaces = filteredWorkspaces.filter((ws) => ws.spec.status === statusFilter)
  }

  return (
    <div className="space-y-4">
      {/* Filter and Actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          {/* Scope Filter - Only for Admins */}
          {isAdmin && (
            <div className="bg-muted flex items-center gap-1 rounded-md p-1">
              <button
                onClick={() => setScope('all')}
                className={`rounded px-3 py-1 text-sm transition-colors ${
                  scopeFilter === 'all'
                    ? 'bg-background shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                All
              </button>
              <button
                onClick={() => setScope('mine')}
                className={`rounded px-3 py-1 text-sm transition-colors ${
                  scopeFilter === 'mine'
                    ? 'bg-background shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                Mine
              </button>
            </div>
          )}

          {/* Status Filter */}
          <div className="bg-muted flex items-center gap-1 rounded-md p-1">
            <button
              onClick={() => setStatus('all')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'all'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setStatus('active')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'active'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Active
            </button>
            <button
              onClick={() => setStatus('suspended')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'suspended'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Suspended
            </button>
            <button
              onClick={() => setStatus('archived')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'archived'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Archived
            </button>
          </div>

          <span className="text-muted-foreground text-sm">
            {filteredWorkspaces.length}{' '}
            {filteredWorkspaces.length === 1 ? 'workspace' : 'workspaces'}
          </span>
        </div>
        <CreateWorkspaceSheet namespace={namespace} user={currentUser} />
      </div>

      {/* Table */}
      <div className="bg-card overflow-hidden rounded-lg border">
        <table className="min-w-full">
          <thead className="bg-muted/50 border-b">
            <tr>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Name
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Owner
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Status
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Environment
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Packages
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Created
              </th>
              <th className="text-muted-foreground px-6 py-3 text-right text-xs font-medium tracking-wider uppercase">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {filteredWorkspaces.map((workspace) => {
              const statusColor =
                workspace.spec.status === 'active'
                  ? 'bg-success/10 text-success'
                  : workspace.spec.status === 'suspended'
                    ? 'bg-warning/10 text-warning'
                    : workspace.spec.status === 'archived'
                      ? 'bg-secondary text-secondary-foreground'
                      : 'bg-secondary text-secondary-foreground'

              const packageCount = workspace.spec.packages?.length || 0
              const installedCount = workspace.status?.installedPackages?.length || 0

              return (
                <tr
                  key={workspace.metadata.uid || workspace.metadata.name}
                  className="hover:bg-muted/50"
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link
                      href={`/workspaces/${workspace.metadata.namespace}/${workspace.metadata.name}`}
                      className="hover:text-primary flex items-center gap-1 text-sm font-semibold"
                    >
                      {formatResourceName(workspace.metadata.name)}
                      <ExternalLink className="h-3 w-3" />
                    </Link>
                    {workspace.spec.description && (
                      <p className="text-muted-foreground mt-0.5 text-xs">
                        {workspace.spec.description}
                      </p>
                    )}
                  </td>
                  <td className="px-6 py-4 text-sm whitespace-nowrap">
                    {workspace.spec.ownedBy || 'unknown'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${statusColor}`}
                    >
                      {workspace.spec.status || 'active'}
                    </span>
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
                  <td className="px-6 py-4 text-sm whitespace-nowrap">
                    {packageCount > 0 ? (
                      <div className="text-xs">
                        <span className="text-muted-foreground">
                          {installedCount}/{packageCount} installed
                        </span>
                      </div>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </td>
                  <td className="text-muted-foreground px-6 py-4 text-sm whitespace-nowrap">
                    {workspace.metadata.creationTimestamp
                      ? new Date(workspace.metadata.creationTimestamp).toLocaleDateString('en-US', {
                          year: 'numeric',
                          month: 'short',
                          day: 'numeric',
                        })
                      : '-'}
                  </td>
                  <td className="px-6 py-4 text-right text-sm whitespace-nowrap">
                    <WorkspaceRowActions workspace={workspace} />
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {filteredWorkspaces.length === 0 && (
        <div className="bg-card rounded-lg border py-12 text-center">
          <p className="text-muted-foreground text-sm">
            {isAdmin && scopeFilter === 'all' && statusFilter === 'active'
              ? 'No active workspaces found'
              : isAdmin && scopeFilter === 'all'
                ? 'No workspaces found'
                : statusFilter === 'active'
                  ? "You don't have any active workspaces"
                  : "You don't have any workspaces yet"}
          </p>
        </div>
      )}
    </div>
  )
}
