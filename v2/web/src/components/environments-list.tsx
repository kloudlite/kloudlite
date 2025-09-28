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

interface Environment {
  id: string
  name: string
  owner: string
  status: 'active' | 'inactive'
  created: string
  services: number
  configs: number
  secrets: number
  workspaces: string[]
  lastDeployed: string
}

interface EnvironmentsListProps {
  environments: Environment[]
  currentUser: string
}

export function EnvironmentsList({ environments, currentUser }: EnvironmentsListProps) {
  const [scopeFilter, setScope] = useState<'all' | 'mine'>('all')
  const [statusFilter, setStatus] = useState<'all' | 'active'>('all')

  let filteredEnvironments = environments

  // Apply scope filter
  if (scopeFilter === 'mine') {
    filteredEnvironments = filteredEnvironments.filter(env => env.owner === currentUser)
  }

  // Apply status filter
  if (statusFilter === 'active') {
    filteredEnvironments = filteredEnvironments.filter(env => env.status === 'active')
  }

  const counts = {
    all: environments.length,
    mine: environments.filter(env => env.owner === currentUser).length,
    active: environments.filter(env => env.status === 'active').length,
    allActive: environments.filter(env => env.status === 'active').length,
    mineActive: environments.filter(env => env.owner === currentUser && env.status === 'active').length,
  }

  return (
    <div className="space-y-4">
      {/* Filter and Actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          {/* Scope Filter */}
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
            {filteredEnvironments.length} {filteredEnvironments.length === 1 ? 'environment' : 'environments'}
          </span>
        </div>
        <Button size="sm" className="gap-2">
          <Plus className="h-4 w-4" />
          New Environment
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
                Resources
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Last Deployed
              </th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {filteredEnvironments.map((env) => (
              <tr key={env.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <Link
                    href={`/environments/${env.id}`}
                    className="text-sm font-medium text-gray-900 hover:text-blue-600 flex items-center gap-1"
                  >
                    {env.name}
                    <ExternalLink className="h-3 w-3" />
                  </Link>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  {env.owner.split('@')[0]}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                    env.status === 'active'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-600'
                  }`}>
                    {env.status}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  <div className="flex items-center gap-4">
                    <span>{env.services} services</span>
                    <span className="text-gray-400">•</span>
                    <span>{env.configs} configs</span>
                    <span className="text-gray-400">•</span>
                    <span>{env.secrets} secrets</span>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  {env.lastDeployed}
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
                        <Link href={`/environments/${env.id}`}>
                          View Details
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem>Clone Environment</DropdownMenuItem>
                      <DropdownMenuItem>Export Config</DropdownMenuItem>
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

      {filteredEnvironments.length === 0 && (
        <div className="bg-white rounded-lg border border-gray-200 text-center py-12">
          <p className="text-sm text-gray-500">
            {scopeFilter === 'mine' && statusFilter === 'active'
              ? "You don't have any active environments"
              : scopeFilter === 'mine'
              ? "You don't have any environments yet"
              : statusFilter === 'active'
              ? "No active environments found"
              : "No environments found"}
          </p>
        </div>
      )}
    </div>
  )
}