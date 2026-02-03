'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import {
  Package,
  Plus,
  X,
  Loader2,
  CheckCircle2,
  XCircle,
  Check,
  ChevronsUpDown,
  Save,
} from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@kloudlite/ui'
import { Popover, PopoverContent, PopoverTrigger } from '@kloudlite/ui'
import { ScrollArea } from '@kloudlite/ui'
import { cn } from '@/lib/utils'
import type { Workspace, PackageSpec, PackageRequest } from '@kloudlite/types'
import { updatePackageRequest, getPackageRequest } from '@/app/actions/workspace.actions'
import { searchPackages, resolvePackageVersion } from '@/app/actions/package.actions'
import { toast } from 'sonner'

interface PackageWithVersion extends PackageSpec {
  displayVersion?: string
  status?: 'installed' | 'pending' | 'failed'
}

interface PackagesManagerProps {
  workspace: Workspace
  initialPackageRequest?: PackageRequest | null
}

export function PackagesManager({ workspace, initialPackageRequest }: PackagesManagerProps) {
  const router = useRouter()
  const [packages, setPackages] = useState<PackageWithVersion[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [hasChanges, setHasChanges] = useState(false)

  // Package search state
  const [newPackageName, setNewPackageName] = useState('')
  const [newPackageVersion, setNewPackageVersion] = useState('')
  const [availableVersions, setAvailableVersions] = useState<string[]>([])
  const [loadingVersions, setLoadingVersions] = useState(false)
  const [searchResults, setSearchResults] = useState<Array<{ name: string; numVersions: number }>>(
    [],
  )
  const [loadingSearch, setLoadingSearch] = useState(false)
  const [comboboxOpen, setComboboxOpen] = useState(false)

  // Initialize packages from initial data or fetch
  useEffect(() => {
    const loadPackageStatus = async () => {
      let pkgReq = initialPackageRequest

      // If no initial data, fetch it
      if (!pkgReq) {
        const pkgReqResult = await getPackageRequest(workspace.metadata.name, workspace.metadata.namespace)
        pkgReq = pkgReqResult.success ? (pkgReqResult.data as unknown as PackageRequest) : null
      }

      // Get status phase from PackageRequest
      const statusPhase = pkgReq?.status?.phase || 'Pending'
      const failedPackage = pkgReq?.status?.failedPackage || ''

      // Load packages from PackageRequest.spec.packages (source of truth)
      const existingPackages: PackageWithVersion[] = (pkgReq?.spec?.packages || []).map((pkg) => {
        let status: 'installed' | 'pending' | 'failed' = 'pending'
        if (statusPhase === 'Ready') {
          status = 'installed'
        } else if (statusPhase === 'Failed' && failedPackage === pkg.name) {
          status = 'failed'
        }

        return {
          ...pkg,
          displayVersion:
            pkg.channel ||
            (pkg.nixpkgsCommit ? `commit:${pkg.nixpkgsCommit.substring(0, 8)}` : undefined),
          status,
        }
      })
      setPackages(existingPackages)
    }

    loadPackageStatus()
  }, [workspace.metadata.name, workspace.metadata.namespace, initialPackageRequest])

  // Search for packages as user types in combobox
  useEffect(() => {
    const searchTimer = setTimeout(async () => {
      if (!comboboxOpen || newPackageName.length < 2) {
        setSearchResults([])
        return
      }

      setLoadingSearch(true)
      const result = await searchPackages(newPackageName)
      setLoadingSearch(false)

      if (result.success && result.data?.packages) {
        const packagesInfo = result.data.packages.map((p) => ({
          name: p.name,
          numVersions: p.num_versions,
        }))
        setSearchResults(packagesInfo.slice(0, 15))
      } else {
        setSearchResults([])
      }
    }, 300)

    return () => clearTimeout(searchTimer)
  }, [newPackageName, comboboxOpen])

  // Load versions when package name is selected
  useEffect(() => {
    const loadVersions = async () => {
      if (!newPackageName.trim() || newPackageName.length < 2) {
        setAvailableVersions([])
        return
      }

      setLoadingVersions(true)
      const result = await searchPackages(newPackageName)
      setLoadingVersions(false)

      if (result.success && result.data?.packages) {
        const pkg = result.data.packages.find((p) => p.name === newPackageName)
        if (pkg && pkg.versions.length > 0) {
          const versions = pkg.versions.map((v) => v.version)
          setAvailableVersions(versions.slice(0, 50))
        } else {
          setAvailableVersions([])
        }
      } else {
        setAvailableVersions([])
      }
    }

    loadVersions()
  }, [newPackageName])

  const addPackage = async () => {
    if (!newPackageName.trim()) {
      toast.error('Please enter a package name')
      return
    }

    if (!newPackageVersion.trim()) {
      toast.error('Please select a version')
      return
    }

    toast.loading('Resolving package version...')
    const result = await resolvePackageVersion(newPackageName.trim(), newPackageVersion.trim())
    toast.dismiss()

    if (!result.success || !result.data) {
      toast.error(result.error || 'Failed to resolve package version')
      return
    }

    const systems = Object.values(result.data.systems)
    if (systems.length === 0) {
      toast.error('No system information available for this package version')
      return
    }

    const commitHash = systems[0].flake_installable.ref.rev
    const attrPath = systems[0].flake_installable.attr_path

    const pkg: PackageWithVersion = {
      name: attrPath,
      nixpkgsCommit: commitHash,
      displayVersion: newPackageVersion.trim(),
      status: 'pending',
    }

    setPackages([...packages, pkg])
    setNewPackageName('')
    setNewPackageVersion('')
    setAvailableVersions([])
    setSearchResults([])
    setHasChanges(true)
    toast.success('Package added')
  }

  const removePackage = (index: number) => {
    setPackages(packages.filter((_, i) => i !== index))
    setHasChanges(true)
  }

  const handleSave = async () => {
    setIsLoading(true)

    // Convert PackageWithVersion to PackageSpec for the API
    const packageSpecs = packages.map(
      ({
        displayVersion: _displayVersion,
        status: _status,
        ...pkg
      }) => pkg,
    )

    // Update PackageRequest directly (creates it if it doesn't exist)
    const result = await updatePackageRequest(
      workspace.metadata.name,
      packageSpecs,
      workspace.metadata.namespace,
    )

    if (result.success) {
      toast.success('Packages updated successfully')
      setHasChanges(false)
      router.refresh()
    } else {
      toast.error(result.error || 'Failed to update packages')
    }

    setIsLoading(false)
  }

  // Get package status info
  const installedCount = packages.filter(p => p.status === 'installed').length
  const pendingCount = packages.filter(p => p.status === 'pending').length
  const failedCount = packages.filter(p => p.status === 'failed').length

  return (
    <div className="space-y-6">
      {/* Status Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">Package Management</h2>
          <p className="text-sm text-muted-foreground">
            Manage packages for this workspace using Nix package manager
          </p>
        </div>
        <div className="flex items-center gap-3">
          {packages.length > 0 && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              {installedCount > 0 && (
                <span className="flex items-center gap-1">
                  <CheckCircle2 className="h-4 w-4 text-success" />
                  {installedCount} installed
                </span>
              )}
              {pendingCount > 0 && (
                <span className="flex items-center gap-1">
                  <Loader2 className="h-4 w-4 text-warning animate-spin" />
                  {pendingCount} pending
                </span>
              )}
              {failedCount > 0 && (
                <span className="flex items-center gap-1">
                  <XCircle className="h-4 w-4 text-destructive" />
                  {failedCount} failed
                </span>
              )}
            </div>
          )}
          {hasChanges && (
            <Button onClick={handleSave} disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Save Changes
                </>
              )}
            </Button>
          )}
        </div>
      </div>

      {/* Add Package Form */}
      <div className="bg-card rounded-lg border p-6">
        <h3 className="text-sm font-medium mb-4">Add New Package</h3>
        <div className="flex gap-4">
          {/* Package Name */}
          <div className="flex-1 space-y-2">
            <Label className="flex h-5 items-center">
              Package Name
              {loadingSearch && <Loader2 className="ml-2 inline h-3 w-3 animate-spin" />}
            </Label>
            <Popover open={comboboxOpen} onOpenChange={setComboboxOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  aria-expanded={comboboxOpen}
                  className={cn(
                    'w-full justify-between',
                    !newPackageName && 'text-muted-foreground',
                  )}
                  disabled={isLoading}
                >
                  <span className="truncate">{newPackageName || 'Search packages...'}</span>
                  <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent
                className="w-[var(--radix-popover-trigger-width)] p-0"
                align="start"
              >
                <Command>
                  <CommandInput
                    placeholder="Search packages..."
                    value={newPackageName}
                    onValueChange={setNewPackageName}
                  />
                  <ScrollArea className="h-[300px]">
                    <CommandList className="p-2">
                      {loadingSearch && (
                        <div className="py-6 text-center text-sm">
                          <Loader2 className="mx-auto mb-2 h-4 w-4 animate-spin" />
                          <p className="text-muted-foreground text-xs">
                            Searching packages...
                          </p>
                        </div>
                      )}
                      {!loadingSearch &&
                        searchResults.length === 0 &&
                        newPackageName.length < 2 && (
                          <div className="text-muted-foreground py-6 text-center text-sm">
                            <Package className="mx-auto mb-2 h-8 w-8 opacity-50" />
                            <p>Type at least 2 characters to search</p>
                            <p className="mt-1 text-xs">e.g., nodejs, python, git</p>
                          </div>
                        )}
                      {!loadingSearch &&
                        searchResults.length === 0 &&
                        newPackageName.length >= 2 && (
                          <CommandEmpty className="text-muted-foreground py-6 text-center text-sm">
                            No packages found matching &quot;{newPackageName}&quot;
                          </CommandEmpty>
                        )}
                      <CommandGroup>
                        {searchResults.map((pkg) => (
                          <CommandItem
                            key={pkg.name}
                            value={pkg.name}
                            onSelect={(value) => {
                              setNewPackageName(value)
                              setComboboxOpen(false)
                            }}
                            className="cursor-pointer"
                          >
                            <Check
                              className={cn(
                                'mr-2 h-4 w-4 flex-shrink-0',
                                newPackageName === pkg.name ? 'opacity-100' : 'opacity-0',
                              )}
                            />
                            <Package className="text-muted-foreground mr-2 h-4 w-4 flex-shrink-0" />
                            <div className="flex min-w-0 flex-1 items-center justify-between">
                              <span className="truncate font-medium">{pkg.name}</span>
                              <span className="text-muted-foreground ml-2 flex-shrink-0 text-xs">
                                {pkg.numVersions}{' '}
                                {pkg.numVersions === 1 ? 'version' : 'versions'}
                              </span>
                            </div>
                          </CommandItem>
                        ))}
                      </CommandGroup>
                    </CommandList>
                  </ScrollArea>
                </Command>
              </PopoverContent>
            </Popover>
          </div>

          {/* Version */}
          <div className="flex-1 space-y-2">
            <Label htmlFor="packageVersion" className="flex h-5 items-center">
              Version
              {loadingVersions && <Loader2 className="ml-2 inline h-3 w-3 animate-spin" />}
            </Label>
            <Select
              value={newPackageVersion}
              onValueChange={setNewPackageVersion}
              disabled={isLoading || availableVersions.length === 0}
            >
              <SelectTrigger>
                <SelectValue
                  placeholder={
                    availableVersions.length === 0 ? 'Select package first' : 'Select version'
                  }
                />
              </SelectTrigger>
              <SelectContent>
                {availableVersions.map((version) => (
                  <SelectItem key={version} value={version}>
                    {version}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Add Button */}
          <div className="flex items-end">
            <Button
              type="button"
              onClick={addPackage}
              disabled={isLoading || !newPackageName.trim() || !newPackageVersion.trim()}
            >
              <Plus className="mr-2 h-4 w-4" />
              Add
            </Button>
          </div>
        </div>
      </div>

      {/* Package List */}
      {packages.length > 0 ? (
        <div className="bg-card rounded-lg border">
          <div className="border-b p-4">
            <h3 className="text-sm font-medium">Configured Packages ({packages.length})</h3>
          </div>
          <div className="divide-y">
            {packages.map((pkg, index) => (
              <div
                key={index}
                className="flex items-center justify-between p-4 hover:bg-muted/50 transition-colors"
              >
                <div className="flex items-center gap-3 min-w-0 flex-1">
                  <div className="flex-shrink-0">
                    {pkg.status === 'installed' && (
                      <CheckCircle2 className="h-5 w-5 text-success" />
                    )}
                    {pkg.status === 'pending' && (
                      <Loader2 className="h-5 w-5 text-warning animate-spin" />
                    )}
                    {pkg.status === 'failed' && (
                      <XCircle className="h-5 w-5 text-destructive" />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="font-mono text-sm truncate">{pkg.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {pkg.displayVersion ? (
                        <span>Version: {pkg.displayVersion}</span>
                      ) : pkg.channel ? (
                        <span>Channel: {pkg.channel}</span>
                      ) : pkg.nixpkgsCommit ? (
                        <span>Commit: {pkg.nixpkgsCommit.substring(0, 8)}</span>
                      ) : (
                        <span>Latest from unstable</span>
                      )}
                    </p>
                  </div>
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => removePackage(index)}
                  disabled={isLoading}
                  className="flex-shrink-0"
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>
        </div>
      ) : (
        <div className="bg-card rounded-lg border p-12 text-center">
          <Package className="mx-auto mb-4 h-12 w-12 text-muted-foreground/50" />
          <h3 className="text-sm font-medium mb-1">No packages configured</h3>
          <p className="text-sm text-muted-foreground">
            Use the form above to add your first package
          </p>
        </div>
      )}
    </div>
  )
}
