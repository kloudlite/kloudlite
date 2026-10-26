'use client'

import { useState, useEffect, useCallback, useMemo } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button, Input } from '@kloudlite/ui'
import { MoreHorizontal, Settings, Search, Loader2 } from 'lucide-react'
import { NewInstallationButton } from '@/components/new-installation-button'
import { DeleteInstallationButton } from '@/components/delete-installation-button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@kloudlite/ui'
import { Tabs, TabsList, TabsTrigger } from '@kloudlite/ui'
import { cn } from '@kloudlite/lib'
import { getInstallationStatus, hasActiveJob } from '@/lib/installation-status'
import type { Installation, BillingAccount } from '@/lib/console/storage'

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
  activeSubscriptions?: Record<string, BillingAccount>
}

function isExpiringSoon(customer: BillingAccount | undefined): boolean {
  if (!customer?.currentPeriodEnd || customer.billingStatus !== 'active') return false
  const msUntilEnd = new Date(customer.currentPeriodEnd).getTime() - Date.now()
  const daysUntilEnd = Math.ceil(msUntilEnd / (24 * 60 * 60 * 1000))
  return daysUntilEnd <= 7 && daysUntilEnd > 0
}

export function InstallationsList({
  installations,
  activeSubscriptions = {},
}: InstallationsListProps) {
  const router = useRouter()
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'installed'>('all')
  const [searchQuery, setSearchQuery] = useState('')

  // Check if installation needs polling (active job or pending uninstall deletion)
  const needsPolling = useCallback((installation: Installation) => {
    return (
      hasActiveJob(installation) ||
      (installation.deployJobOperation === 'uninstall' && installation.deployJobStatus !== 'failed')
    )
  }, [])

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
      // Sync job status from ACA API → DB for active installations (parallel)
      const pollingInstallations = installations.filter(needsPolling)
      await Promise.allSettled(
        pollingInstallations.map((inst) =>
          fetch(`/api/installations/${inst.id}/job-status`)
        )
      )
      router.refresh()
    }, 5000)
    return () => clearInterval(interval)
  }, [hasAnyActiveJob, installations, router, needsPolling])

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

  const liveRegionMessage = useMemo(() => {
    const count = filteredInstallations.length
    const label = count === 1 ? 'installation' : 'installations'
    if (searchQuery.trim()) {
      return `${count} ${label} match search "${searchQuery.trim()}".`
    }
    return `${count} ${label} shown in ${statusFilter} filter.`
  }, [filteredInstallations.length, searchQuery, statusFilter])

  return (
    <div className="space-y-5">
      <p className="sr-only" role="status" aria-live="polite" aria-atomic="true">
        {liveRegionMessage}
      </p>
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
      <div className="mb-5">
        <Tabs value={statusFilter} onValueChange={(v) => setStatusFilter(v as 'all' | 'pending' | 'installed')}>
          <TabsList className="inline-flex gap-1 rounded-lg bg-muted/50 p-1">
            <TabsTrigger value="all" className="rounded-md px-3.5 py-1.5 text-sm data-[state=active]:bg-background data-[state=active]:shadow-sm">All</TabsTrigger>
            <TabsTrigger value="pending" className="rounded-md px-3.5 py-1.5 text-sm data-[state=active]:bg-background data-[state=active]:shadow-sm">Pending</TabsTrigger>
            <TabsTrigger value="installed" className="rounded-md px-3.5 py-1.5 text-sm data-[state=active]:bg-background data-[state=active]:shadow-sm">Installed</TabsTrigger>
          </TabsList>
        </Tabs>
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
                    Location
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
                            {installation.cloudLocation ? (
                              <span className="text-muted-foreground/50 font-mono text-[11px]">
                                {installation.cloudLocation}
                              </span>
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
                        {installation.cloudLocation ? (
                          <span className="text-primary/80 font-mono text-[13px]">
                            {installation.cloudLocation}
                          </span>
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
                          {activeSubscriptions[installation.id]?.hasPaymentIssue && (
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
                                <DeleteInstallationButton
                                  installationId={installation.id}
                                  installationName={installation.name}
                                  hasSecretKey={!!installation.secretKey}
                                  cloudProvider={installation.cloudProvider}
                                  variant="menu"
                                />
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
