'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@kloudlite/ui'
import { Loader2, UserPlus, X } from 'lucide-react'
import { Input } from '@kloudlite/ui'
import { updateEnvironmentAccess } from '@/app/actions/environment.actions'
import { toast } from 'sonner'
import type { Visibility } from '@kloudlite/types'

interface AccessSettingsProps {
  environmentId: string
  visibility: Visibility
  sharedWith: string[]
  owner: string
}

export function AccessSettings({ environmentId, visibility, sharedWith, owner }: AccessSettingsProps) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [currentVisibility, setCurrentVisibility] = useState<Visibility>(visibility)
  const [currentSharedWith, setCurrentSharedWith] = useState<string[]>(sharedWith)
  const [newMember, setNewMember] = useState('')

  const handleVisibilityChange = (newVisibility: Visibility) => {
    setCurrentVisibility(newVisibility)
    saveChanges(newVisibility, currentSharedWith)
  }

  const handleAddMember = () => {
    if (!newMember.trim()) return
    if (currentSharedWith.includes(newMember.trim())) {
      toast.error('Member already added')
      return
    }
    const updated = [...currentSharedWith, newMember.trim()]
    setCurrentSharedWith(updated)
    setNewMember('')
    saveChanges(currentVisibility, updated)
  }

  const handleRemoveMember = (member: string) => {
    const updated = currentSharedWith.filter((m) => m !== member)
    setCurrentSharedWith(updated)
    saveChanges(currentVisibility, updated)
  }

  const saveChanges = (vis: Visibility, shared: string[]) => {
    startTransition(async () => {
      const result = await updateEnvironmentAccess(environmentId, {
        visibility: vis,
        sharedWith: vis === 'shared' ? shared : undefined,
      })
      if (result.success) {
        toast.success('Access settings updated')
        router.refresh()
      } else {
        toast.error(result.error || 'Failed to update access settings')
      }
    })
  }

  return (
    <div className="space-y-4">
      <div className="mb-4">
        <h3 className="text-lg font-medium">Access Control</h3>
        <p className="text-muted-foreground text-sm">
          Manage environment visibility and access permissions
        </p>
      </div>
      <div className="bg-card rounded-lg border p-6">
        <div className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium">Owner</label>
            <p className="text-muted-foreground text-sm">{owner}</p>
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium">Environment Visibility</label>
            <div className="space-y-2">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="visibility"
                  checked={currentVisibility === 'private'}
                  onChange={() => handleVisibilityChange('private')}
                  disabled={isPending}
                  className="text-info"
                />
                <span className="text-sm">
                  Private - Only accessible to owner
                </span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="visibility"
                  checked={currentVisibility === 'shared'}
                  onChange={() => handleVisibilityChange('shared')}
                  disabled={isPending}
                  className="text-info"
                />
                <span className="text-sm">Shared - Accessible to owner and invited members</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="visibility"
                  checked={currentVisibility === 'public'}
                  onChange={() => handleVisibilityChange('public')}
                  disabled={isPending}
                  className="text-info"
                />
                <span className="text-sm">Public - Accessible to everyone</span>
              </label>
            </div>
          </div>

          {currentVisibility === 'shared' && (
            <div className="pt-2">
              <h4 className="mb-2 text-sm font-medium">Shared With</h4>
              {currentSharedWith.length > 0 ? (
                <div className="space-y-2 mb-3">
                  {currentSharedWith.map((member) => (
                    <div key={member} className="bg-muted/50 flex items-center justify-between rounded p-2">
                      <span className="text-sm">{member}</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveMember(member)}
                        disabled={isPending}
                        className="text-destructive hover:text-destructive/80 h-7 w-7 p-0"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground text-sm mb-3">No members added yet</p>
              )}
              <div className="flex gap-2">
                <Input
                  placeholder="Enter username"
                  value={newMember}
                  onChange={(e) => setNewMember(e.target.value)}
                  disabled={isPending}
                  onKeyDown={(e) => e.key === 'Enter' && handleAddMember()}
                  className="flex-1"
                />
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleAddMember}
                  disabled={isPending || !newMember.trim()}
                >
                  {isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <UserPlus className="h-4 w-4" />}
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
