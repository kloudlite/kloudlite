'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Plus, MoreHorizontal, ExternalLink } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface Workspace {
  id: string
  name: string
  description: string
  status: 'active' | 'idle'
  lastActivity: string
  branch: string
  team: number
  environment: string
  language: string
  framework: string
}

interface WorkspacesListProps {
  workspaces: Workspace[]
  currentUser: string
  isAdmin?: boolean
}

export function WorkspacesList({ workspaces, currentUser, isAdmin = false }: WorkspacesListProps) {
  const [scopeFilter, setScope] = useState<'all' | 'mine'>('mine')
  const [statusFilter, setStatus] = useState<'all' | 'active'>('all')

  // For demo purposes, assign workspaces to different users if admin view
  const workspacesWithOwner = workspaces.map((ws, index) => ({
    ...ws,
    owner: isAdmin && scopeFilter === 'all' && index % 3 !== 0
      ? `user${index}@team.com`
      : currentUser
  }))

  let filteredWorkspaces = workspacesWithOwner

  // Apply scope filter (only for admins)
  if (isAdmin && scopeFilter === 'mine') {
    filteredWorkspaces = filteredWorkspaces.filter(ws => ws.owner === currentUser)
  }

  // Apply status filter
  if (statusFilter === 'active') {
    filteredWorkspaces = filteredWorkspaces.filter(ws => ws.status === 'active')
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
          </div>

          <span className="text-sm text-gray-500">
            {filteredWorkspaces.length} {filteredWorkspaces.length === 1 ? 'workspace' : 'workspaces'}
          </span>
        </div>
        <Button size="sm" className="gap-2">
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
                Environment
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Stack
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Last Activity
              </th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {filteredWorkspaces.map((workspace) => (
              <tr key={workspace.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <Link
                    href={`/workspaces/${workspace.id}`}
                    className="text-sm font-medium text-gray-900 hover:text-blue-600 flex items-center gap-1"
                  >
                    {workspace.name}
                    <ExternalLink className="h-3 w-3" />
                  </Link>
                  <p className="text-xs text-gray-500 mt-0.5">{workspace.description}</p>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  {workspace.owner.split('@')[0]}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                    workspace.status === 'active'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-600'
                  }`}>
                    {workspace.status}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  {workspace.environment}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  <div className="flex items-center gap-1">
                    <span>{workspace.language}</span>
                    <span className="text-gray-400">•</span>
                    <span>{workspace.framework}</span>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  {workspace.lastActivity}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem asChild>
                        <Link href={`/workspaces/${workspace.id}`}>
                          Open Workspace
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem>View Logs</DropdownMenuItem>
                      <DropdownMenuItem>Settings</DropdownMenuItem>
                      <DropdownMenuItem className="text-red-600">
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </td>
              </tr>
            ))}
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