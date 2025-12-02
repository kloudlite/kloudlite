import { AlertCircle } from 'lucide-react'
import { registryService } from '@/lib/services/registry.service'
import { RepositoryList } from '../_components/repository-list'

export default async function ContainerReposPage() {
  let repositories: { name: string }[] = []
  let error: string | null = null

  try {
    const response = await registryService.listRepositories()
    repositories = response.repositories || []
  } catch (err) {
    console.error('Failed to fetch repositories:', err)
    error = err instanceof Error ? err.message : 'Failed to load repositories'
  }

  if (error) {
    return (
      <div className="bg-card rounded-lg border border-destructive/50 p-8 text-center">
        <AlertCircle className="mx-auto h-10 w-10 text-destructive/70" />
        <h3 className="mt-4 text-lg font-medium text-destructive">Failed to load repositories</h3>
        <p className="mt-2 text-sm text-muted-foreground">{error}</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Make sure the registry is accessible and try again.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Header with count */}
      <div className="flex items-center justify-between">
        <span className="text-muted-foreground text-sm">
          {repositories.length} {repositories.length === 1 ? 'repository' : 'repositories'}
        </span>
      </div>

      {/* Repository List */}
      <RepositoryList repositories={repositories} />
    </div>
  )
}
