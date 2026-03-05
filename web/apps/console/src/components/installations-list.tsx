'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button, Input } from '@kloudlite/ui'
import { MoreHorizontal, ExternalLink, Settings, Search, Loader2 } from 'lucide-react'
import { NewInstallationButton } from '@/components/new-installation-button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import { cn } from '@kloudlite/lib'
import type { Installation, Invoice, Subscription } from '@/lib/console/storage'

const providerConfig: Record<string, { label: string; className: string }> = {
  aws: {
    label: 'AWS',
    className: 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20',
  },
  gcp: {
    label: 'GCP',
    className: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20',
  },
  azure: {
    label: 'Azure',
    className: 'bg-sky-500/10 text-sky-700 dark:text-sky-400 border-sky-500/20',
  },
  oci: { label: 'Kloudlite', className: 'bg-primary/10 text-primary border-primary/20' },
}

function ProviderBadge({ provider }: { provider?: string }) {
  if (!provider || !providerConfig[provider]) {
    return <span className="text-muted-foreground/40">{'\u2014'}</span>
  }
  const { label, className } = providerConfig[provider]
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-md border px-2 py-0.5 text-[11px] font-medium',
        className,
      )}
    >
      {label}
    </span>
  )
}

interface InstallationsListProps {
  installations: Installation[]
  pendingInvoices: Record<string, Invoice>
  activeSubscriptions?: Record<string, Subscription>
}

function isExpiringSoon(sub: Subscription | undefined): boolean {
  if (!sub?.currentEnd || sub.status !== 'active') return false
  const msUntilEnd = new Date(sub.currentEnd).getTime() - Date.now()
  const daysUntilEnd = Math.ceil(msUntilEnd / (24 * 60 * 60 * 1000))
  return daysUntilEnd <= 7 && daysUntilEnd > 0
}

