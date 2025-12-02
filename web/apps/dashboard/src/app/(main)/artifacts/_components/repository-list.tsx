'use client'

import { Container, User } from 'lucide-react'
import type { RepositoryInfo } from '@/lib/services/registry.service'

interface RepositoryListProps {
  repositories: RepositoryInfo[]
}

export function RepositoryList({ repositories }: RepositoryListProps) {
  if (repositories.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-12 text-center">
        <Container className="mx-auto h-12 w-12 text-muted-foreground/50" />
        <h3 className="mt-4 text-lg font-medium">No container images</h3>
        <p className="mt-2 text-sm text-muted-foreground">
          Push your first image to the registry to see it here.
        </p>
        <div className="mt-4 rounded-md bg-muted p-4 text-left">
          <p className="text-xs text-muted-foreground mb-2">Example:</p>
          <code className="text-xs">
            docker push cr.yourdomain.com/username/image:tag
          </code>
        </div>
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <div className="grid gap-0 divide-y">
        {repositories.map((repo) => {
          // Parse namespace and image name from repository name
          const parts = repo.name.split('/')
          const namespace = parts.length > 1 ? parts[0] : null
          const imageName = parts.length > 1 ? parts.slice(1).join('/') : repo.name

          return (
            <div
              key={repo.name}
              className="flex items-center justify-between p-4 hover:bg-muted/50 transition-colors"
            >
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                  <Container className="h-5 w-5 text-primary" />
                </div>
                <div>
                  <div className="font-medium">{imageName}</div>
                  {namespace && (
                    <div className="flex items-center gap-1 text-sm text-muted-foreground">
                      <User className="h-3 w-3" />
                      {namespace}
                    </div>
                  )}
                </div>
              </div>
              <div className="text-sm text-muted-foreground">
                {repo.name}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
