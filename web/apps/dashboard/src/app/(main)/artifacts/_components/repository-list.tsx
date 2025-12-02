'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import {
  Container,
  User,
  Tag,
  Trash2,
  Loader2,
  Copy,
  Check,
  MoreHorizontal,
  ChevronDown,
  ChevronRight,
  Package,
} from 'lucide-react'
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@kloudlite/ui'
import { toast } from 'sonner'
import type { RepositoryInfo } from '@/lib/services/registry.service'
import { listTags, deleteTag, deleteRepository } from '@/app/actions/registry.actions'

interface RepositoryListProps {
  repositories: RepositoryInfo[]
}

interface TagsState {
  loading: boolean
  tags: string[]
  error: string | null
}

export function RepositoryList({ repositories: initialRepositories }: RepositoryListProps) {
  const [repositories, setRepositories] = useState(initialRepositories)
  const [expandedRepo, setExpandedRepo] = useState<string | null>(null)
  const [tagsState, setTagsState] = useState<TagsState>({ loading: false, tags: [], error: null })
  const [deletingTag, setDeletingTag] = useState<string | null>(null)
  const [tagToDelete, setTagToDelete] = useState<{ repo: string; tag: string } | null>(null)
  const [repoToDelete, setRepoToDelete] = useState<string | null>(null)
  const [deletingRepo, setDeletingRepo] = useState<string | null>(null)
  const [copiedImage, setCopiedImage] = useState<string | null>(null)
  const [, startTransition] = useTransition()
  const router = useRouter()

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
      // If repository not found, it may have been deleted - remove it from the list
      if (result.error?.includes('not found') || result.error?.includes('404')) {
        setRepositories((prev) => prev.filter((r) => r.name !== repoName))
        setExpandedRepo(null)
        toast.info('Repository no longer exists', {
          description: 'This repository has been removed from the list',
        })
      } else {
        setTagsState({
          loading: false,
          tags: [],
          error: result.error || 'Failed to load tags',
        })
      }
    }
  }

  const handleDeleteTag = async () => {
    if (!tagToDelete) return

    setDeletingTag(tagToDelete.tag)
    const result = await deleteTag(tagToDelete.repo, tagToDelete.tag)
    if (result.success) {
      toast.success('Tag deleted', {
        description: `${tagToDelete.tag} has been deleted from ${tagToDelete.repo}`,
      })
      setTagsState((prev) => ({
        ...prev,
        tags: prev.tags.filter((t) => t !== tagToDelete.tag),
      }))
    } else {
      toast.error('Failed to delete tag', {
        description: result.error || 'An error occurred',
      })
    }
    setDeletingTag(null)
    setTagToDelete(null)
  }

  const handleDeleteRepository = async () => {
    if (!repoToDelete) return

    setDeletingRepo(repoToDelete)
    const result = await deleteRepository(repoToDelete)
    if (result.success) {
      toast.success('Repository deleted', {
        description: `${repoToDelete} has been deleted`,
      })
      setRepositories((prev) => prev.filter((r) => r.name !== repoToDelete))
      if (expandedRepo === repoToDelete) {
        setExpandedRepo(null)
      }
      startTransition(() => {
        router.refresh()
      })
    } else {
      toast.error('Failed to delete repository', {
        description: result.error || 'An error occurred',
      })
    }
    setDeletingRepo(null)
    setRepoToDelete(null)
  }

  const handleCopyImage = async (repoName: string, tag?: string) => {
    const fullImageName = tag
      ? `cr.beanbag.khost.dev/${repoName}:${tag}`
      : `cr.beanbag.khost.dev/${repoName}`
    await navigator.clipboard.writeText(fullImageName)
    setCopiedImage(tag ? `${repoName}:${tag}` : repoName)
    toast.success('Copied to clipboard', {
      description: fullImageName,
    })
    setTimeout(() => setCopiedImage(null), 2000)
  }

  if (repositories.length === 0) {
    return (
      <div className="bg-card rounded-lg border py-12 text-center">
        <Package className="mx-auto h-12 w-12 text-muted-foreground/50" />
        <h3 className="mt-4 text-lg font-medium">No container repos</h3>
        <p className="mt-2 text-sm text-muted-foreground">
          Container images can only be pushed from your workspaces.
        </p>
      </div>
    )
  }

  return (
    <>
      {/* Table */}
      <div className="bg-card overflow-hidden rounded-lg border">
        <table className="min-w-full">
          <thead className="bg-muted/50 border-b">
            <tr>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Repository
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Owner
              </th>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                Tags
              </th>
              <th className="text-muted-foreground px-6 py-3 text-right text-xs font-medium tracking-wider uppercase">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {repositories.map((repo) => {
              const parts = repo.name.split('/')
              const namespace = parts.length > 1 ? parts[0] : 'library'
              const imageName = parts.length > 1 ? parts.slice(1).join('/') : repo.name
              const isExpanded = expandedRepo === repo.name
              const isDeleting = deletingRepo === repo.name

              return (
                <>
                  <tr key={repo.name} className={`hover:bg-muted/50 ${isDeleting ? 'opacity-50' : ''}`}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <button
                        onClick={() => handleRepoClick(repo.name)}
                        className="flex items-center gap-3 hover:text-primary transition-colors"
                        disabled={isDeleting}
                      >
                        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10">
                          <Container className="h-4 w-4 text-primary" />
                        </div>
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{repo.name}</span>
                          {isExpanded ? (
                            <ChevronDown className="h-4 w-4 text-muted-foreground" />
                          ) : (
                            <ChevronRight className="h-4 w-4 text-muted-foreground" />
                          )}
                        </div>
                      </button>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                        <User className="h-3.5 w-3.5" />
                        {namespace}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {isExpanded ? (
                        tagsState.loading ? (
                          <span className="text-sm text-muted-foreground">Loading...</span>
                        ) : tagsState.error ? (
                          <span className="text-sm text-destructive">Error</span>
                        ) : (
                          <span className="inline-flex items-center gap-1 rounded-full bg-secondary px-2 py-0.5 text-xs font-medium">
                            <Tag className="h-3 w-3" />
                            {tagsState.tags.length}
                          </span>
                        )
                      ) : (
                        <button
                          onClick={() => handleRepoClick(repo.name)}
                          className="text-sm text-muted-foreground hover:text-foreground"
                        >
                          Click to view
                        </button>
                      )}
                    </td>
                    <td className="px-6 py-4 text-right whitespace-nowrap">
                      {isDeleting ? (
                        <span className="text-muted-foreground text-xs flex items-center justify-end gap-2">
                          <Loader2 className="h-3 w-3 animate-spin" />
                          Deleting...
                        </span>
                      ) : (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleRepoClick(repo.name)}>
                              {isExpanded ? 'Hide Tags' : 'View Tags'}
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleCopyImage(repo.name)}>
                              <Copy className="mr-2 h-4 w-4" />
                              Copy Image URL
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className="text-destructive"
                              onClick={() => setRepoToDelete(repo.name)}
                            >
                              <Trash2 className="mr-2 h-4 w-4" />
                              Delete Repository
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      )}
                    </td>
                  </tr>

                  {/* Expanded Tags Row */}
                  {isExpanded && (
                    <tr key={`${repo.name}-tags`}>
                      <td colSpan={4} className="bg-muted/30 px-6 py-4">
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
                          <div className="grid gap-2">
                            {tagsState.tags.map((tag) => (
                              <div
                                key={tag}
                                className="flex items-center justify-between rounded-md bg-background px-4 py-2.5 border"
                              >
                                <div className="flex items-center gap-3">
                                  <Tag className="h-4 w-4 text-muted-foreground" />
                                  <div>
                                    <code className="text-sm font-medium">{tag}</code>
                                    <p className="text-xs text-muted-foreground mt-0.5">
                                      cr.beanbag.khost.dev/{repo.name}:{tag}
                                    </p>
                                  </div>
                                </div>
                                <div className="flex items-center gap-2">
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-8 px-2"
                                    onClick={() => handleCopyImage(repo.name, tag)}
                                  >
                                    {copiedImage === `${repo.name}:${tag}` ? (
                                      <Check className="h-4 w-4 text-green-500" />
                                    ) : (
                                      <Copy className="h-4 w-4" />
                                    )}
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-8 px-2 text-destructive hover:text-destructive hover:bg-destructive/10"
                                    onClick={() => setTagToDelete({ repo: repo.name, tag })}
                                    disabled={deletingTag === tag}
                                  >
                                    {deletingTag === tag ? (
                                      <Loader2 className="h-4 w-4 animate-spin" />
                                    ) : (
                                      <Trash2 className="h-4 w-4" />
                                    )}
                                  </Button>
                                </div>
                              </div>
                            ))}
                          </div>
                        )}
                      </td>
                    </tr>
                  )}
                </>
              )
            })}
          </tbody>
        </table>
      </div>

      {/* Delete Tag Dialog */}
      <AlertDialog open={!!tagToDelete} onOpenChange={() => setTagToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete tag?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete tag{' '}
              <code className="bg-muted px-1.5 py-0.5 rounded text-sm">{tagToDelete?.tag}</code> from{' '}
              <code className="bg-muted px-1.5 py-0.5 rounded text-sm">{tagToDelete?.repo}</code>?
              <br />
              <span className="text-destructive mt-2 block">This action cannot be undone.</span>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteTag}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete Tag
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Delete Repository Dialog */}
      <AlertDialog open={!!repoToDelete} onOpenChange={() => setRepoToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete repository?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete repository{' '}
              <code className="bg-muted px-1.5 py-0.5 rounded text-sm">{repoToDelete}</code>?
              <br />
              <span className="text-destructive mt-2 block">
                This will delete all tags and cannot be undone.
              </span>
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
