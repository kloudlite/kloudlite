'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Package, Plus, X, Loader2, CheckCircle2, XCircle, AlertCircle, Check, ChevronsUpDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import type { Workspace, PackageSpec } from '@/types/workspace'
import { updateWorkspace } from '@/app/actions/workspace.actions'
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
  const [searchResults, setSearchResults] = useState<Array<{ name: string; numVersions: number }>>([])
  const [loadingSearch, setLoadingSearch] = useState(false)
  const [comboboxOpen, setComboboxOpen] = useState(false)

  // Initialize packages when sheet opens
  useEffect(() => {
    if (open) {
      // Create maps for status tracking
      const installedPackagesMap = new Map(
        workspace.status?.installedPackages?.map(pkg => [pkg.name, pkg]) || []
      )
      const failedPackagesSet = new Set(workspace.status?.failedPackages || [])

      const existingPackages: PackageWithVersion[] = (workspace.spec.packages || []).map(pkg => {
        const isInstalled = installedPackagesMap.has(pkg.name)
        const isFailed = failedPackagesSet.has(pkg.name)
        const status = isInstalled ? 'installed' : isFailed ? 'failed' : 'pending'

        return {
          ...pkg,
          displayVersion: pkg.channel || (pkg.nixpkgsCommit ? `commit:${pkg.nixpkgsCommit.substring(0, 8)}` : undefined),
          status,
          installedInfo: installedPackagesMap.get(pkg.name)
        }
      })
      setPackages(existingPackages)
    }
  }, [open, workspace.spec.packages, workspace.status?.installedPackages, workspace.status?.failedPackages])

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
        const packagesInfo = result.data.packages.map(p => ({
          name: p.name,
          numVersions: p.num_versions
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
        const pkg = result.data.packages.find(p => p.name === newPackageName)
        if (pkg && pkg.versions.length > 0) {
          const versions = pkg.versions.map(v => v.version)
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
      displayVersion: newPackageVersion.trim()
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
    const packageSpecs: PackageSpec[] = packages.map(({ displayVersion, status, installedInfo, ...pkg }) => pkg)

    const result = await updateWorkspace(
      workspace.metadata.name,
      workspace.metadata.namespace,
      {
        spec: {
          ...workspace.spec,
          packages: packageSpecs,
        },
      }
    )

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

  const packageCount = packages.length
  const installedCount = packages.filter(p => p.status === 'installed').length
  const pendingCount = packages.filter(p => p.status === 'pending').length
  const failedCount = packages.filter(p => p.status === 'failed').length

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        {trigger || (
          <Button variant="outline" size="sm">
            <Package className="h-4 w-4 mr-1" />
            View Packages
          </Button>
        )}
      </SheetTrigger>
      <SheetContent className="w-full sm:max-w-2xl flex flex-col p-6">
        <SheetHeader className="mb-6">
          <SheetTitle>Workspace Packages</SheetTitle>
          <SheetDescription>
            Manage packages for this workspace using Nix package manager
          </SheetDescription>
        </SheetHeader>

        <ScrollArea className="flex-1">
          <div className="space-y-6 pr-4">

            {/* Add Package Form */}
            <div className="space-y-3 p-4 border rounded-lg bg-muted/50">
              <h4 className="text-sm font-medium">Add New Package</h4>

              <div className="flex gap-3">
                {/* Package Name */}
                <div className="flex-1 space-y-2">
                  <Label className="h-5 flex items-center">
                    Package Name *
                    {loadingSearch && <Loader2 className="inline ml-2 h-3 w-3 animate-spin" />}
                  </Label>
                  <Popover open={comboboxOpen} onOpenChange={setComboboxOpen}>
                    <PopoverTrigger asChild>
                      <Button
                        variant="outline"
                        role="combobox"
                        aria-expanded={comboboxOpen}
                        className={cn(
                          "w-full justify-between",
                          !newPackageName && "text-muted-foreground"
                        )}
                        disabled={isLoading}
                      >
                        <span className="truncate">
                          {newPackageName || "Search packages..."}
                        </span>
                        <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start">
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
                                <Loader2 className="h-4 w-4 animate-spin mx-auto mb-2" />
                                <p className="text-xs text-muted-foreground">Searching packages...</p>
                              </div>
                            )}
                            {!loadingSearch && searchResults.length === 0 && newPackageName.length < 2 && (
                              <div className="py-6 text-center text-sm text-muted-foreground">
                                <Package className="h-8 w-8 mx-auto mb-2 opacity-50" />
                                <p>Type at least 2 characters to search</p>
                                <p className="text-xs mt-1">e.g., nodejs, python, git</p>
                              </div>
                            )}
                            {!loadingSearch && searchResults.length === 0 && newPackageName.length >= 2 && (
                              <CommandEmpty className="py-6 text-center text-sm text-muted-foreground">
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
                                      "mr-2 h-4 w-4 flex-shrink-0",
                                      newPackageName === pkg.name ? "opacity-100" : "opacity-0"
                                    )}
                                  />
                                  <Package className="mr-2 h-4 w-4 flex-shrink-0 text-muted-foreground" />
                                  <div className="flex-1 flex items-center justify-between min-w-0">
                                    <span className="font-medium truncate">{pkg.name}</span>
                                    <span className="ml-2 text-xs text-muted-foreground flex-shrink-0">
                                      {pkg.numVersions} {pkg.numVersions === 1 ? 'version' : 'versions'}
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
                  <Label htmlFor="packageVersion" className="h-5 flex items-center">
                    Version *
                    {loadingVersions && <Loader2 className="inline ml-2 h-3 w-3 animate-spin" />}
                  </Label>
                  <Select
                    value={newPackageVersion}
                    onValueChange={setNewPackageVersion}
                    disabled={isLoading || availableVersions.length === 0}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={
                        availableVersions.length === 0
                          ? "Select package first"
                          : "Select version"
                      } />
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
                <Plus className="h-4 w-4 mr-2" />
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
                      className="flex items-start justify-between p-3 bg-card border rounded-lg"
                    >
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-sm truncate">{pkg.name}</span>
                          <div className="flex-shrink-0">
                            {pkg.status === 'installed' && (
                              <CheckCircle2 className="h-4 w-4 text-green-600 dark:text-green-400" />
                            )}
                            {pkg.status === 'pending' && (
                              <Loader2 className="h-4 w-4 text-yellow-600 dark:text-yellow-400 animate-spin" />
                            )}
                            {pkg.status === 'failed' && (
                              <XCircle className="h-4 w-4 text-red-600 dark:text-red-400" />
                            )}
                          </div>
                        </div>
                        <div className="text-xs text-muted-foreground mt-1">
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
              <div className="text-center py-8 text-muted-foreground">
                <Package className="h-12 w-12 mx-auto mb-3 opacity-50" />
                <p className="text-sm">No packages configured</p>
                <p className="text-xs">Use the form above to add your first package</p>
              </div>
            )}
          </div>
        </ScrollArea>

        <div className="flex gap-2 pt-6 mt-6 border-t">
          <Button variant="outline" onClick={handleCancel} disabled={isLoading} className="flex-1">
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isLoading} className="flex-1">
            {isLoading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
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
