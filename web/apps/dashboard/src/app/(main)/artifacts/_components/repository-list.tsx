'use client'

import { useState } from 'react'
import { Container, User, ChevronRight, ChevronDown, Tag, Trash2, Loader2, Copy, Check } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@kloudlite/ui'
import type { RepositoryInfo } from '@/lib/services/registry.service'
import { listTags, deleteTag, deleteRepository } from '@/app/actions/registry.actions'

interface RepositoryListProps {
  repositories: RepositoryInfo[]
  onRepositoryDeleted?: (repoName: string) => void
}

interface TagsState {
  loading: boolean
  tags: string[]
  error: string | null
}

export function RepositoryList({ repositories, onRepositoryDeleted }: RepositoryListProps) {
  const [expandedRepo, setExpandedRepo] = useState<string | null>(null)
  const [tagsState, setTagsState] = useState<TagsState>({ loading: false, tags: [], error: null })
  const [deletingTag, setDeletingTag] = useState<string | null>(null)
  const [tagToDelete, setTagToDelete] = useState<{ repo: string; tag: string } | null>(null)
  const [repoToDelete, setRepoToDelete] = useState<string | null>(null)
  const [deletingRepo, setDeletingRepo] = useState<string | null>(null)
  const [copiedTag, setCopiedTag] = useState<string | null>(null)

  const handleRepoClick = async (repoName: string) => {
    if (expandedRepo === repoName) {
      setExpandedRepo(null)
      return
    }

    setExpandedRepo(repoName)
    setTagsState({ loading: true, tags: [], error: null })

    const result = await listTags(repoName)
    if (result.success) {
      setTagsState({ loading: false, tags: result.data.tags || [], error: null })
    } else {
      setTagsState({
        loading: false,
        tags: [],
        error: result.error || 'Failed to load tags',
      })
    }
  }

  const handleDeleteTag = async () => {
    if (!tagToDelete) return

    setDeletingTag(tagToDelete.tag)
    const result = await deleteTag(tagToDelete.repo, tagToDelete.tag)
    if (result.success) {
      // Remove the deleted tag from the list
      setTagsState((prev) => ({
        ...prev,
        tags: prev.tags.filter((t) => t !== tagToDelete.tag),
      }))
    } else {
      console.error('Failed to delete tag:', result.error)
    }
    setDeletingTag(null)
    setTagToDelete(null)
  }

  const handleCopyTag = async (repoName: string, tag: string) => {
    const fullImageName = `cr.beanbag.khost.dev/${repoName}:${tag}`
    await navigator.clipboard.writeText(fullImageName)
    setCopiedTag(`${repoName}:${tag}`)
    setTimeout(() => setCopiedTag(null), 2000)
  }

  const handleDeleteRepository = async () => {
    if (!repoToDelete) return

    setDeletingRepo(repoToDelete)
    const result = await deleteRepository(repoToDelete)
    if (result.success) {
      onRepositoryDeleted?.(repoToDelete)
      if (expandedRepo === repoToDelete) {
        setExpandedRepo(null)
      }
    } else {
      console.error('Failed to delete repository:', result.error)
    }
    setDeletingRepo(null)
    setRepoToDelete(null)
  }

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
    <>
      <div className="rounded-lg border">
        <div className="grid gap-0 divide-y">
          {repositories.map((repo) => {
            const parts = repo.name.split('/')
            const namespace = parts.length > 1 ? parts[0] : null
            const imageName = parts.length > 1 ? parts.slice(1).join('/') : repo.name
            const isExpanded = expandedRepo === repo.name

            return (
              <div key={repo.name}>
                <button
                  onClick={() => handleRepoClick(repo.name)}
                  className="flex w-full items-center justify-between p-4 hover:bg-muted/50 transition-colors text-left"
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
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">{repo.name}</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 w-7 p-0 text-destructive hover:text-destructive hover:bg-destructive/10"
                      onClick={(e) => {
                        e.stopPropagation()
                        setRepoToDelete(repo.name)
                      }}
                      disabled={deletingRepo === repo.name}
                      title="Delete repository"
                    >
                      {deletingRepo === repo.name ? (
                        <Loader2 className="h-3 w-3 animate-spin" />
                      ) : (
                        <Trash2 className="h-3 w-3" />
                      )}
                    </Button>
                    {isExpanded ? (
                      <ChevronDown className="h-4 w-4 text-muted-foreground" />
                    ) : (
                      <ChevronRight className="h-4 w-4 text-muted-foreground" />
                    )}
                  </div>
                </button>

                {isExpanded && (
                  <div className="border-t bg-muted/30 px-4 py-3">
                    {tagsState.loading ? (
                      <div className="flex items-center gap-2 text-sm text-muted-foreground py-2">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Loading tags...
                      </div>
                    ) : tagsState.error ? (
                      <div className="text-sm text-destructive py-2">{tagsState.error}</div>
                    ) : tagsState.tags.length === 0 ? (
                      <div className="text-sm text-muted-foreground py-2">No tags found</div>
                    ) : (
                      <div className="space-y-1">
                        <div className="text-xs text-muted-foreground mb-2 flex items-center gap-1">
                          <Tag className="h-3 w-3" />
                          {tagsState.tags.length} {tagsState.tags.length === 1 ? 'tag' : 'tags'}
                        </div>
                        <div className="grid gap-1">
                          {tagsState.tags.map((tag) => (
                            <div
                              key={tag}
                              className="flex items-center justify-between rounded-md bg-background px-3 py-2 text-sm"
                            >
                              <div className="flex items-center gap-2">
                                <Tag className="h-3 w-3 text-muted-foreground" />
                                <code className="text-xs">{tag}</code>
                              </div>
                              <div className="flex items-center gap-1">
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="h-7 w-7 p-0"
                                  onClick={(e) => {
                                    e.stopPropagation()
                                    handleCopyTag(repo.name, tag)
                                  }}
                                  title="Copy full image name"
                                >
                                  {copiedTag === `${repo.name}:${tag}` ? (
                                    <Check className="h-3 w-3 text-green-500" />
                                  ) : (
                                    <Copy className="h-3 w-3" />
                                  )}
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="h-7 w-7 p-0 text-destructive hover:text-destructive hover:bg-destructive/10"
                                  onClick={(e) => {
                                    e.stopPropagation()
                                    setTagToDelete({ repo: repo.name, tag })
                                  }}
                                  disabled={deletingTag === tag}
                                  title="Delete tag"
                                >
                                  {deletingTag === tag ? (
                                    <Loader2 className="h-3 w-3 animate-spin" />
                                  ) : (
                                    <Trash2 className="h-3 w-3" />
                                  )}
                                </Button>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      <AlertDialog open={!!tagToDelete} onOpenChange={() => setTagToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete tag?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete tag <code className="bg-muted px-1 rounded">{tagToDelete?.tag}</code> from{' '}
              <code className="bg-muted px-1 rounded">{tagToDelete?.repo}</code>? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteTag}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={!!repoToDelete} onOpenChange={() => setRepoToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete repository?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete repository <code className="bg-muted px-1 rounded">{repoToDelete}</code>?
              This will delete all tags and cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteRepository}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete Repository
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
