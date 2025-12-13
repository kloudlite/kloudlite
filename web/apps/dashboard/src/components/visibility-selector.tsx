'use client'

import { useState } from 'react'
import { Globe, Users, Lock, X } from 'lucide-react'
import { Label } from '@kloudlite/ui'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Button } from '@kloudlite/ui'
import type { Visibility } from '@kloudlite/types'

interface VisibilitySelectorProps {
  visibility: Visibility
  sharedWith: string[]
  onVisibilityChange: (visibility: Visibility) => void
  onSharedWithChange: (users: string[]) => void
  disabled?: boolean
}

export function VisibilitySelector({
  visibility,
  sharedWith,
  onVisibilityChange,
  onSharedWithChange,
  disabled = false,
}: VisibilitySelectorProps) {
  const [newUser, setNewUser] = useState('')

  const addUser = () => {
    const user = newUser.trim()
    if (user && !sharedWith.includes(user)) {
      onSharedWithChange([...sharedWith, user])
      setNewUser('')
    }
  }

  const removeUser = (user: string) => {
    onSharedWithChange(sharedWith.filter((u) => u !== user))
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      addUser()
    }
  }

  return (
    <div className="space-y-3">
      <div className="space-y-2">
        <Label>Visibility</Label>
        <Select
          value={visibility}
          onValueChange={(value) => onVisibilityChange(value as Visibility)}
          disabled={disabled}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="private">
              <div className="flex items-center gap-2">
                <Lock className="h-4 w-4" />
                <span>Private</span>
              </div>
            </SelectItem>
            <SelectItem value="shared">
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4" />
                <span>Shared</span>
              </div>
            </SelectItem>
            <SelectItem value="open">
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4" />
                <span>Open</span>
              </div>
            </SelectItem>
          </SelectContent>
        </Select>
        <p className="text-muted-foreground text-xs">
          {visibility === 'private' && 'Only you can see this'}
          {visibility === 'shared' && 'Share with specific team members'}
          {visibility === 'open' && 'Visible to all team members'}
        </p>
      </div>

      {visibility === 'shared' && (
        <div className="space-y-2">
          <Label>Share with</Label>
          <div className="flex gap-2">
            <Input
              placeholder="Enter username"
              value={newUser}
              onChange={(e) => setNewUser(e.target.value)}
              onKeyDown={handleKeyDown}
              disabled={disabled}
              className="flex-1"
            />
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={addUser}
              disabled={disabled || !newUser.trim()}
            >
              Add
            </Button>
          </div>
          {sharedWith.length > 0 && (
            <div className="flex flex-wrap gap-2 mt-2">
              {sharedWith.map((user) => (
                <span
                  key={user}
                  className="bg-muted inline-flex items-center gap-1 rounded-full px-3 py-1 text-sm"
                >
                  {user}
                  <button
                    type="button"
                    onClick={() => removeUser(user)}
                    disabled={disabled}
                    className="hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </span>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// Badge component for displaying visibility in lists
export function VisibilityBadge({ visibility }: { visibility?: Visibility }) {
  if (!visibility || visibility === 'private') {
    return (
      <span className="inline-flex items-center gap-1 text-muted-foreground" title="Private">
        <Lock className="h-3 w-3" />
      </span>
    )
  }

  if (visibility === 'shared') {
    return (
      <span className="inline-flex items-center gap-1 text-blue-600 dark:text-blue-400" title="Shared with specific users">
        <Users className="h-3 w-3" />
      </span>
    )
  }

  if (visibility === 'open') {
    return (
      <span className="inline-flex items-center gap-1 text-green-600 dark:text-green-400" title="Open to all team members">
        <Globe className="h-3 w-3" />
      </span>
    )
  }

  return null
}
