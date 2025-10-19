'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Plus, Copy, Trash2, Key, Eye, EyeOff, Check } from 'lucide-react'
import { toast } from 'sonner'
import { createConnectionToken, deleteConnectionToken } from '@/app/actions/connection-token.actions'
import type { ConnectionToken } from '@/lib/services/connection-token.service'

interface ConnectionTokensListProps {
  tokens: ConnectionToken[]
}

export function ConnectionTokensList({ tokens: initialTokens }: ConnectionTokensListProps) {
  const [tokens, setTokens] = useState(initialTokens)
  const [newTokenName, setNewTokenName] = useState('')
  const [newlyCreatedJWT, setNewlyCreatedJWT] = useState<string | null>(null)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [showTokenDialogOpen, setShowTokenDialogOpen] = useState(false)
  const [visibleTokens, setVisibleTokens] = useState<Set<string>>(new Set())
  const [copiedToken, setCopiedToken] = useState<string | null>(null)
  const [isCreating, setIsCreating] = useState(false)
  const [deletingToken, setDeletingToken] = useState<string | null>(null)

  const handleCreateToken = async () => {
    if (!newTokenName.trim()) return

    setIsCreating(true)
    const result = await createConnectionToken({ displayName: newTokenName.trim() })
    setIsCreating(false)

    if (result.success && result.data) {
      setTokens([...tokens, result.data.token])
      setNewlyCreatedJWT(result.data.jwt)
      setNewTokenName('')
      setCreateDialogOpen(false)
      setShowTokenDialogOpen(true)
      toast.success('Connection token created successfully')
    } else {
      toast.error(result.error || 'Failed to create token')
    }
  }

  const handleDeleteToken = async (name: string) => {
    setDeletingToken(name)
    const result = await deleteConnectionToken(name)
    setDeletingToken(null)

    if (result.success) {
      setTokens(tokens.filter(t => t.metadata.name !== name))
      toast.success('Connection token deleted')
    } else {
      toast.error(result.error || 'Failed to delete token')
    }
  }

  const handleCopyToken = async (token: string) => {
    await navigator.clipboard.writeText(token)
    setCopiedToken(token)
    toast.success('Token copied to clipboard')
    setTimeout(() => setCopiedToken(null), 2000)
  }

  const toggleTokenVisibility = (name: string) => {
    const newVisible = new Set(visibleTokens)
    if (newVisible.has(name)) {
      newVisible.delete(name)
    } else {
      newVisible.add(name)
    }
    setVisibleTokens(newVisible)
  }

  const maskToken = (token: string) => {
    if (!token || token.length < 16) return token
    return token.substring(0, 8) + '...' + token.substring(token.length - 8)
  }

  return (
    <>
      <div className="flex justify-end mb-4">
        <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Create Token
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Connection Token</DialogTitle>
              <DialogDescription>
                Create a new token to access Kloudlite workspaces from external tools like VS Code, CLI, etc.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="token-name">Token Name</Label>
                <Input
                  id="token-name"
                  placeholder="e.g., VS Code Extension, CI/CD Pipeline"
                  value={newTokenName}
                  onChange={(e) => setNewTokenName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && newTokenName.trim()) {
                      handleCreateToken()
                    }
                  }}
                />
                <p className="text-sm text-muted-foreground">
                  Give your token a descriptive name to remember where it's used
                </p>
              </div>
            </div>
            <DialogFooter>
              <Button onClick={handleCreateToken} disabled={!newTokenName.trim() || isCreating}>
                {isCreating ? 'Creating...' : 'Create Token'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Token Display Dialog */}
        <Dialog open={showTokenDialogOpen} onOpenChange={setShowTokenDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Token Created Successfully!</DialogTitle>
              <DialogDescription>
                Make sure to copy your token now. You won't be able to see it again!
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="flex items-center gap-2">
                <Input
                  readOnly
                  value={newlyCreatedJWT || ''}
                  className="font-mono text-sm"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => newlyCreatedJWT && handleCopyToken(newlyCreatedJWT)}
                >
                  {copiedToken === newlyCreatedJWT ? (
                    <Check className="h-4 w-4" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
            <DialogFooter>
              <Button onClick={() => {
                setShowTokenDialogOpen(false)
                setNewlyCreatedJWT(null)
              }}>
                Done
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Key className="h-5 w-5" />
            Active Tokens
          </CardTitle>
          <CardDescription>
            Tokens allow external applications to authenticate with your Kloudlite account
          </CardDescription>
        </CardHeader>
        <CardContent>
          {tokens.length === 0 ? (
            <div className="text-center py-12">
              <Key className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold mb-2">No tokens yet</h3>
              <p className="text-muted-foreground mb-4">
                Create your first connection token to start using external tools
              </p>
            </div>
          ) : (
            <table className="min-w-full">
              <thead className="bg-muted/50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Name</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Token</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Created</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Last Used</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {tokens.map((token) => {
                  const displayToken = token.status?.token || 'Hidden'
                  const isVisible = visibleTokens.has(token.metadata.name)

                  return (
                    <tr key={token.metadata.name} className="hover:bg-muted/50">
                      <td className="px-6 py-4 whitespace-nowrap font-medium">{token.spec.displayName}</td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2">
                          <code className="text-sm font-mono bg-muted px-2 py-1 rounded">
                            {isVisible ? displayToken : maskToken(displayToken)}
                          </code>
                          {displayToken !== 'Hidden' && (
                            <>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8"
                                onClick={() => toggleTokenVisibility(token.metadata.name)}
                              >
                                {isVisible ? (
                                  <EyeOff className="h-4 w-4" />
                                ) : (
                                  <Eye className="h-4 w-4" />
                                )}
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8"
                                onClick={() => handleCopyToken(displayToken)}
                              >
                                {copiedToken === displayToken ? (
                                  <Check className="h-4 w-4" />
                                ) : (
                                  <Copy className="h-4 w-4" />
                                )}
                              </Button>
                            </>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        {token.metadata.creationTimestamp
                          ? new Date(token.metadata.creationTimestamp).toLocaleDateString('en-US', {
                              year: 'numeric',
                              month: 'short',
                              day: 'numeric'
                            })
                          : 'N/A'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        {token.status?.lastUsed
                          ? new Date(token.status.lastUsed).toLocaleDateString('en-US', {
                              year: 'numeric',
                              month: 'short',
                              day: 'numeric'
                            })
                          : 'Never'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right">
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8"
                              disabled={deletingToken === token.metadata.name}
                            >
                              <Trash2 className="h-4 w-4 text-destructive" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Delete Token?</AlertDialogTitle>
                              <AlertDialogDescription>
                                This will permanently delete the token "{token.spec.displayName}".
                                Applications using this token will no longer be able to access your workspaces.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction
                                onClick={() => handleDeleteToken(token.metadata.name)}
                                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                              >
                                Delete
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          )}
        </CardContent>
      </Card>

      <Card className="mt-6">
        <CardHeader>
          <CardTitle>Security Best Practices</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm text-muted-foreground">
          <div className="flex gap-2">
            <span className="font-semibold text-foreground">•</span>
            <p>Treat tokens like passwords - never share them publicly or commit them to version control</p>
          </div>
          <div className="flex gap-2">
            <span className="font-semibold text-foreground">•</span>
            <p>Create separate tokens for different applications or use cases</p>
          </div>
          <div className="flex gap-2">
            <span className="font-semibold text-foreground">•</span>
            <p>Delete tokens that are no longer needed or may have been compromised</p>
          </div>
          <div className="flex gap-2">
            <span className="font-semibold text-foreground">•</span>
            <p>Regularly review and rotate your tokens for enhanced security</p>
          </div>
        </CardContent>
      </Card>
    </>
  )
}
