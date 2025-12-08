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
} from 'lucide-react'
import { Button } from '@kloudlite/ui'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@kloudlite/ui'
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
import { updateWorkspace, getPackageRequest } from '@/app/actions/workspace.actions'
import { searchPackages, resolvePackageVersion } from '@/app/actions/package.actions'
import { toast } from 'sonner'

interface PackageWithVersion extends PackageSpec {
  displayVersion?: string
  status?: 'installed' | 'pending' | 'failed'
  installedInfo?: {
    name: string
    version?: string
  }
}

interface PackagesSheetProps {
  workspace: Workspace
  trigger?: React.ReactNode
}

export function PackagesSheet({ workspace, trigger }: PackagesSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [packages, setPackages] = useState<PackageWithVersion[]>([])
  const [isLoading, setIsLoading] = useState(false)

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

  // Initialize packages when sheet opens - fetch status from PackageRequest
  useEffect(() => {
    if (open) {
      const loadPackageStatus = async () => {
        // Fetch package status from PackageRequest (source of truth)
        const pkgReqResult = await getPackageRequest(workspace.metadata.name, workspace.metadata.namespace)
        const pkgReq: PackageRequest | null = pkgReqResult.success ? pkgReqResult.data : null

        // Create maps for status tracking from PackageRequest
        const installedPackagesMap = new Map(
          pkgReq?.status?.installedPackages?.map((pkg) => [pkg.name, pkg]) || [],
        )
        const failedPackagesSet = new Set(pkgReq?.status?.failedPackages || [])

        const existingPackages: PackageWithVersion[] = (workspace.spec.packages || []).map((pkg) => {
          const isInstalled = installedPackagesMap.has(pkg.name)
          const isFailed = failedPackagesSet.has(pkg.name)
          const status = isInstalled ? 'installed' : isFailed ? 'failed' : 'pending'

          return {
            ...pkg,
            displayVersion:
              pkg.channel ||
              (pkg.nixpkgsCommit ? `commit:${pkg.nixpkgsCommit.substring(0, 8)}` : undefined),
            status,
            installedInfo: installedPackagesMap.get(pkg.name),
          }
        })
        setPackages(existingPackages)
      }

      loadPackageStatus()
    }
  }, [open, workspace.metadata.name, workspace.metadata.namespace, workspace.spec.packages])

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
    }

    setPackages([...packages, pkg])
    setNewPackageName('')
    setNewPackageVersion('')
    setAvailableVersions([])
    setSearchResults([])
    toast.success('Package added')
  }

  const removePackage = (index: number) => {
    setPackages(packages.filter((_, i) => i !== index))
  }

  const handleSave = async () => {
    setIsLoading(true)

    // Convert PackageWithVersion to PackageSpec
    const packageSpecs: PackageSpec[] = packages.map(
      ({
        displayVersion: _displayVersion,
        status: _status,
        installedInfo: _installedInfo,
        ...pkg
      }) => pkg,
    )

    const result = await updateWorkspace(workspace.metadata.name, workspace.metadata.namespace, {
      spec: {
        ...workspace.spec,
        packages: packageSpecs,
      },
    })

    if (result.success) {
      toast.success('Packages updated successfully')
      setOpen(false)
      router.refresh()
    } else {
      toast.error(result.error || 'Failed to update packages')
    }

    setIsLoading(false)
  }

  const handleCancel = () => {
    setNewPackageName('')
    setNewPackageVersion('')
    setAvailableVersions([])
    setSearchResults([])
    setOpen(false)
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        {trigger || (
          <Button variant="outline" size="sm">
            <Package className="mr-1 h-4 w-4" />
            View Packages
          </Button>
        )}
      </SheetTrigger>
      <SheetContent className="flex w-full flex-col p-6 sm:max-w-2xl">
        <SheetHeader className="mb-6">
          <SheetTitle>Workspace Packages</SheetTitle>
          <SheetDescription>
            Manage packages for this workspace using Nix package manager
          </SheetDescription>
        </SheetHeader>

        <ScrollArea className="flex-1">
          <div className="space-y-6 pr-4">
            {/* Add Package Form */}
            <div className="bg-muted/50 space-y-3 rounded-lg border p-4">
              <h4 className="text-sm font-medium">Add New Package</h4>

              <div className="flex gap-3">
                {/* Package Name */}
                <div className="flex-1 space-y-2">
                  <Label className="flex h-5 items-center">
                    Package Name *
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
                    Version *
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
              </div>

              <Button
                type="button"
                variant="default"
                size="sm"
                onClick={addPackage}
                disabled={isLoading || !newPackageName.trim() || !newPackageVersion.trim()}
                className="w-full"
              >
                <Plus className="mr-2 h-4 w-4" />
                Add Package
              </Button>
            </div>

            {/* Package List */}
            {packages.length > 0 && (
              <div className="space-y-2">
                <h4 className="text-sm font-medium">Configured Packages ({packages.length})</h4>
                <div className="space-y-2">
                  {packages.map((pkg, index) => (
                    <div
                      key={index}
                      className="bg-card flex items-start justify-between rounded-lg border p-3"
                    >
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <span className="truncate font-mono text-sm">{pkg.name}</span>
                          <div className="flex-shrink-0">
                            {pkg.status === 'installed' && (
                              <CheckCircle2 className="text-success h-4 w-4" />
                            )}
                            {pkg.status === 'pending' && (
                              <Loader2 className="text-warning h-4 w-4 animate-spin" />
                            )}
                            {pkg.status === 'failed' && (
                              <XCircle className="text-destructive h-4 w-4" />
                            )}
                          </div>
                        </div>
                        <div className="text-muted-foreground mt-1 text-xs">
                          {pkg.installedInfo?.version ? (
                            <span>Version: {pkg.installedInfo.version}</span>
                          ) : pkg.displayVersion ? (
                            <span>Version: {pkg.displayVersion}</span>
                          ) : pkg.channel ? (
                            <span>Channel: {pkg.channel}</span>
                          ) : pkg.nixpkgsCommit ? (
                            <span>Commit: {pkg.nixpkgsCommit.substring(0, 8)}</span>
                          ) : (
                            <span>Latest from unstable</span>
                          )}
                        </div>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => removePackage(index)}
                        disabled={isLoading}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {packages.length === 0 && (
              <div className="text-muted-foreground py-8 text-center">
                <Package className="mx-auto mb-3 h-12 w-12 opacity-50" />
                <p className="text-sm">No packages configured</p>
                <p className="text-xs">Use the form above to add your first package</p>
              </div>
            )}
          </div>
        </ScrollArea>

        <div className="mt-6 flex gap-2 border-t pt-6">
          <Button variant="outline" onClick={handleCancel} disabled={isLoading} className="flex-1">
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isLoading} className="flex-1">
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              'Save Changes'
            )}
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  )
}