export function InstallationsList({
  installations,
  pendingInvoices,
  activeSubscriptions = {},
}: InstallationsListProps) {
  const router = useRouter()
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'installed'>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [underlineStyle, setUnderlineStyle] = useState({ left: 0, width: 0 })

  const allRef = useRef<HTMLButtonElement>(null)
  const pendingRef = useRef<HTMLButtonElement>(null)
  const installedRef = useRef<HTMLButtonElement>(null)

  // Check if an installation has an active job (running or pending)
  const hasActiveJob = (installation: Installation) => {
    return (
      (installation.acaJobStatus === 'running' || installation.acaJobStatus === 'pending') &&
      (installation.acaJobOperation === 'install' || installation.acaJobOperation === 'uninstall')
    )
  }

  // Helper function to get installation status
  const getInstallationStatus = (installation: Installation) => {
    // Uninstall operations: show UNINSTALLING until the record is auto-deleted
    if (installation.acaJobOperation === 'uninstall') {
      if (installation.acaJobStatus === 'failed') {
        return {
          status: 'UNINSTALL FAILED',
          statusColor: 'bg-red-500/10 text-red-700 dark:text-red-400 border border-red-500/20',
          isPending: false,
          isActiveJob: false,
          stepInfo: undefined,
        }
      }
      return {
        status: 'UNINSTALLING',
        statusColor: 'bg-red-500/10 text-red-700 dark:text-red-400 border border-red-500/20',
        isPending: false,
        isActiveJob: true,
        stepInfo:
          installation.acaJobCurrentStep && installation.acaJobTotalSteps
            ? `Step ${installation.acaJobCurrentStep}/${installation.acaJobTotalSteps}`
            : undefined,
      }
    }

    // Check active install jobs
    if (hasActiveJob(installation) && installation.acaJobOperation === 'install') {
      return {
        status: 'INSTALLING',
        statusColor: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/20',
        isPending: false,
        isActiveJob: true,
        stepInfo:
          installation.acaJobCurrentStep && installation.acaJobTotalSteps
            ? `Step ${installation.acaJobCurrentStep}/${installation.acaJobTotalSteps}`
            : undefined,
      }
    }

    // Check for failed jobs (installation not ready and job failed)
    if (installation.acaJobStatus === 'failed' && !installation.deploymentReady) {
      return {
        status: 'ERROR',
        statusColor: 'bg-red-500/10 text-red-700 dark:text-red-400 border border-red-500/20',
        isPending: true,
        isActiveJob: false,
        stepInfo: undefined,
      }
    }

    if (!installation.secretKey) {
      return {
        status: 'NOT INSTALLED',
        statusColor: 'bg-foreground/[0.06] text-foreground border border-foreground/10',
        isPending: true,
        isActiveJob: false,
        stepInfo: undefined,
      }
    }
    if (!installation.subdomain) {
      return {
        status: 'PENDING',
        statusColor:
          'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border border-yellow-500/20',
        isPending: true,
        isActiveJob: false,
        stepInfo: undefined,
      }
    }
    if (!installation.deploymentReady) {
      return {
        status: 'CONFIGURING',
        statusColor: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/20',
        isPending: true,
        isActiveJob: false,
        stepInfo: undefined,
      }
    }
    return {
      status: 'ACTIVE',
      statusColor: 'bg-green-500/10 text-green-700 dark:text-green-400 border border-green-500/20',
      isPending: false,
      isActiveJob: false,
      stepInfo: undefined,
    }
  }

  // Check if installation needs polling (active job or pending uninstall deletion)
  const needsPolling = useCallback((installation: Installation) => {
    return (
      hasActiveJob(installation) ||
      (installation.acaJobOperation === 'uninstall' && installation.acaJobStatus !== 'failed')
    )
  }, [])

  // Filter change handlers
  const handleFilterAll = useCallback(() => setStatusFilter('all'), [])
  const handleFilterPending = useCallback(() => setStatusFilter('pending'), [])
  const handleFilterInstalled = useCallback(() => setStatusFilter('installed'), [])

  // Installation navigation handlers
  const handleContinueInstallation = useCallback(
    (installationId: string) => {
      router.push(`/api/installations/${installationId}/continue`)
    },
    [router],
  )

  const handleViewSettings = useCallback(
    (installationId: string) => {
      router.push(`/installations/${installationId}`)
    },
    [router],
  )

  // Auto-refresh when any installation has an active job or pending uninstall
  const hasAnyActiveJob = installations.some(needsPolling)
  useEffect(() => {
    if (!hasAnyActiveJob) return
    const interval = setInterval(async () => {
      // Sync job status from ACA API → DB for active installations
      for (const inst of installations) {
        if (needsPolling(inst)) {
          try {
            await fetch(`/api/installations/${inst.id}/job-status`)
          } catch {}
        }
      }
      router.refresh()
    }, 5000)
    return () => clearInterval(interval)
  }, [hasAnyActiveJob, router, installations, needsPolling])

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

  // Apply search filter
  if (searchQuery.trim()) {
    filteredInstallations = filteredInstallations.filter((install) => {
      const query = searchQuery.toLowerCase()
      return (
        install.name?.toLowerCase().includes(query) ||
        install.description?.toLowerCase().includes(query) ||
        install.subdomain?.toLowerCase().includes(query)
      )
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
          width: underlineWidth,
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
      {/* Page Header */}
      <div className="mb-6 flex items-center justify-between pb-6">
        <h1 className="text-foreground text-2xl font-semibold tracking-tight">Installations</h1>
        <div className="flex items-center gap-3">
          <div className="relative w-64">
            <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
            <Input
              type="text"
              placeholder="Search installations..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>
          <NewInstallationButton />
        </div>
      </div>

      {/* Status Filter Tabs */}
      <div className="border-foreground/10 mb-5 border-b">
        <div className="relative inline-flex gap-1">
          <button
            ref={allRef}
            onClick={handleFilterAll}
            className={cn(
              'relative cursor-pointer px-5 py-2 text-sm font-medium transition-all duration-200',
              'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
              statusFilter === 'all'
                ? 'text-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            All
          </button>
          <button
            ref={pendingRef}
            onClick={handleFilterPending}
            className={cn(
              'relative cursor-pointer px-5 py-2 text-sm font-medium transition-all duration-200',
              'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
              statusFilter === 'pending'
                ? 'text-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            Pending
          </button>
          <button
            ref={installedRef}
            onClick={handleFilterInstalled}
            className={cn(
              'relative cursor-pointer px-5 py-2 text-sm font-medium transition-all duration-200',
              'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
              statusFilter === 'installed'
                ? 'text-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            Installed
          </button>

          {/* Animated underline with CSS transition */}
          {underlineStyle.width > 0 && (
            <div
              className="bg-primary absolute bottom-0 h-[2px] transition-all duration-300 ease-out"
              style={{
                left: `${underlineStyle.left}px`,
                width: `${underlineStyle.width}px`,
              }}
            />
          )}
        </div>
      </div>

      {/* Table */}
      {filteredInstallations.length > 0 ? (
        <div className="border-foreground/10 overflow-hidden rounded-lg border">
          <div className="overflow-x-auto">
            <table className="min-w-full">
              <thead>
                <tr className="border-foreground/10 bg-muted/30 border-b">
                  <th className="text-muted-foreground w-[30%] px-6 py-3.5 text-left text-xs font-semibold tracking-wide">
                    Name
                  </th>
                  <th className="text-muted-foreground hidden w-[10%] px-6 py-3.5 text-left text-xs font-semibold tracking-wide lg:table-cell">
                    Provider
                  </th>
                  <th className="text-muted-foreground hidden w-[30%] px-6 py-3.5 text-left text-xs font-semibold tracking-wide md:table-cell">
                    Domain
                  </th>
                  <th className="text-muted-foreground w-[15%] px-6 py-3.5 text-left text-xs font-semibold tracking-wide">
                    Status
                  </th>
                  <th className="text-muted-foreground w-[15%] px-6 py-3.5 text-right text-xs font-semibold tracking-wide">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-background divide-foreground/5 divide-y">
                {filteredInstallations.map((installation) => {
                  const { status, statusColor, isPending, isActiveJob, stepInfo } =
                    getInstallationStatus(installation)
                  // Validate subdomain before constructing URL
                  const isValidSubdomain =
                    installation.subdomain &&
                    installation.subdomain !== '0.0.0.0' &&
                    !installation.subdomain.includes('0.0.0.0')
                  const installationUrl = isValidSubdomain
                    ? `https://${installation.subdomain}.${domain}`
                    : null
                  const displayName =
                    installation.name || installation.subdomain || 'Unnamed Installation'

                  return (
                    <tr
                      key={installation.id}
                      className="group hover:bg-muted/20 relative transition-colors"
                    >
                      <td className="relative px-6 py-3">
                        <div className="relative z-10">
                          <Link
                            href={`/installations/${installation.id}`}
                            className="text-foreground group-hover:text-primary cursor-pointer text-sm font-medium transition-colors hover:cursor-pointer"
                          >
                            {displayName}
                          </Link>
                          {installation.description && installation.description !== displayName && (
                            <div className="text-muted-foreground/60 mt-0.5 line-clamp-1 text-xs">
                              {installation.description}
                            </div>
                          )}
                          {/* Show domain on mobile */}
                          <div className="mt-1 md:hidden">
                            {installationUrl ? (
                              <a
                                href={installationUrl}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-primary/80 hover:text-primary flex items-center gap-1 font-mono text-[11px] transition-colors hover:underline"
                              >
                                {installation.subdomain}.{domain}
                                <ExternalLink className="h-3 w-3" />
                              </a>
                            ) : (
                              <span className="text-muted-foreground/50 text-xs">
                                Not configured
                              </span>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="hidden px-6 py-3 lg:table-cell">
                        <ProviderBadge provider={installation.cloudProvider} />
                      </td>
                      <td className="hidden px-6 py-3 md:table-cell">
                        {installationUrl ? (
                          <a
                            href={installationUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-primary/80 hover:text-primary inline-flex items-center gap-1.5 font-mono text-[13px] transition-colors hover:underline"
                          >
                            {installation.subdomain}.{domain}
                            <ExternalLink className="h-3 w-3" />
                          </a>
                        ) : (
                          <span className="text-muted-foreground/50 text-sm">Not configured</span>
                        )}
                      </td>
                      <td className="px-6 py-3">
                        <div className="flex items-center gap-2">
                          <span
                            className={`inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-[10px] font-semibold tracking-wider whitespace-nowrap uppercase ${statusColor}`}
                          >
                            {isActiveJob && <Loader2 className="h-3 w-3 animate-spin" />}
                            {status}
                          </span>
                          {isExpiringSoon(activeSubscriptions[installation.id]) && (
                            <span className="inline-flex items-center rounded-md border border-amber-500/20 bg-amber-500/10 px-2 py-0.5 text-[10px] font-semibold tracking-wider whitespace-nowrap text-amber-700 uppercase dark:text-amber-400">
                              Expiring Soon
                            </span>
                          )}
                          {pendingInvoices[installation.id] && (
                            <span className="inline-flex items-center rounded-md border border-amber-500/20 bg-amber-500/10 px-2 py-0.5 text-[10px] font-semibold tracking-wider whitespace-nowrap text-amber-700 uppercase dark:text-amber-400">
                              Payment Due
                            </span>
                          )}
                          {stepInfo && (
                            <span className="text-muted-foreground text-[11px] whitespace-nowrap">
                              {stepInfo}
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-3 text-right">
                        <div className="flex items-center justify-end gap-2">
                          {isPending && !isActiveJob && (
                            <Button
                              variant="default"
                              size="sm"
                              onClick={() => handleContinueInstallation(installation.id)}
                            >
                              Continue
                            </Button>
                          )}
                          {!isActiveJob && (
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon">
                                  <MoreHorizontal className="h-4 w-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                {!isPending && (
                                  <DropdownMenuItem
                                    onSelect={() => handleViewSettings(installation.id)}
                                  >
                                    <Settings className="mr-2 h-4 w-4" />
                                    Settings
                                  </DropdownMenuItem>
                                )}
                              </DropdownMenuContent>
                            </DropdownMenu>
                          )}
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
        <div className="border-foreground/10 bg-muted/10 rounded-lg border py-12 text-center">
          <div className="mx-auto max-w-md px-4">
            <div className="bg-muted border-foreground/10 mx-auto mb-3 flex h-10 w-10 items-center justify-center rounded-lg border">
              <Search className="text-muted-foreground h-4 w-4" />
            </div>
            <h3 className="text-foreground mb-1 text-sm font-semibold">
              {searchQuery.trim()
                ? 'No installations found'
                : statusFilter === 'pending'
                  ? 'No pending installations'
                  : statusFilter === 'installed'
                    ? 'No active installations'
                    : 'No installations'}
            </h3>
            <p className="text-muted-foreground mb-5 text-sm leading-relaxed">
              {searchQuery.trim()
                ? 'Try adjusting your search query or filters'
                : statusFilter === 'all'
                  ? 'Create your first installation to get started'
                  : 'Adjust filters to see more results'}
            </p>
          </div>
        </div>
      )}
    </div>
  )
}
