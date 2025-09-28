'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { updateProvider } from '../../actions'

interface Provider {
  type: string
  enabled: boolean
  clientId: string
  clientSecret?: string
}

interface ProviderCardProps {
  provider: Provider
  displayName: string
}

export function ProviderCard({ provider, displayName }: ProviderCardProps) {
  const router = useRouter()
  const [isEditing, setIsEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [formData, setFormData] = useState({
    enabled: provider.enabled,
    clientId: provider.clientId || '',
    clientSecret: provider.clientSecret || '',
  })
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null)

  const handleSave = async () => {
    setSaving(true)
    setMessage(null)

    try {
      const result = await updateProvider(provider.type, {
        ...provider,
        ...formData,
      })

      if (result.success) {
        setMessage({ type: 'success', text: 'Saved successfully' })
        setIsEditing(false)
        router.refresh() // Refresh server-side data
        setTimeout(() => setMessage(null), 3000)
      } else {
        setMessage({ type: 'error', text: result.error || 'Failed to save' })
      }
    } catch (error) {
      setMessage({ type: 'error', text: 'An error occurred' })
    } finally {
      setSaving(false)
    }
  }

  const handleCancel = () => {
    setFormData({
      enabled: provider.enabled,
      clientId: provider.clientId || '',
      clientSecret: provider.clientSecret || '',
    })
    setIsEditing(false)
    setMessage(null)
  }

  const handleToggle = async (checked: boolean) => {
    setSaving(true)
    setMessage(null)

    try {
      const result = await updateProvider(provider.type, {
        ...provider,
        ...formData,
        enabled: checked,
      })

      if (result.success) {
        setFormData(prev => ({ ...prev, enabled: checked }))
        setMessage({ type: 'success', text: 'Updated successfully' })
        router.refresh() // Refresh server-side data
        setTimeout(() => setMessage(null), 3000)
      } else {
        setMessage({ type: 'error', text: result.error || 'Failed to update' })
      }
    } catch (error) {
      setMessage({ type: 'error', text: 'An error occurred' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card className="border-gray-200">
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-medium">{displayName}</CardTitle>
          <Switch
            checked={formData.enabled}
            onCheckedChange={handleToggle}
            disabled={saving || isEditing}
            className="data-[state=checked]:bg-gray-900"
          />
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {message && (
          <div className={`text-sm ${message.type === 'success' ? 'text-green-600' : 'text-red-600'}`}>
            {message.text}
          </div>
        )}

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
                className="bg-gray-900 hover:bg-gray-800"
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
                <p className="text-xs text-gray-500 mb-1">Client ID</p>
                <p className="text-sm font-mono text-gray-900 truncate">
                  {formData.clientId || 'Not configured'}
                </p>
              </div>
              <div>
                <p className="text-xs text-gray-500 mb-1">Client Secret</p>
                <p className="text-sm font-mono text-gray-900">
                  {formData.clientSecret ? '••••••••' : 'Not configured'}
                </p>
              </div>
            </div>

            <Button
              onClick={() => setIsEditing(true)}
              variant="outline"
              size="sm"
              className="w-full"
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