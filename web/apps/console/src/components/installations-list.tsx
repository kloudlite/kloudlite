'use client'

import { useState, useRef, useEffect } from 'react'
import Link from 'next/link'
import { Button } from '@kloudlite/ui'
import { Plus, MoreHorizontal, ExternalLink, Settings } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import { cn } from '@kloudlite/lib'
import type { Installation } from '@/lib/console/supabase-storage-service'
import { NewInstallationDialog } from './new-installation-dialog'

interface InstallationsListProps {
  installations: Installation[]
}

export function InstallationsList({ installations }: InstallationsListProps) {
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'installed'>('all')
  const [underlineStyle, setUnderlineStyle] = useState({ left: 0, width: 0 })
  const [showNewDialog, setShowNewDialog] = useState(false)

  const allRef = useRef<HTMLButtonElement>(null)
  const pendingRef = useRef<HTMLButtonElement>(null)
  const installedRef = useRef<HTMLButtonElement>(null)

  // Helper function to get installation status and next step
  const getInstallationStatus = (installation: Installation) => {
    if (!installation.secretKey) {
      return {
        status: 'NOT INSTALLED',
        statusColor: 'bg-foreground/[0.06] text-foreground border border-foreground/10',
        nextStep: `/api/installations/${installation.id}/continue`,
        isPending: true,
      }
    }
    if (!installation.subdomain) {
      return {
        status: 'PENDING',
        statusColor: 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border border-yellow-500/20',
        nextStep: `/api/installations/${installation.id}/continue`,
        isPending: true,
      }
    }
    if (!installation.deploymentReady) {
      return {
        status: 'CONFIGURING',
        statusColor: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/20',
        nextStep: `/api/installations/${installation.id}/continue`,
        isPending: true,
      }
    }
    return {
      status: 'ACTIVE',
      statusColor: 'bg-green-500/10 text-green-700 dark:text-green-400 border border-green-500/20',
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

  // Calculate underline position and width
  useEffect(() => {
    const updateUnderline = () => {
      let activeRef: HTMLButtonElement | null = null
      if (statusFilter === 'all') activeRef = allRef.current
      else if (statusFilter === 'pending') activeRef = pendingRef.current
      else if (statusFilter === 'installed') activeRef = installedRef.current

      if (activeRef) {
        const { offsetLeft, offsetWidth } = activeRef
        const underlineWidth = offsetWidth * 0.6
        const underlineLeft = offsetLeft + (offsetWidth - underlineWidth) / 2
        setUnderlineStyle({ left: underlineLeft, width: underlineWidth })
      }
    }

    updateUnderline()
    window.addEventListener('resize', updateUnderline)
    return () => window.removeEventListener('resize', updateUnderline)
  }, [statusFilter])

  const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'

  return (
    <div className="space-y-6">
      {/* Filter and Actions */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 w-full sm:w-auto">
          {/* Status Filter */}
          <div className="inline-flex gap-1 relative">
            <button
              ref={allRef}
              onClick={() => setStatusFilter('all')}
              className={cn(
                'relative px-6 py-2.5 text-base font-medium transition-all duration-200 cursor-pointer',
                'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
                statusFilter === 'all'
                  ? 'text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              All
            </button>
            <button
              ref={pendingRef}
              onClick={() => setStatusFilter('pending')}
              className={cn(
                'relative px-6 py-2.5 text-base font-medium transition-all duration-200 cursor-pointer',
                'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
                statusFilter === 'pending'
                  ? 'text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              Pending
            </button>
            <button
              ref={installedRef}
              onClick={() => setStatusFilter('installed')}
              className={cn(
                'relative px-6 py-2.5 text-base font-medium transition-all duration-200 cursor-pointer',
                'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
                statusFilter === 'installed'
                  ? 'text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              Installed
            </button>

            {/* Animated underline */}
            <div
              className="absolute bottom-1 h-[2px] bg-primary transition-all duration-300 ease-out"
              style={{
                left: `${underlineStyle.left}px`,
                width: `${underlineStyle.width}px`,
              }}
            />
          </div>

          <span className="text-muted-foreground text-sm font-medium">
            {filteredInstallations.length}{' '}
            {filteredInstallations.length === 1 ? 'installation' : 'installations'}
          </span>
        </div>
        <Button size="lg" className="w-full sm:w-auto" onClick={() => setShowNewDialog(true)}>
          <Plus className="h-4 w-4" />
          New Installation
        </Button>
      </div>

      <NewInstallationDialog open={showNewDialog} onOpenChange={setShowNewDialog} />

      {/* Table */}
      {filteredInstallations.length > 0 ? (
        <div className="overflow-hidden border border-foreground/10 shadow-sm">
          <div className="overflow-x-auto">
            <table className="min-w-full">
              <thead>
                <tr className="border-b border-foreground/10 bg-muted/50">
                  <th className="text-muted-foreground px-6 py-3 text-left text-[11px] font-medium uppercase tracking-wider">
                    Name
                  </th>
                  <th className="text-muted-foreground px-6 py-3 text-left text-[11px] font-medium uppercase tracking-wider hidden md:table-cell">
                    Domain
                  </th>
                  <th className="text-muted-foreground px-6 py-3 text-left text-[11px] font-medium uppercase tracking-wider">
                    Status
                  </th>
                  <th className="text-muted-foreground px-6 py-3 text-right text-[11px] font-medium uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-background divide-y divide-foreground/10">
                {filteredInstallations.map((installation) => {
                  const { status, statusColor, nextStep } = getInstallationStatus(installation)
                  // Validate subdomain before constructing URL
                  const isValidSubdomain =
                    installation.subdomain &&
                    installation.subdomain !== '0.0.0.0' &&
                    !installation.subdomain.includes('0.0.0.0')
                  const installationUrl = isValidSubdomain
                    ? `https://${installation.subdomain}.${domain}`
                    : null

                  return (
                    <tr key={installation.id} className="group hover:bg-foreground/[0.015] transition-colors relative">
                      <td className="px-6 py-4 relative">
                        {/* Left accent bar */}
                        <div className="absolute left-0 top-0 w-[2px] h-full bg-primary/80 scale-y-0 group-hover:scale-y-100 transition-transform duration-200 origin-top" />

                        <div className="relative z-10">
                          <div className="text-sm font-medium text-foreground group-hover:text-foreground transition-colors">
                            {installation.name}
                          </div>
                          {installation.description && (
                            <div className="text-muted-foreground/60 mt-0.5 text-xs line-clamp-1 leading-relaxed">
                              {installation.description}
                            </div>
                          )}
                          {/* Show domain on mobile */}
                          <div className="md:hidden mt-1">
                            {installationUrl ? (
                              <a
                                href={installationUrl}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-primary/80 hover:text-primary flex items-center gap-1 text-[11px] font-mono hover:underline transition-colors"
                              >
                                {installation.subdomain}.{domain}
                                <ExternalLink className="h-3 w-3" />
                              </a>
                            ) : (
                              <span className="text-muted-foreground/50 text-xs">Not configured</span>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 hidden md:table-cell">
                        {installationUrl ? (
                          <a
                            href={installationUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-primary/80 hover:text-primary inline-flex items-center gap-1.5 text-[13px] font-mono hover:underline transition-colors"
                          >
                            {installation.subdomain}.{domain}
                            <ExternalLink className="h-3 w-3" />
                          </a>
                        ) : (
                          <span className="text-muted-foreground/50 text-sm">Not configured</span>
                        )}
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className={`inline-flex px-2 py-0.5 text-[11px] font-medium rounded-md ${statusColor}`}
                        >
                          {status}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <div className="flex items-center justify-end gap-2">
                          {nextStep ? (
                            <Button asChild variant="default" size="sm">
                              <a href={nextStep}>Continue</a>
                            </Button>
                          ) : installationUrl ? (
                            <Button asChild variant="default" size="sm">
                              <a href={installationUrl} target="_blank" rel="noopener noreferrer">
                                Open
                              </a>
                            </Button>
                          ) : (
                            <Button asChild variant="outline" size="sm" className="hidden sm:inline-flex">
                              <Link href={`/installations/${installation.id}`}>Details</Link>
                            </Button>
                          )}
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon">
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
        </div>
      ) : (
        <div className="border border-foreground/10 shadow-sm py-16 text-center bg-muted/20">
          <div className="mx-auto max-w-md px-4">
            <div className="mx-auto w-12 h-12 bg-muted border border-foreground/10 flex items-center justify-center mb-4">
              <Plus className="h-5 w-5 text-muted-foreground" />
            </div>
            <h3 className="text-foreground text-base font-semibold mb-1">
              {statusFilter === 'pending'
                ? 'No pending installations'
                : statusFilter === 'installed'
                  ? 'No active installations'
                  : 'No installations'}
            </h3>
            <p className="text-muted-foreground text-sm mb-6 leading-relaxed">
              {statusFilter === 'all'
                ? 'Create your first installation to get started'
                : 'Adjust filters to see more results'}
            </p>
            {statusFilter === 'all' && (
              <Button size="lg" onClick={() => setShowNewDialog(true)}>
                <Plus className="h-4 w-4" />
                New Installation
              </Button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
