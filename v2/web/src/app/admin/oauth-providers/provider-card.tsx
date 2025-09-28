'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { toast } from 'sonner'
import { updateOAuthProvider } from './oauth-actions'
import type { OAuthProvider } from '@/lib/services/oauth-provider.service'

interface ProviderCardProps {
  provider: OAuthProvider
  displayName: string
}

export function ProviderCard({ provider, displayName }: ProviderCardProps) {
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
    } catch (error) {
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
        setFormData(prev => ({ ...prev, enabled: checked }))
        toast.success(`${displayName} ${checked ? 'enabled' : 'disabled'} successfully`)
        router.refresh() // Refresh server-side data
      } else {
        toast.error(result.error || 'Failed to update provider')
      }
    } catch (error) {
      toast.error('An error occurred while updating')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card className="border border-gray-200 shadow-sm hover:shadow-md transition-shadow">
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-medium text-gray-900">{displayName}</CardTitle>
          <Switch
            checked={formData.enabled}
            onCheckedChange={handleToggle}
            disabled={saving || isEditing}
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
                onChange={(e) => setFormData(prev => ({ ...prev, clientId: e.target.value }))}
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
                onChange={(e) => setFormData(prev => ({ ...prev, clientSecret: e.target.value }))}
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
                className="bg-blue-600 hover:bg-blue-700 text-white"
              >
                {saving ? 'Saving...' : 'Save'}
              </Button>
              <Button
                onClick={handleCancel}
                disabled={saving}
                variant="outline"
                size="sm"
              >
                Cancel
              </Button>
            </div>
          </>
        ) : (
          <>
            <div className="space-y-3">
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-1">Client ID</p>
                <p className="text-sm font-mono text-gray-700 truncate">
                  {formData.clientId || 'Not configured'}
                </p>
              </div>
              <div>
                <p className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-1">Client Secret</p>
                <p className="text-sm font-mono text-gray-700">
                  {formData.clientSecret ? '••••••••' : 'Not configured'}
                </p>
              </div>
            </div>

            <Button
              onClick={() => setIsEditing(true)}
              variant="outline"
              size="sm"
              className="w-full border-gray-300 hover:bg-gray-50"
              disabled={saving}
            >
              Configure
            </Button>
          </>
        )}
      </CardContent>
    </Card>
  )
}