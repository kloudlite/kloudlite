'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Building2, Check, ChevronsUpDown, Plus, Loader2 } from 'lucide-react'
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Input,
  Label,
} from '@kloudlite/ui'

interface Org {
  id: string
  name: string
  slug: string
}

interface OrgSwitcherProps {
  orgs: Org[]
  currentOrgId: string
}

function slugify(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 63)
}

export function OrgSwitcher({ orgs, currentOrgId }: OrgSwitcherProps) {
  const router = useRouter()
  const [switching, setSwitching] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [newOrgName, setNewOrgName] = useState('')
  const [newOrgSlug, setNewOrgSlug] = useState('')
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(false)
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')
  const currentOrg = orgs.find((o) => o.id === currentOrgId) || orgs[0]

  const handleSwitch = async (orgId: string) => {
    if (orgId === currentOrgId) return
    setSwitching(true)
    try {
      await fetch('/api/orgs/select', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ orgId }),
      })
      router.refresh()
    } finally {
      setSwitching(false)
    }
  }

  const handleNameChange = (name: string) => {
    setNewOrgName(name)
    if (!slugManuallyEdited) {
      setNewOrgSlug(slugify(name))
    }
  }

  const handleCreate = async () => {
    if (!newOrgName.trim() || !newOrgSlug.trim()) return
    setCreating(true)
    setCreateError('')
    try {
      const res = await fetch('/api/orgs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newOrgName.trim(), slug: newOrgSlug.trim() }),
      })
      const data = await res.json()
      if (!res.ok) {
        setCreateError(data.error || 'Failed to create organization')
        return
      }
      // Select the new org
      await fetch('/api/orgs/select', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ orgId: data.organization.id }),
      })
      setCreateOpen(false)
      setNewOrgName('')
      setNewOrgSlug('')
      setSlugManuallyEdited(false)
      router.refresh()
    } catch {
      setCreateError('Something went wrong')
    } finally {
      setCreating(false)
    }
  }

  return (
    <>
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            className="gap-2 px-2 hover:bg-muted/50 transition-colors"
            disabled={switching}
          >
            <Building2 className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium truncate max-w-[150px]">
              {currentOrg?.name}
            </span>
            <ChevronsUpDown className="h-3 w-3 text-muted-foreground" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-56">
          <DropdownMenuLabel>Organizations</DropdownMenuLabel>
          <DropdownMenuSeparator />
          {orgs.map((org) => (
            <DropdownMenuItem
              key={org.id}
              onClick={() => handleSwitch(org.id)}
              className="cursor-pointer justify-between"
            >
              <span className="truncate">{org.name}</span>
              {org.id === currentOrgId && (
                <Check className="h-4 w-4 text-primary" />
              )}
            </DropdownMenuItem>
          ))}
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={() => setCreateOpen(true)}
            className="cursor-pointer"
          >
            <Plus className="h-4 w-4 mr-2" />
            Create Organization
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Create Organization</DialogTitle>
            <DialogDescription>
              Create a new organization to manage installations and team members.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="org-name">Name</Label>
              <Input
                id="org-name"
                value={newOrgName}
                onChange={(e) => handleNameChange(e.target.value)}
                placeholder="My Organization"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="org-slug">Slug</Label>
              <Input
                id="org-slug"
                value={newOrgSlug}
                onChange={(e) => {
                  setNewOrgSlug(e.target.value)
                  setSlugManuallyEdited(true)
                }}
                placeholder="my-organization"
              />
              <p className="text-xs text-muted-foreground">
                Lowercase, alphanumeric and hyphens only, 3-63 characters.
              </p>
            </div>
            {createError && (
              <p className="text-sm text-destructive">{createError}</p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)} disabled={creating}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={creating || !newOrgName.trim() || !newOrgSlug.trim()}
            >
              {creating && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
