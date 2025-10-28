'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Plus, MoreHorizontal, ExternalLink, Settings } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Installation } from '@/lib/console/supabase-storage-service'

interface InstallationsListProps {
  installations: Installation[]
}

export function InstallationsList({ installations }: InstallationsListProps) {
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'installed'>('all')

  // Helper function to get installation status and next step
  const getInstallationStatus = (installation: Installation) => {
    if (!installation.secretKey) {
      return {
        status: 'Not Installed',
        statusColor: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
        nextStep: `/api/installations/${installation.id}/continue`,
        isPending: true,
      }
    }
    if (!installation.subdomain) {
      return {
        status: 'Pending Domain',
        statusColor: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
        nextStep: `/api/installations/${installation.id}/continue`,
        isPending: true,
      }
    }
    if (!installation.deploymentReady) {
      return {
        status: 'Configuring',
        statusColor: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
        nextStep: `/api/installations/${installation.id}/continue`,
        isPending: true,
      }
    }
    return {
      status: 'Active',
      statusColor: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      nextStep: null,
      isPending: false,
    }
  }

  // Apply status filter
  let filteredInstallations = installations
  if (statusFilter === 'pending') {
    filteredInstallations = filteredInstallations.filter((install) => {
      const { isPending } = getInstallationStatus(install)
      return isPending
    })
  } else if (statusFilter === 'installed') {
    filteredInstallations = filteredInstallations.filter((install) => {
      const { isPending } = getInstallationStatus(install)
      return !isPending
    })
  }

  const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'

  return (
    <div className="space-y-4">
      {/* Filter and Actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          {/* Status Filter */}
          <div className="bg-muted flex items-center gap-1 rounded-md p-1">
            <button
              onClick={() => setStatusFilter('all')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'all'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              All
            </button>
            <button
              onClick={() => setStatusFilter('pending')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'pending'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Pending
            </button>
            <button
              onClick={() => setStatusFilter('installed')}
              className={`rounded px-3 py-1 text-sm transition-colors ${
                statusFilter === 'installed'
                  ? 'bg-background shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Installed
            </button>
          </div>

          <span className="text-muted-foreground text-sm">
            {filteredInstallations.length}{' '}
            {filteredInstallations.length === 1 ? 'installation' : 'installations'}
          </span>
        </div>
        <Button asChild size="sm" className="gap-2">
          <Link href="/installations/new">
            <Plus className="h-4 w-4" />
            New Installation
          </Link>
        </Button>
      </div>

      {/* Table */}
      {filteredInstallations.length > 0 ? (
        <div className="bg-card overflow-hidden rounded-lg border">
          <table className="min-w-full">
            <thead className="bg-muted/50 border-b">
              <tr>
                <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  Name
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  Domain
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                  Status
                </th>
                <th className="text-muted-foreground px-6 py-3 text-right text-xs font-medium tracking-wider uppercase">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {filteredInstallations.map((installation) => {
                const { status, statusColor, nextStep } = getInstallationStatus(installation)
                const installationUrl = installation.subdomain
                  ? `https://${installation.subdomain}.${domain}`
                  : null

                return (
                  <tr key={installation.id} className="hover:bg-muted/50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <div className="text-sm font-semibold">{installation.name}</div>
                        {installation.description && (
                          <div className="text-muted-foreground mt-0.5 text-xs">
                            {installation.description}
                          </div>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm whitespace-nowrap">
                      {installation.subdomain ? (
                        <a
                          href={installationUrl!}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary flex items-center gap-1 font-mono hover:underline"
                        >
                          {installation.subdomain}.{domain}
                          <ExternalLink className="h-3 w-3" />
                        </a>
                      ) : (
                        <span className="text-muted-foreground text-xs">Not configured</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${statusColor}`}
                      >
                        {status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right text-sm whitespace-nowrap">
                      <div className="flex items-center justify-end gap-2">
                        {nextStep ? (
                          <Button asChild variant="default" size="sm">
                            <a href={nextStep}>Continue Setup</a>
                          </Button>
                        ) : installationUrl ? (
                          <Button asChild variant="default" size="sm">
                            <a href={installationUrl} target="_blank" rel="noopener noreferrer">
                              Open
                            </a>
                          </Button>
                        ) : (
                          <Button asChild variant="outline" size="sm">
                            <Link href={`/installations/${installation.id}`}>View Details</Link>
                          </Button>
                        )}
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem asChild>
                              <Link href={`/installations/${installation.id}`}>
                                <Settings className="mr-2 h-4 w-4" />
                                Settings
                              </Link>
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="bg-card rounded-lg border py-12 text-center">
          <p className="text-muted-foreground text-sm">
            {statusFilter === 'pending'
              ? 'No pending installations found'
              : statusFilter === 'installed'
                ? 'No installed installations found'
                : 'No installations found'}
          </p>
          {statusFilter === 'all' && (
            <Button asChild className="mt-4" size="sm">
              <Link href="/installations/new">
                <Plus className="mr-2 h-4 w-4" />
                Create Your First Installation
              </Link>
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
