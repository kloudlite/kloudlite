'use client'

import { useState } from 'react'
import Link from 'next/link'
import { ExternalLink } from 'lucide-react'
import type { Workspace } from '@/types/workspace'
import { WorkspaceRowActions } from './workspace-row-actions'
import { CreateWorkspaceSheet } from './create-workspace-sheet'

interface WorkspacesListProps {
  workspaces: Workspace[]
  currentUser: string
  isAdmin?: boolean
  namespace?: string
}

export function WorkspacesList({ workspaces, currentUser, isAdmin = false, namespace = 'default' }: WorkspacesListProps) {
  const [scopeFilter, setScope] = useState<'all' | 'mine'>('mine')
  const [statusFilter, setStatus] = useState<'all' | 'active' | 'suspended' | 'archived'>('all')

  let filteredWorkspaces = workspaces

  // Apply scope filter (only for admins)
  if (isAdmin && scopeFilter === 'mine') {
    filteredWorkspaces = filteredWorkspaces.filter(ws => ws.spec.owner === currentUser)
  }

  // Apply status filter
  if (statusFilter !== 'all') {
    filteredWorkspaces = filteredWorkspaces.filter(ws => ws.spec.status === statusFilter)
  }

  return (
    <div className="space-y-4">
      {/* Filter and Actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          {/* Scope Filter - Only for Admins */}
          {isAdmin && (
            <div className="flex items-center gap-1 p-1 bg-muted rounded-md">
              <button
                onClick={() => setScope('all')}
                className={`px-3 py-1 text-sm rounded transition-colors ${
                  scopeFilter === 'all'
                    ? 'bg-background shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                All
              </button>
              <button
                onClick={() => setScope('mine')}
                className={`px-3 py-1 text-sm rounded transition-colors ${
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
          <div className="flex items-center gap-1 p-1 bg-muted rounded-md">
            <button
              onClick={() => setStatus('all')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                statusFilter === 'all'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setStatus('active')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                statusFilter === 'active'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Active
            </button>
            <button
              onClick={() => setStatus('suspended')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                statusFilter === 'suspended'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Suspended
            </button>
            <button
              onClick={() => setStatus('archived')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                statusFilter === 'archived'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Archived
            </button>
          </div>

          <span className="text-sm text-muted-foreground">
            {filteredWorkspaces.length} {filteredWorkspaces.length === 1 ? 'workspace' : 'workspaces'}
          </span>
        </div>
        <CreateWorkspaceSheet namespace={namespace} user={currentUser} />
      </div>

      {/* Table */}
      <div className="bg-card rounded-lg border overflow-hidden">
        <table className="min-w-full">
          <thead className="bg-muted/50 border-b">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Owner
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Environment
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Packages
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Created
              </th>
              <th className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {filteredWorkspaces.map((workspace) => {
              const statusColor = workspace.spec.status === 'active'
                ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                : workspace.spec.status === 'suspended'
                ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
                : workspace.spec.status === 'archived'
                ? 'bg-secondary text-secondary-foreground'
                : 'bg-secondary text-secondary-foreground'

              const packageCount = workspace.spec.packages?.length || 0
              const installedCount = workspace.status?.installedPackages?.length || 0

              return (
                <tr key={workspace.metadata.uid || workspace.metadata.name} className="hover:bg-muted/50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link
                      href={`/workspaces/${workspace.metadata.namespace}/${workspace.metadata.name}`}
                      className="text-sm font-semibold hover:text-primary flex items-center gap-1"
                    >
                      {workspace.spec.displayName || workspace.metadata.name}
                      <ExternalLink className="h-3 w-3" />
                    </Link>
                    {workspace.spec.description && (
                      <p className="text-xs text-muted-foreground mt-0.5">{workspace.spec.description}</p>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {workspace.spec.owner.split('@')[0]}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${statusColor}`}>
                      {workspace.spec.status || 'active'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {workspace.status?.connectedEnvironment ? (
                      <div className="flex items-center gap-2">
                        <span className="text-foreground font-medium">
                          {workspace.status.connectedEnvironment.name}
                        </span>
                        {workspace.status.connectedEnvironment.connected ? (
                          <span className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">
                            Connected
                          </span>
                        ) : (
                          <span className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400">
                            Inactive
                          </span>
                        )}
                      </div>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
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
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {workspace.metadata.creationTimestamp
                      ? new Date(workspace.metadata.creationTimestamp).toLocaleDateString('en-US', {
                          year: 'numeric',
                          month: 'short',
                          day: 'numeric'
                        })
                      : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <WorkspaceRowActions workspace={workspace} />
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {filteredWorkspaces.length === 0 && (
        <div className="bg-card rounded-lg border text-center py-12">
          <p className="text-sm text-muted-foreground">
            {isAdmin && scopeFilter === 'all' && statusFilter === 'active'
              ? "No active workspaces found"
              : isAdmin && scopeFilter === 'all'
              ? "No workspaces found"
              : statusFilter === 'active'
              ? "You don't have any active workspaces"
              : "You don't have any workspaces yet"}
          </p>
        </div>
      )}
    </div>
  )
}