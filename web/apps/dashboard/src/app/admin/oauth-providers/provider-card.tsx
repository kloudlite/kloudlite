'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle, Button, Input, Label, Switch } from '@kloudlite/ui'
import { toast } from 'sonner'
import { updateOAuthProvider } from './oauth-actions'
import type { OAuthProvider } from '@/lib/services/oauth-provider.service'

interface ProviderCardProps {
  provider: OAuthProvider
  displayName: string
  isReadOnly?: boolean
}

export function ProviderCard({ provider, displayName, isReadOnly = false }: ProviderCardProps) {
  const router = useRouter()
  const [isEditing, setIsEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [formData, setFormData] = useState({
    type: provider.type,
    enabled: provider.enabled,
    clientId: provider.clientId || '',
    clientSecret: provider.clientSecret || '',
  })

  const handleSave = async () => {
    setSaving(true)

    try {
      const result = await updateOAuthProvider(provider.type, formData)

      if (result.success) {
        toast.success(`${displayName} configuration saved successfully`)
        setIsEditing(false)
        router.refresh() // Refresh server-side data
      } else {
        toast.error(result.error || 'Failed to save configuration')
      }
    } catch (_error) {
      toast.error('An error occurred while saving')
    } finally {
      setSaving(false)
    }
  }

  const handleCancel = () => {
    setFormData({
      type: provider.type,
      enabled: provider.enabled,
      clientId: provider.clientId || '',
      clientSecret: provider.clientSecret || '',
    })
    setIsEditing(false)
  }

  const handleToggle = async (checked: boolean) => {
    setSaving(true)

    try {
      const updatedProvider = { ...formData, enabled: checked }
      const result = await updateOAuthProvider(provider.type, updatedProvider)

      if (result.success) {
        setFormData((prev) => ({ ...prev, enabled: checked }))
        toast.success(`${displayName} ${checked ? 'enabled' : 'disabled'} successfully`)
        router.refresh() // Refresh server-side data
      } else {
        toast.error(result.error || 'Failed to update provider')
      }
    } catch (_error) {
      toast.error('An error occurred while updating')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card className="border border-gray-200 shadow-sm transition-shadow hover:shadow-md">
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-medium text-gray-900">{displayName}</CardTitle>
          <Switch
            checked={formData.enabled}
            onCheckedChange={handleToggle}
            disabled={saving || isEditing || isReadOnly}
            className="data-[state=checked]:bg-blue-600"
          />
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {isEditing ? (
          <>
            <div className="space-y-2">
              <Label htmlFor={`${provider.type}-client-id`} className="text-sm">
                Client ID
              </Label>
              <Input
                id={`${provider.type}-client-id`}
                type="text"
                value={formData.clientId}
                onChange={(e) => setFormData((prev) => ({ ...prev, clientId: e.target.value }))}
                disabled={saving}
                className="h-9 text-sm"
                placeholder="Enter client ID"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor={`${provider.type}-client-secret`} className="text-sm">
                Client Secret
              </Label>
              <Input
                id={`${provider.type}-client-secret`}
                type="password"
                value={formData.clientSecret}
                onChange={(e) => setFormData((prev) => ({ ...prev, clientSecret: e.target.value }))}
                disabled={saving}
                className="h-9 text-sm"
                placeholder="Enter client secret"
              />
            </div>

            <div className="flex gap-2 pt-2">
              <Button
                onClick={handleSave}
                disabled={saving}
                size="sm"
                className="bg-blue-600 text-white hover:bg-blue-700"
              >
                {saving ? 'Saving...' : 'Save'}
              </Button>
              <Button onClick={handleCancel} disabled={saving} variant="outline" size="sm">
                Cancel
              </Button>
            </div>
          </>
        ) : (
          <>
            <div className="space-y-3">
              <div>
                <p className="mb-1 text-xs font-medium tracking-wider text-gray-500 uppercase">
                  Client ID
                </p>
                <p className="truncate font-mono text-sm text-gray-700">
                  {formData.clientId || 'Not configured'}
                </p>
              </div>
              <div>
                <p className="mb-1 text-xs font-medium tracking-wider text-gray-500 uppercase">
                  Client Secret
                </p>
                <p className="font-mono text-sm text-gray-700">
                  {formData.clientSecret ? '••••••••' : 'Not configured'}
                </p>
              </div>
            </div>

            {!isReadOnly && (
              <Button
                onClick={() => setIsEditing(true)}
                variant="outline"
                size="sm"
                className="w-full border-gray-300 hover:bg-gray-50"
                disabled={saving}
              >
                Configure
              </Button>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}
