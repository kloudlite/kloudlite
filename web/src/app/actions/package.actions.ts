'use server'

const DEVBOX_SEARCH_API = 'https://search.devbox.sh'

// API Response Types
export interface PackageVersion {
  commit_hash: string
  last_updated: number
  version: string
  platforms: string[]
  summary: string
  homepage: string
  license: string
  name: string
}

export interface PackageSearchResult {
  name: string
  num_versions: number
  versions: PackageVersion[]
}

export interface SearchResponse {
  num_results: number
  packages: PackageSearchResult[]
}

export interface FlakeRef {
  type: string
  owner: string
  repo: string
  rev: string
}

export interface FlakeInstallable {
  ref: FlakeRef
  attr_path: string
}

export interface SystemInfo {
  flake_installable: FlakeInstallable
  last_updated: string
  outputs: Array<{
    name: string
    path: string
    default?: boolean
    nar: string
  }>
}

export interface ResolveResponse {
  name: string
  version: string
  summary: string
  systems: Record<string, SystemInfo>
}

/**
 * Search for Nix packages using Devbox API
 */
export async function searchPackages(query: string): Promise<{ success: boolean; data?: SearchResponse; error?: string }> {
  try {
    if (!query.trim()) {
      return { success: false, error: 'Query is required' }
    }

    const url = `${DEVBOX_SEARCH_API}/v1/search?q=${encodeURIComponent(query)}`
    const response = await fetch(url, {
      headers: {
        'User-Agent': 'Kloudlite/1.0',
      },
      next: {
        revalidate: 3600, // Cache for 1 hour
      },
    })

    if (!response.ok) {
      return { success: false, error: `API returned ${response.status}: ${response.statusText}` }
    }

    const data: SearchResponse = await response.json()
    return { success: true, data }
  } catch (err) {
    console.error('Package search error:', err)
    const error = err instanceof Error ? err : new Error("Unknown error")
    return { success: false, error: error.message }
  }
}

/**
 * Resolve a package name and semantic version to a specific nixpkgs commit
 */
export async function resolvePackageVersion(
  name: string,
  version: string
): Promise<{ success: boolean; data?: ResolveResponse; error?: string }> {
  try {
    if (!name.trim() || !version.trim()) {
      return { success: false, error: 'Package name and version are required' }
    }

    const url = `${DEVBOX_SEARCH_API}/v2/resolve?name=${encodeURIComponent(name)}&version=${encodeURIComponent(version)}`
    const response = await fetch(url, {
      headers: {
        'User-Agent': 'Kloudlite/1.0',
      },
      next: {
        revalidate: 86400, // Cache for 24 hours (versions don't change)
      },
    })

    if (response.status === 404) {
      return { success: false, error: `Version ${version} not found for package ${name}` }
    }

    if (!response.ok) {
      return { success: false, error: `API returned ${response.status}: ${response.statusText}` }
    }

    const data: ResolveResponse = await response.json()
    return { success: true, data }
  } catch (err) {
    console.error('Package resolve error:', err)
    const error = err instanceof Error ? err : new Error("Unknown error")
    return { success: false, error: error.message }
  }
}

/**
 * Get available versions for a package
 */
export async function getPackageVersions(packageName: string): Promise<{ success: boolean; versions?: string[]; error?: string }> {
  const result = await searchPackages(packageName)

  if (!result.success || !result.data) {
    return { success: false, error: result.error }
  }

  // Find exact package match
  const pkg = result.data.packages.find(p => p.name === packageName)
  if (!pkg) {
    return { success: false, error: `Package ${packageName} not found` }
  }

  // Extract version strings
  const versions = pkg.versions.map(v => v.version)
  return { success: true, versions }
}
