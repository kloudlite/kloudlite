'use client'

import { useState, useEffect } from 'react'
import { Loader2, UserPlus, Trash2, AlertCircle } from 'lucide-react'

interface TenantUser {
  metadata: { name: string }
  spec: { email?: string; displayName?: string; roles?: string[]; active?: boolean }
}

export function TeamManagementClient({
  installationId,
  isOwner,
}: {
  installationId: string
  apiServerUrl?: string
  isOwner: boolean
}) {
  const [users, setUsers] = useState<TenantUser[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [newEmail, setNewEmail] = useState('')
  const [adding, setAdding] = useState(false)

  useEffect(() => {
    fetch(`/api/installations/${installationId}/users`)
      .then((r) => r.json())
      .then((data) => {
        if (data.error) throw new Error(data.error)
        setUsers(data.users || [])
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [installationId])

  async function addUser() {
    if (!newEmail.trim()) return
    setAdding(true)
    setError(null)
    try {
      const r = await fetch(`/api/installations/${installationId}/users`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: newEmail.trim() }),
      })
      const data = await r.json()
      if (data.error) throw new Error(data.error)
      setUsers((prev) => [...prev, data.user])
      setNewEmail('')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to add user')
    } finally {
      setAdding(false)
    }
  }

  async function removeUser(name: string) {
    if (!confirm('Remove this user from the installation?')) return
    try {
      const r = await fetch(`/api/installations/${installationId}/users`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
      })
      const data = await r.json()
      if (data.error) throw new Error(data.error)
      setUsers((prev) => prev.filter((u) => u.metadata.name !== name))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to remove user')
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error && users.length === 0) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center max-w-md">
          <AlertCircle className="mx-auto mb-4 h-10 w-10 text-destructive" />
          <h3 className="text-foreground text-lg font-medium mb-1">Couldn't load team</h3>
          <p className="text-muted-foreground text-sm">
            {error.includes('API server URL')
              ? 'This installation has no API server configured yet. Complete the setup first.'
              : error}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="border border-foreground/10 rounded-lg bg-background">
      {isOwner && (
        <div className="border-b border-foreground/10 p-4">
          <div className="flex gap-2">
            <input
              type="email"
              value={newEmail}
              onChange={(e) => setNewEmail(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && addUser()}
              placeholder="Enter email address..."
              className="flex-1 rounded-md border border-foreground/10 bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-foreground/20"
            />
            <button
              onClick={addUser}
              disabled={adding || !newEmail.trim()}
              className="inline-flex items-center gap-1.5 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              {adding ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <UserPlus className="size-4" />
              )}
              Add User
            </button>
          </div>
          {error && <p className="mt-2 text-xs text-destructive">{error}</p>}
        </div>
      )}

      {users.length === 0 ? (
        <div className="p-8 text-center text-muted-foreground text-sm">
          No users found. Add users to grant them access to this installation.
        </div>
      ) : (
        <div className="divide-y divide-foreground/10">
          {users.map((user) => (
            <div key={user.metadata.name} className="flex items-center justify-between px-4 py-3">
              <div>
                <p className="text-foreground text-sm font-medium">
                  {user.spec.displayName || user.metadata.name}
                </p>
                <p className="text-muted-foreground text-xs">{user.spec.email}</p>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-muted-foreground text-xs capitalize">
                  {user.spec.roles?.join(', ') || 'member'}
                </span>
                {isOwner && (
                  <button
                    onClick={() => removeUser(user.metadata.name)}
                    className="text-muted-foreground hover:text-destructive transition-colors"
                    title="Remove user"
                  >
                    <Trash2 className="size-4" />
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
