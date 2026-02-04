'use client'

import { useState, useEffect } from 'react'
import {
  Package,
  CheckCircle2,
  XCircle,
  Loader2,
  Copy,
  Check,
  ExternalLink,
} from 'lucide-react'
import type { Workspace, PackageRequest } from '@kloudlite/types'
import { getPackageRequest } from '@/app/actions/workspace.actions'

interface PackageWithStatus {
  name: string
  version?: string
  channel?: string
  nixpkgsCommit?: string
  status: 'installed' | 'pending' | 'failed'
}

interface PackagesListProps {
  workspace: Workspace
  initialPackageRequest?: PackageRequest | null
}

export function PackagesList({ workspace, initialPackageRequest }: PackagesListProps) {
  const [packages, setPackages] = useState<PackageWithStatus[]>([])
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null)

  useEffect(() => {
    const loadPackages = async () => {
      let pkgReq = initialPackageRequest

      if (!pkgReq) {
        const result = await getPackageRequest(workspace.metadata.name, workspace.metadata.namespace)
        pkgReq = result.success ? (result.data as unknown as PackageRequest) : null
      }

      const statusPhase = pkgReq?.status?.phase || 'Pending'
      const failedPackage = pkgReq?.status?.failedPackage || ''

      const existingPackages: PackageWithStatus[] = (pkgReq?.spec?.packages || []).map((pkg) => {
        let status: 'installed' | 'pending' | 'failed' = 'pending'
        if (statusPhase === 'Ready') {
          status = 'installed'
        } else if (statusPhase === 'Failed' && failedPackage === pkg.name) {
          status = 'failed'
        }

        return {
          name: pkg.name,
          version: pkg.channel || (pkg.nixpkgsCommit ? pkg.nixpkgsCommit.substring(0, 8) : undefined),
          channel: pkg.channel,
          nixpkgsCommit: pkg.nixpkgsCommit,
          status,
        }
      })

      setPackages(existingPackages)
    }

    loadPackages()
  }, [workspace.metadata.name, workspace.metadata.namespace, initialPackageRequest])

  const handleCopy = (command: string, id: string) => {
    navigator.clipboard.writeText(command)
    setCopiedCommand(id)
    setTimeout(() => setCopiedCommand(null), 2000)
  }

  const installedCount = packages.filter(p => p.status === 'installed').length
  const pendingCount = packages.filter(p => p.status === 'pending').length
  const failedCount = packages.filter(p => p.status === 'failed').length

  return (
    <div className="space-y-6">
      {/* Package List */}
      <div className="bg-card rounded-lg border">
        <div className="border-b p-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-semibold">Installed Packages</h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                {packages.length === 0
                  ? 'No packages installed yet'
                  : `${packages.length} package${packages.length === 1 ? '' : 's'} configured`
                }
              </p>
            </div>
            {packages.length > 0 && (
              <div className="flex items-center gap-2">
                {installedCount > 0 && (
                  <span className="inline-flex items-center gap-1.5 rounded-md bg-green-100 dark:bg-green-900/30 px-2.5 py-1 text-xs font-medium text-green-700 dark:text-green-400">
                    <CheckCircle2 className="h-3 w-3" />
                    {installedCount} installed
                  </span>
                )}
                {pendingCount > 0 && (
                  <span className="inline-flex items-center gap-1.5 rounded-md bg-yellow-100 dark:bg-yellow-900/30 px-2.5 py-1 text-xs font-medium text-yellow-700 dark:text-yellow-400">
                    <Loader2 className="h-3 w-3 animate-spin" />
                    {pendingCount} pending
                  </span>
                )}
                {failedCount > 0 && (
                  <span className="inline-flex items-center gap-1.5 rounded-md bg-red-100 dark:bg-red-900/30 px-2.5 py-1 text-xs font-medium text-red-700 dark:text-red-400">
                    <XCircle className="h-3 w-3" />
                    {failedCount} failed
                  </span>
                )}
              </div>
            )}
          </div>
        </div>

        {packages.length > 0 ? (
          <div className="divide-y">
            {packages.map((pkg, index) => (
              <div
                key={index}
                className="flex items-center gap-4 px-4 py-3"
              >
                <div className={`flex-shrink-0 rounded-lg p-2 ${
                  pkg.status === 'installed'
                    ? 'bg-green-100 dark:bg-green-900/30'
                    : pkg.status === 'failed'
                      ? 'bg-red-100 dark:bg-red-900/30'
                      : 'bg-yellow-100 dark:bg-yellow-900/30'
                }`}>
                  <Package className={`h-4 w-4 ${
                    pkg.status === 'installed'
                      ? 'text-green-600 dark:text-green-400'
                      : pkg.status === 'failed'
                        ? 'text-red-600 dark:text-red-400'
                        : 'text-yellow-600 dark:text-yellow-400'
                  }`} />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="font-mono text-sm font-medium truncate">{pkg.name}</p>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {pkg.channel
                      ? `Channel: ${pkg.channel}`
                      : pkg.nixpkgsCommit
                        ? `Commit: ${pkg.nixpkgsCommit.substring(0, 8)}`
                        : 'Latest version'
                    }
                  </p>
                </div>
                <div className="flex-shrink-0">
                  {pkg.status === 'installed' && (
                    <span className="text-xs text-green-600 dark:text-green-400 font-medium">Installed</span>
                  )}
                  {pkg.status === 'pending' && (
                    <span className="text-xs text-yellow-600 dark:text-yellow-400 font-medium flex items-center gap-1">
                      <Loader2 className="h-3 w-3 animate-spin" />
                      Installing
                    </span>
                  )}
                  {pkg.status === 'failed' && (
                    <span className="text-xs text-red-600 dark:text-red-400 font-medium">Failed</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="p-8 text-center">
            <div className="mx-auto w-12 h-12 rounded-md bg-muted flex items-center justify-center mb-4">
              <Package className="h-6 w-6 text-muted-foreground" />
            </div>
            <h4 className="text-sm font-medium mb-1">No packages installed</h4>
            <p className="text-xs text-muted-foreground max-w-sm mx-auto">
              Use the CLI commands below to install packages in your workspace
            </p>
          </div>
        )}
      </div>

      {/* CLI Instructions */}
      <div className="bg-card rounded-lg border">
        <div className="border-b p-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-semibold">Managing Packages</h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                Use these commands inside your workspace terminal
              </p>
            </div>
            <a
              href="https://search.nixos.org/packages"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
            >
              Search Nix Packages
              <ExternalLink className="h-3 w-3" />
            </a>
          </div>
        </div>
        <div className="p-4">
          <div className="rounded-lg bg-zinc-950 dark:bg-zinc-900 p-4 font-mono text-sm">
            <div className="space-y-3">
              <CommandLine
                comment="Install a package"
                command="kl pkg add nodejs"
                onCopy={handleCopy}
                copied={copiedCommand === 'add'}
                id="add"
              />
              <CommandLine
                comment="Install specific version"
                command="kl pkg add nodejs@20.10.0"
                onCopy={handleCopy}
                copied={copiedCommand === 'add-version'}
                id="add-version"
              />
              <CommandLine
                comment="Remove a package"
                command="kl pkg remove nodejs"
                onCopy={handleCopy}
                copied={copiedCommand === 'remove'}
                id="remove"
              />
              <CommandLine
                comment="List installed packages"
                command="kl pkg list"
                onCopy={handleCopy}
                copied={copiedCommand === 'list'}
                id="list"
              />
              <CommandLine
                comment="Search for packages"
                command="kl pkg search python"
                onCopy={handleCopy}
                copied={copiedCommand === 'search'}
                id="search"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function CommandLine({
  comment,
  command,
  onCopy,
  copied,
  id
}: {
  comment: string
  command: string
  onCopy: (command: string, id: string) => void
  copied: boolean
  id: string
}) {
  return (
    <div className="group">
      <div className="text-zinc-500 text-xs mb-1"># {comment}</div>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-green-400">$</span>
          <span className="text-zinc-100">{command}</span>
        </div>
        <button
          onClick={() => onCopy(command, id)}
          className="opacity-0 group-hover:opacity-100 transition-opacity p-1 hover:bg-zinc-800 rounded"
        >
          {copied ? (
            <Check className="h-3.5 w-3.5 text-green-400" />
          ) : (
            <Copy className="h-3.5 w-3.5 text-zinc-500" />
          )}
        </button>
      </div>
    </div>
  )
}
