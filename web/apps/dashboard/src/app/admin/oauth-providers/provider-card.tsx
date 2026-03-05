'use client'

import { useState, useTransition, memo } from 'react'
import { useRouter } from 'next/navigation'
import {
  Alert,
  AlertDescription,
  Badge,
  Button,
  Input,
  Label,
  Switch,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@kloudlite/ui'
import { Eye, EyeOff } from 'lucide-react'
import { toast } from 'sonner'
import { updateOAuthProvider } from './oauth-actions'
import type { OAuthProvider } from '@/lib/services/oauth-provider.service'

const PROVIDERS = [
  {
    type: 'google',
    name: 'Google',
    description: 'Allow users to sign in with their Google workspace or personal accounts',
  },
  {
    type: 'github',
    name: 'GitHub',
    description: 'Allow users to sign in with their GitHub accounts',
  },
  {
    type: 'microsoft',
    name: 'Microsoft',
    description: 'Allow users to sign in with their Microsoft or Azure AD accounts',
  },
] as const

interface OAuthProvidersListProps {
  providers: Record<string, OAuthProvider>
  isReadOnly?: boolean
}

export const OAuthProvidersList = memo(function OAuthProvidersList({ providers, isReadOnly = false }: OAuthProvidersListProps) {
  const router = useRouter()
  const [isSaving, startSaveTransition] = useTransition()
  const [isToggling, startToggleTransition] = useTransition()
  const [editingType, setEditingType] = useState<string | null>(null)
  const [formData, setFormData] = useState({ clientId: '', clientSecret: '', tenantId: '' })
  const [formError, setFormError] = useState('')
  const [showSecret, setShowSecret] = useState(false)

  const getProvider = (type: string): OAuthProvider => {
    return providers[type] || { type, enabled: false, clientId: '', clientSecret: '' }
  }

  const handleEdit = (type: string) => {
    const provider = getProvider(type)
    setFormData({ clientId: provider.clientId || '', clientSecret: provider.clientSecret || '', tenantId: provider.tenantId || '' })
    setFormError('')
    setShowSecret(false)
    setEditingType(type)
  }

  const handleSave = () => {
    if (!editingType) return
    const provider = getProvider(editingType)
    setFormError('')

    startSaveTransition(async () => {
      try {
        const result = await updateOAuthProvider(editingType, {
          ...provider,
          clientId: formData.clientId,
          clientSecret: formData.clientSecret,
          ...(editingType === 'microsoft' ? { tenantId: formData.tenantId } : {}),
        })

        if (result.success) {
          toast.success(`${editingType} configuration saved`)
          setEditingType(null)
          router.refresh()
        } else {
          setFormError(result.error || 'Failed to save configuration')
        }
      } catch {
        setFormError('An unexpected error occurred')
      }
    })
  }

  const handleToggle = (type: string, checked: boolean) => {
    const provider = getProvider(type)

    startToggleTransition(async () => {
      try {
        const result = await updateOAuthProvider(type, { ...provider, enabled: checked })

        if (result.success) {
          toast.success(`${type} ${checked ? 'enabled' : 'disabled'}`)
          router.refresh()
        } else {
          toast.error(result.error || 'Failed to update provider')
        }
      } catch {
        toast.error('An unexpected error occurred')
      }
    })
  }

  const editingMeta = PROVIDERS.find((p) => p.type === editingType)

  return (
    <>
      <div className="space-y-6">
        {PROVIDERS.map(({ type, name, description }) => {
          const provider = getProvider(type)
          const isConfigured = !!provider.clientId

          return (
            <div key={type} className="bg-card rounded-lg border">
              {/* Header */}
              <div className="flex items-start justify-between p-5">
                <div className="space-y-1">
                  <div className="flex items-center gap-3">
                    <h3 className="text-foreground text-base font-medium">{name}</h3>
                    <Badge variant={isConfigured ? 'success' : 'secondary'}>
                      {isConfigured ? 'Configured' : 'Not configured'}
                    </Badge>
                  </div>
                  <p className="text-muted-foreground text-sm">{description}</p>
                </div>
                <Switch
                  checked={provider.enabled}
                  onCheckedChange={(checked) => handleToggle(type, checked)}
                  disabled={isToggling || isReadOnly || !isConfigured}
                />
              </div>

              {/* Credentials */}
              <div className="border-t px-5 py-4">
                <div className="grid grid-cols-2 gap-x-8 gap-y-3">
                  <div>
                    <p className="text-muted-foreground mb-1 text-xs font-medium uppercase tracking-wider">Client ID</p>
                    <p className="text-foreground font-mono text-sm">
                      {provider.clientId || <span className="text-muted-foreground">Not set</span>}
                    </p>
                  </div>
                  <div>
                    <p className="text-muted-foreground mb-1 text-xs font-medium uppercase tracking-wider">Client Secret</p>
                    <p className="text-foreground font-mono text-sm">
                      {provider.clientSecret
                        ? '••••••••••••••••'
                        : <span className="text-muted-foreground">Not set</span>}
                    </p>
                  </div>
                </div>

                {!isReadOnly && (
                  <div className="mt-4">
                    <Button variant="outline" size="sm" onClick={() => handleEdit(type)} disabled={isSaving}>
                      {isConfigured ? 'Edit credentials' : 'Set up credentials'}
                    </Button>
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>

      {/* Edit Dialog */}
      <Dialog open={!!editingType} onOpenChange={(open) => !open && setEditingType(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              Configure {editingMeta?.name} OAuth
            </DialogTitle>
            <DialogDescription>
              Enter the OAuth 2.0 client credentials from your {editingMeta?.name} developer console
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {formError && (
              <Alert variant="destructive">
                <AlertDescription>{formError}</AlertDescription>
              </Alert>
            )}

            <div className="space-y-2">
              <Label htmlFor="client-id">Client ID</Label>
              <Input
                id="client-id"
                type="text"
                value={formData.clientId}
                onChange={(e) => setFormData((prev) => ({ ...prev, clientId: e.target.value }))}
                disabled={isSaving}
                placeholder="Enter client ID"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="client-secret">Client Secret</Label>
              <div className="relative">
                <Input
                  id="client-secret"
                  type={showSecret ? 'text' : 'password'}
                  value={formData.clientSecret}
                  onChange={(e) => setFormData((prev) => ({ ...prev, clientSecret: e.target.value }))}
                  disabled={isSaving}
                  placeholder="Enter client secret"
                  className="pr-10"
                />
                <button
                  type="button"
                  onClick={() => setShowSecret(!showSecret)}
                  className="text-muted-foreground hover:text-foreground absolute right-3 top-1/2 -translate-y-1/2"
                >
                  {showSecret ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
            </div>

            {editingType === 'microsoft' && (
              <div className="space-y-2">
                <Label htmlFor="tenant-id">Tenant ID</Label>
                <Input
                  id="tenant-id"
                  type="text"
                  value={formData.tenantId}
                  onChange={(e) => setFormData((prev) => ({ ...prev, tenantId: e.target.value }))}
                  disabled={isSaving}
                  placeholder="Enter tenant ID (or leave empty for 'common')"
                />
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditingType(null)} disabled={isSaving}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={isSaving}>
              {isSaving ? 'Saving...' : 'Save credentials'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
})
