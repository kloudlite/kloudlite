import { registryService } from '@/lib/services/registry.service'
import { RepositoryList } from '../_components/repository-list'

export default async function ContainerImagesPage() {
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
      <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-6 text-center">
        <p className="text-sm text-destructive">{error}</p>
        <p className="mt-2 text-xs text-muted-foreground">
          Make sure the registry is accessible and try again.
        </p>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          {repositories.length} {repositories.length === 1 ? 'repository' : 'repositories'}
        </p>
      </div>
      <RepositoryList repositories={repositories} />
    </div>
  )
}
