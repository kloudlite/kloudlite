'use client'

import { useState, useRef, useEffect } from 'react'
import { useRouter } from 'next/navigation'
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

interface InstallationsListProps {
  installations: Installation[]
}

export function InstallationsList({ installations }: InstallationsListProps) {
  const router = useRouter()
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'installed'>('all')
  const [underlineStyle, setUnderlineStyle] = useState({ left: 0, width: 0 })

  const allRef = useRef<HTMLButtonElement>(null)
  const pendingRef = useRef<HTMLButtonElement>(null)
  const installedRef = useRef<HTMLButtonElement>(null)

  // Helper function to get installation status
  const getInstallationStatus = (installation: Installation) => {
    if (!installation.secretKey) {
      return {
        status: 'NOT INSTALLED',
        statusColor: 'bg-foreground/[0.06] text-foreground border border-foreground/10',
        isPending: true,
      }
    }
    if (!installation.subdomain) {
      return {
        status: 'PENDING',
        statusColor: 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border border-yellow-500/20',
        isPending: true,
      }
    }
    if (!installation.deploymentReady) {
      return {
        status: 'CONFIGURING',
        statusColor: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/20',
        isPending: true,
      }
    }
    return {
      status: 'ACTIVE',
      statusColor: 'bg-green-500/10 text-green-700 dark:text-green-400 border border-green-500/20',
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

  // Update underline position
  useEffect(() => {
    const updatePosition = () => {
      let activeRef: HTMLButtonElement | null = null
      if (statusFilter === 'all') activeRef = allRef.current
      else if (statusFilter === 'pending') activeRef = pendingRef.current
      else if (statusFilter === 'installed') activeRef = installedRef.current

      if (activeRef) {
        const fullWidth = activeRef.offsetWidth
        const underlineWidth = fullWidth * 0.6 // 60% of button width
        const leftOffset = activeRef.offsetLeft + (fullWidth - underlineWidth) / 2

        setUnderlineStyle({
          left: leftOffset,
          width: underlineWidth
        })
      }
    }

    // Small delay to ensure layout is ready
    setTimeout(updatePosition, 10)

    window.addEventListener('resize', updatePosition)
    return () => window.removeEventListener('resize', updatePosition)
  }, [statusFilter])

  const domain = process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'

  return (
    <div className="space-y-5">
      {/* Filter and Actions */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 w-full sm:w-auto">
          {/* Status Filter */}
          <div className="inline-flex gap-1 relative">
            <button
              ref={allRef}
              onClick={() => setStatusFilter('all')}
              className={cn(
                'relative px-5 py-2 text-sm font-medium transition-all duration-200 cursor-pointer',
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
                'relative px-5 py-2 text-sm font-medium transition-all duration-200 cursor-pointer',
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
                'relative px-5 py-2 text-sm font-medium transition-all duration-200 cursor-pointer',
                'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
                statusFilter === 'installed'
                  ? 'text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              Installed
            </button>

            {/* Animated underline with CSS transition */}
            {underlineStyle.width > 0 && (
              <div
                className="absolute bottom-0.5 h-[2px] bg-primary transition-all duration-300 ease-out"
                style={{
                  left: `${underlineStyle.left}px`,
                  width: `${underlineStyle.width}px`,
                }}
              />
            )}
          </div>
        </div>
        <Button size="default" className="w-full sm:w-auto" onClick={() => router.push('/installations/new')}>
          <Plus className="h-4 w-4" />
          New Installation
        </Button>
      </div>

      {/* Table */}
      {filteredInstallations.length > 0 ? (
        <div className="overflow-hidden border border-foreground/10 rounded-lg">
          <div className="overflow-x-auto">
            <table className="min-w-full">
              <thead>
                <tr className="border-b border-foreground/10 bg-muted/30">
                  <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide">
                    Name
                  </th>
                  <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide hidden md:table-cell">
                    Domain
                  </th>
                  <th className="text-muted-foreground px-6 py-3.5 text-left text-xs font-semibold tracking-wide">
                    Status
                  </th>
                  <th className="text-muted-foreground px-6 py-3.5 text-right text-xs font-semibold tracking-wide">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-background divide-y divide-foreground/5">
                {filteredInstallations.map((installation) => {
                  const { status, statusColor, isPending } = getInstallationStatus(installation)
                  // Validate subdomain before constructing URL
                  const isValidSubdomain =
                    installation.subdomain &&
                    installation.subdomain !== '0.0.0.0' &&
                    !installation.subdomain.includes('0.0.0.0')
                  const installationUrl = isValidSubdomain
                    ? `https://${installation.subdomain}.${domain}`
                    : null

                  return (
                    <tr key={installation.id} className="group hover:bg-muted/20 transition-colors relative">
                      <td className="px-6 py-3.5 relative">
                        {/* Left accent bar */}
                        <div className="absolute left-0 top-0 w-[2px] h-full bg-primary scale-y-0 group-hover:scale-y-100 transition-transform duration-200 origin-top" />

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
                      <td className="px-6 py-3.5 hidden md:table-cell">
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
                      <td className="px-6 py-3.5">
                        <span
                          className={`inline-flex px-2.5 py-1 text-[10px] font-semibold uppercase tracking-wider rounded-md ${statusColor}`}
                        >
                          {status}
                        </span>
                      </td>
                      <td className="px-6 py-3.5 text-right">
                        <div className="flex items-center justify-end gap-2">
                          {isPending ? (
                            <Button
                              variant="default"
                              size="sm"
                              onClick={() => router.push(`/installations/${installation.id}`)}
                            >
                              Continue
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
                              <DropdownMenuItem
                                onSelect={() => router.push(`/installations/${installation.id}`)}
                              >
                                <Settings className="mr-2 h-4 w-4" />
                                Settings
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
        <div className="border border-foreground/10 rounded-lg py-12 text-center bg-muted/10">
          <div className="mx-auto max-w-md px-4">
            <div className="mx-auto w-10 h-10 bg-muted rounded-lg border border-foreground/10 flex items-center justify-center mb-3">
              <Plus className="h-4 w-4 text-muted-foreground" />
            </div>
            <h3 className="text-foreground text-sm font-semibold mb-1">
              {statusFilter === 'pending'
                ? 'No pending installations'
                : statusFilter === 'installed'
                  ? 'No active installations'
                  : 'No installations'}
            </h3>
            <p className="text-muted-foreground text-sm mb-5 leading-relaxed">
              {statusFilter === 'all'
                ? 'Create your first installation to get started'
                : 'Adjust filters to see more results'}
            </p>
            {statusFilter === 'all' && (
              <Button size="default" onClick={() => router.push('/installations/new')}>
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
