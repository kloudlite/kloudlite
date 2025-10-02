'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Plus, MoreHorizontal, ExternalLink, Loader2 } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Workspace } from '@/types/workspace'
import { workspaceService } from '@/services/workspace-service'

interface WorkspacesListProps {
  workspaces: Workspace[]
  currentUser: string
  isAdmin?: boolean
  namespace?: string
}

export function WorkspacesList({ workspaces, currentUser, isAdmin = false, namespace = 'default' }: WorkspacesListProps) {
  const router = useRouter()
  const [scopeFilter, setScope] = useState<'all' | 'mine'>('mine')
  const [statusFilter, setStatus] = useState<'all' | 'active' | 'suspended' | 'archived'>('all')
  const [deletingWorkspace, setDeletingWorkspace] = useState<string | null>(null)

  let filteredWorkspaces = workspaces

  // Apply scope filter (only for admins)
  if (isAdmin && scopeFilter === 'mine') {
    filteredWorkspaces = filteredWorkspaces.filter(ws => ws.spec.owner === currentUser)
  }

  // Apply status filter
  if (statusFilter !== 'all') {
    filteredWorkspaces = filteredWorkspaces.filter(ws => ws.spec.status === statusFilter)
  }

  const handleDelete = async (workspace: Workspace) => {
    if (!confirm(`Are you sure you want to delete workspace "${workspace.metadata.name}"?`)) {
      return
    }

    setDeletingWorkspace(workspace.metadata.name)
    try {
      await workspaceService.delete(workspace.metadata.name, workspace.metadata.namespace)
      router.refresh()
    } catch (error) {
      console.error('Failed to delete workspace:', error)
      alert('Failed to delete workspace')
    } finally {
      setDeletingWorkspace(null)
    }
  }

  const handleWorkspaceAction = async (workspace: Workspace, action: 'suspend' | 'activate' | 'archive') => {
    try {
      if (action === 'suspend') {
        await workspaceService.suspend(workspace.metadata.name, workspace.metadata.namespace)
      } else if (action === 'activate') {
        await workspaceService.activate(workspace.metadata.name, workspace.metadata.namespace)
      } else if (action === 'archive') {
        await workspaceService.archive(workspace.metadata.name, workspace.metadata.namespace)
      }
      router.refresh()
    } catch (error) {
      console.error(`Failed to ${action} workspace:`, error)
      alert(`Failed to ${action} workspace`)
    }
  }

  return (
    <div className="space-y-4">
      {/* Filter and Actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          {/* Scope Filter - Only for Admins */}
          {isAdmin && (
            <div className="flex items-center gap-1 p-1 bg-gray-100 rounded-md">
              <button
                onClick={() => setScope('all')}
                className={`px-3 py-1 text-sm rounded transition-colors ${
                  scopeFilter === 'all'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                All
              </button>
              <button
                onClick={() => setScope('mine')}
                className={`px-3 py-1 text-sm rounded transition-colors ${
                  scopeFilter === 'mine'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Mine
              </button>
            </div>
          )}

          {/* Status Filter */}
          <div className="flex items-center gap-1 p-1 bg-gray-100 rounded-md">
            <button
              onClick={() => setStatus('all')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                statusFilter === 'all'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setStatus('active')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                statusFilter === 'active'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Active
            </button>
            <button
              onClick={() => setStatus('suspended')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                statusFilter === 'suspended'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Suspended
            </button>
            <button
              onClick={() => setStatus('archived')}
              className={`px-3 py-1 text-sm rounded transition-colors ${
                statusFilter === 'archived'
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Archived
            </button>
          </div>

          <span className="text-sm text-gray-500">
            {filteredWorkspaces.length} {filteredWorkspaces.length === 1 ? 'workspace' : 'workspaces'}
          </span>
        </div>
        <Button
          size="sm"
          className="gap-2"
          onClick={() => router.push('/workspaces/new')}
        >
          <Plus className="h-4 w-4" />
          New Workspace
        </Button>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <table className="min-w-full">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Owner
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Work Machine
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Resources
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Created
              </th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {filteredWorkspaces.map((workspace) => {
              const isDeleting = deletingWorkspace === workspace.metadata.name
              const statusColor = workspace.spec.status === 'active'
                ? 'bg-green-100 text-green-800'
                : workspace.spec.status === 'suspended'
                ? 'bg-yellow-100 text-yellow-800'
                : workspace.spec.status === 'archived'
                ? 'bg-gray-100 text-gray-600'
                : 'bg-gray-100 text-gray-600'

              return (
                <tr key={workspace.metadata.uid || workspace.metadata.name} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link
                      href={`/workspaces/${workspace.metadata.namespace}/${workspace.metadata.name}`}
                      className="text-sm font-semibold text-gray-900 hover:text-blue-600 flex items-center gap-1"
                    >
                      {workspace.spec.displayName || workspace.metadata.name}
                      <ExternalLink className="h-3 w-3" />
                    </Link>
                    {workspace.spec.description && (
                      <p className="text-xs text-gray-500 mt-0.5">{workspace.spec.description}</p>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                    {workspace.spec.owner.split('@')[0]}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${statusColor}`}>
                      {workspace.spec.status || 'active'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                    {workspace.spec.workMachineRef?.name || '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                    <div className="text-xs">
                      {workspace.spec.resourceQuota ? (
                        <>
                          <div>CPU: {workspace.spec.resourceQuota.cpu || '-'}</div>
                          <div>Mem: {workspace.spec.resourceQuota.memory || '-'}</div>
                        </>
                      ) : (
                        '-'
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                    {workspace.metadata.creationTimestamp
                      ? new Date(workspace.metadata.creationTimestamp).toLocaleDateString()
                      : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 w-8 p-0"
                          disabled={isDeleting}
                        >
                          {isDeleting ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <MoreHorizontal className="h-4 w-4" />
                          )}
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem asChild>
                          <Link href={`/workspaces/${workspace.metadata.namespace}/${workspace.metadata.name}`}>
                            Open Workspace
                          </Link>
                        </DropdownMenuItem>
                        {workspace.spec.status !== 'suspended' && (
                          <DropdownMenuItem onClick={() => handleWorkspaceAction(workspace, 'suspend')}>
                            Suspend
                          </DropdownMenuItem>
                        )}
                        {workspace.spec.status === 'suspended' && (
                          <DropdownMenuItem onClick={() => handleWorkspaceAction(workspace, 'activate')}>
                            Activate
                          </DropdownMenuItem>
                        )}
                        {workspace.spec.status !== 'archived' && (
                          <DropdownMenuItem onClick={() => handleWorkspaceAction(workspace, 'archive')}>
                            Archive
                          </DropdownMenuItem>
                        )}
                        <DropdownMenuItem>Settings</DropdownMenuItem>
                        <DropdownMenuItem
                          className="text-red-600"
                          onClick={() => handleDelete(workspace)}
                        >
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {filteredWorkspaces.length === 0 && (
        <div className="bg-white rounded-lg border border-gray-200 text-center py-12">
          <p className="text-sm text-gray-500">
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