'use client'

import { useState, useTransition, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, X, Loader2, Package, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { createWorkspace } from '@/app/actions/workspace.actions'
import { searchPackages, resolvePackageVersion } from '@/app/actions/package.actions'
import { toast } from 'sonner'
import type { PackageSpec } from '@/types/workspace'

interface PackageWithVersion extends PackageSpec {
  displayVersion?: string // Semantic version for display (e.g., "20.10.0")
}

interface CreateWorkspaceSheetProps {
  namespace: string
  user: string
}

export function CreateWorkspaceSheet({ namespace, user }: CreateWorkspaceSheetProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()

  // Basic fields
  const [name, setName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [description, setDescription] = useState('')

  // Package management
  const [packages, setPackages] = useState<PackageWithVersion[]>([])
  const [newPackageName, setNewPackageName] = useState('')
  const [newPackageVersion, setNewPackageVersion] = useState('')
  const [availableVersions, setAvailableVersions] = useState<string[]>([])
  const [loadingVersions, setLoadingVersions] = useState(false)
  const [searchResults, setSearchResults] = useState<string[]>([])
  const [loadingSearch, setLoadingSearch] = useState(false)
  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null)

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
    }, 300) // Debounce 300ms

    return () => clearTimeout(searchTimer)
  }, [newPackageName])

  // Load versions when package name is selected/confirmed
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
          setAvailableVersions(versions.slice(0, 50)) // Limit to 50 versions
        }
      }
    }

    loadVersions()
  }, [newPackageName])

  // Clean up polling interval on unmount
  useEffect(() => {
    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current)
        pollIntervalRef.current = null
      }
    }
  }, [])

  const addPackage = async () => {
    if (!newPackageName.trim()) {
      toast.error('Please enter a package name')
      return
    }

    if (!newPackageVersion.trim()) {
      toast.error('Please select a version')
      return
    }

    // Resolve semantic version to commit hash
    toast.loading('Resolving package version...')
    const result = await resolvePackageVersion(newPackageName.trim(), newPackageVersion.trim())
    toast.dismiss()

    if (!result.success || !result.data) {
      toast.error(result.error || 'Failed to resolve package version')
      return
    }

    // Extract commit hash from any system (they should all have the same commit)
    const systems = Object.values(result.data.systems)
    if (systems.length === 0) {
      toast.error('No system information available for this package version')
      return
    }

    const commitHash = systems[0].flake_installable.ref.rev
    const attrPath = systems[0].flake_installable.attr_path

    const pkg: PackageWithVersion = {
      name: attrPath, // Use attr_path (e.g., "nodejs_20") for installation
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim()) {
      toast.error('Please enter a workspace name')
      return
    }

    if (!displayName.trim()) {
      toast.error('Please enter a display name')
      return
    }

    startTransition(async () => {
      // Convert PackageWithVersion to PackageSpec (remove displayVersion)
      const packageSpecs: PackageSpec[] = packages.map(({ displayVersion, ...pkg }) => pkg)

      const result = await createWorkspace(namespace, {
        name: name.trim().toLowerCase().replace(/[^a-z0-9-]/g, '-'),
        spec: {
          displayName: displayName.trim(),
          description: description.trim() || undefined,
          owner: user,
          packages: packageSpecs.length > 0 ? packageSpecs : undefined,
          status: 'active',
        },
      })

      if (result.success) {
        toast.success('Workspace created successfully')
        setOpen(false)
        setName('')
        setDisplayName('')
        setDescription('')
        setPackages([])

        // Immediately refresh and then poll for a few seconds to catch state changes
        router.refresh()

        // Clear any existing interval before starting a new one
        if (pollIntervalRef.current) {
          clearInterval(pollIntervalRef.current)
        }

        // Poll every second for 10 seconds to catch the workspace state updates
        let pollCount = 0
        pollIntervalRef.current = setInterval(() => {
          router.refresh()
          pollCount++
          if (pollCount >= 10 && pollIntervalRef.current) {
            clearInterval(pollIntervalRef.current)
            pollIntervalRef.current = null
          }
        }, 1000)
      } else {
        toast.error(result.error || 'Failed to create workspace')
      }
    })
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button size="sm" className="gap-2">
          <Plus className="h-4 w-4" />
          New Workspace
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-full sm:max-w-2xl">
        <form onSubmit={handleSubmit} className="flex h-full flex-col">
          <SheetHeader>
            <SheetTitle>Create Workspace</SheetTitle>
            <SheetDescription>
              Create a new development workspace with optional package installations
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 space-y-6 overflow-y-auto p-4">
          {/* Basic Information */}
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="displayName">Display Name *</Label>
              <Input
                id="displayName"
                placeholder="My Workspace"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                disabled={isPending}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="name">Resource Name *</Label>
              <Input
                id="name"
                placeholder="my-workspace"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isPending}
                className="font-mono text-sm"
              />
              <p className="text-xs text-muted-foreground">
                Lowercase letters, numbers, and hyphens only
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="Describe your workspace..."
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={isPending}
                rows={3}
              />
            </div>
          </div>

          {/* Packages Section */}
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <Package className="h-4 w-4" />
              <Label>Nix Packages</Label>
            </div>
            <p className="text-sm text-muted-foreground">
              Add packages to install in your workspace using Nix package manager
            </p>

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
                      disabled={isPending}
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
                    disabled={isPending}
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
                  disabled={isPending || availableVersions.length === 0}
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
                disabled={isPending || !newPackageName.trim() || !newPackageVersion.trim()}
                className="w-full"
              >
                <Plus className="h-4 w-4 mr-2" />
                Add Package
              </Button>
            </div>
          </div>
        </div>

        <SheetFooter className="p-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Workspace
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
