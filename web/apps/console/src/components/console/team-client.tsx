'use client'

import { Trash2, AlertCircle } from 'lucide-react'
import { useState } from 'react'

interface Member {
  id: string
  userId: string
  role: string
  email?: string
  name?: string
}

export function TeamManagementClient({
  orgId,
  members: initialMembers,
  currentUserId,
  userRole,
}: {
  orgId: string
  members: Member[]
  currentUserId: string
  userRole: string
}) {
  const [members, setMembers] = useState(initialMembers)
  const [error, setError] = useState<string | null>(null)

  async function removeMember(memberId: string) {
    if (!confirm('Remove this member from the organization?')) return
    setError(null)
    try {
      const r = await fetch(`/api/orgs/${orgId}/members/${memberId}`, { method: 'DELETE' })
      if (!r.ok) throw new Error((await r.json()).error || 'Failed to remove')
      setMembers((prev) => prev.filter((m) => m.id !== memberId))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to remove member')
    }
  }

  return (
    <div className="border border-foreground/10 rounded-lg bg-background">
      {error && (
        <div className="border-b border-foreground/10 px-4 py-3 bg-destructive/5">
          <p className="text-xs text-destructive flex items-center gap-1.5">
            <AlertCircle className="size-3.5" />
            {error}
          </p>
        </div>
      )}

      {members.length === 0 ? (
        <div className="p-8 text-center text-muted-foreground text-sm">
          No members found.
        </div>
      ) : (
        <div className="divide-y divide-foreground/10">
          {members.map((member) => (
            <div key={member.id} className="flex items-center justify-between px-4 py-3">
              <div>
                <p className="text-foreground text-sm font-medium">
                  {member.name || member.email || member.userId}
                </p>
                {member.email && member.name && (
                  <p className="text-muted-foreground text-xs">{member.email}</p>
                )}
              </div>
              <div className="flex items-center gap-3">
                <span className="text-muted-foreground text-xs capitalize bg-muted/50 px-2 py-0.5 rounded">
                  {member.role}
                </span>
                {(userRole === 'owner' || userRole === 'admin') && member.userId !== currentUserId && (
                  <button
                    onClick={() => removeMember(member.id)}
                    className="text-muted-foreground hover:text-destructive transition-colors"
                    title="Remove member"
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
