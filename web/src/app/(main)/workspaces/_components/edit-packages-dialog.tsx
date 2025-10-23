'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Package, Plus, X, Loader2, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { Workspace, PackageSpec } from '@/types/workspace'
import { updateWorkspace } from '@/app/actions/workspace.actions'
import { searchPackages, resolvePackageVersion } from '@/app/actions/package.actions'
import { toast } from 'sonner'

interface PackageWithVersion extends PackageSpec {
  displayVersion?: string // Semantic version for display (e.g., "20.10.0")
}

interface EditPackagesDialogProps {
  workspace: Workspace
}

export function EditPackagesDialog({ workspace }: EditPackagesDialogProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [packages, setPackages] = useState<PackageWithVersion[]>([])
  const [isLoading, setIsLoading] = useState(false)

  // Package search state
  const [newPackageName, setNewPackageName] = useState('')
  const [newPackageVersion, setNewPackageVersion] = useState('')
  const [availableVersions, setAvailableVersions] = useState<string[]>([])
  const [loadingVersions, setLoadingVersions] = useState(false)
  const [searchResults, setSearchResults] = useState<string[]>([])
  const [loadingSearch, setLoadingSearch] = useState(false)

  // Initialize packages when dialog opens
  useEffect(() => {
    if (open) {
      const existingPackages: PackageWithVersion[] = (workspace.spec.packages || []).map(pkg => ({
        ...pkg,
        displayVersion: pkg.channel || (pkg.nixpkgsCommit ? `commit:${pkg.nixpkgsCommit.substring(0, 8)}` : undefined)
      }))
      setPackages(existingPackages)
    }
  }, [open, workspace.spec.packages])

  // Search for packages as user types
  useEffect(() => {
    const searchTimer = setTimeout(async () => {
      if (newPackageName.length < 2) {
        setSearchResults([])
        return
      }

      setLoadingSearch(true)
      const result = await searchPackages(newPackageName)
      setLoadingSearch(false)

      if (result.success && result.data) {
        const packageNames = result.data.packages.map(p => p.name)
        setSearchResults(packageNames.slice(0, 10))
      }
    }, 300)

    return () => clearTimeout(searchTimer)
  }, [newPackageName])

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

      if (result.success && result.data) {
        const pkg = result.data.packages.find(p => p.name === newPackageName)
        if (pkg && pkg.versions.length > 0) {
          const versions = pkg.versions.map(v => v.version)
          setAvailableVersions(versions.slice(0, 50))
        }
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
    const packageSpecs: PackageSpec[] = packages.map(({ displayVersion: _displayVersion, ...pkg }) => pkg)

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

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">
          <Package className="h-4 w-4 mr-1" />
          Edit Packages
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>Edit Workspace Packages</DialogTitle>
          <DialogDescription>
            Add or remove packages for this workspace using Nix package manager
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto space-y-4 py-4">
          {/* Package List */}
          {packages.length > 0 && (
            <div className="space-y-2">
              {packages.map((pkg, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between p-3 bg-muted rounded-lg"
                >
                  <div className="flex-1">
                    <div className="font-medium">{pkg.name}</div>
                    <div className="text-xs text-muted-foreground mt-1">
                      {pkg.displayVersion ? (
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
          )}

          {/* Add Package Form */}
          <div className="space-y-3 p-4 border rounded-lg">
            <div className="space-y-2">
              <Label htmlFor="packageName">
                Package Name *
                {loadingSearch && <Loader2 className="inline ml-2 h-3 w-3 animate-spin" />}
              </Label>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="packageName"
                  placeholder="Search packages (e.g., nodejs, vim, git)"
                  value={newPackageName}
                  onChange={(e) => setNewPackageName(e.target.value)}
                  disabled={isLoading}
                  className="pl-9"
                />
              </div>
              {searchResults.length > 0 && newPackageName.length >= 2 && (
                <div className="border rounded-md bg-background shadow-lg max-h-40 overflow-y-auto">
                  {searchResults.map((pkgName) => (
                    <button
                      key={pkgName}
                      type="button"
                      onClick={() => {
                        setNewPackageName(pkgName)
                        setSearchResults([])
                      }}
                      className="w-full text-left px-3 py-2 hover:bg-muted text-sm"
                    >
                      {pkgName}
                    </button>
                  ))}
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="packageVersion">
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
                      ? "Select a package first"
                      : "Select a version"
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
              <p className="text-xs text-muted-foreground">
                Semantic version will be resolved to exact nixpkgs commit
              </p>
            </div>

            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={addPackage}
              disabled={isLoading || !newPackageName.trim() || !newPackageVersion.trim()}
              className="w-full"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add Package
            </Button>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleCancel} disabled={isLoading}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isLoading}>
            {isLoading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Saving...
              </>
            ) : (
              'Save Changes'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
