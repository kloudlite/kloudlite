'use client'

import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Loader2, AlertCircle, ChevronDown, ChevronRight, Copy } from 'lucide-react'
import { cloneEnvironment } from '@/app/actions/environment.actions'
import type { EnvironmentUIModel } from '@/types/environment'

interface CloneEnvironmentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sourceEnvironment: EnvironmentUIModel
  onSuccess?: () => void
  currentUser?: string
}

export function CloneEnvironmentDialog({
  open,
  onOpenChange,
  sourceEnvironment,
  onSuccess,
  currentUser = 'test-user',
}: CloneEnvironmentDialogProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [formData, setFormData] = useState({
    name: `${sourceEnvironment.name}-copy`,
    targetNamespace: '',
  })

  const validateNamespace = (name: string): string | null => {
    if (!name) {
      return 'Namespace name is required'
    }
    if (name.length > 63) {
      return 'Namespace name must be no more than 63 characters'
    }
    const dnsLabelRegex = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
    if (!dnsLabelRegex.test(name)) {
      return 'Namespace name must consist of lower case alphanumeric characters or "-", and must start and end with an alphanumeric character'
    }

    const reservedNamespaces = [
      'kube-system',
      'kube-public',
      'kube-node-lease',
      'default',
      'kloudlite-system',
    ]

    if (reservedNamespaces.includes(name)) {
      return `Cannot use reserved namespace name: ${name}`
    }

    const reservedPrefixes = ['kube-', 'openshift-', 'kubernetes-']
    for (const prefix of reservedPrefixes) {
      if (name.startsWith(prefix)) {
        return `Cannot use namespace name with reserved prefix: ${prefix}`
      }
    }

    return null
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    // Validate environment name
    const nameError = validateNamespace(formData.name)
    if (nameError) {
      setError(`Environment name: ${nameError}`)
      return
    }

    // Auto-generate or validate namespace
    const targetNamespace = formData.targetNamespace || `env-${formData.name}`
    const namespaceError = validateNamespace(targetNamespace)
    if (namespaceError) {
      setError(namespaceError)
      return
    }

    setLoading(true)

    try {
      const result = await cloneEnvironment(
        sourceEnvironment.name,
        formData.name,
        targetNamespace,
        true, // cloneEnvVars - always true, controller handles all resources
        true, // cloneFiles - always true, controller handles all resources
        currentUser,
      )

      if (result.success) {
        // Reset form
        setFormData({
          name: `${sourceEnvironment.name}-copy`,
          targetNamespace: '',
        })
        onOpenChange(false)

        // Call success callback
        if (onSuccess) {
          onSuccess()
        }
      } else {
        setError(result.error || 'Failed to clone environment. Please try again.')
      }
    } catch (err) {
      console.error('Failed to clone environment:', err)
      const error = err instanceof Error ? err : new Error('Failed to clone environment')
      setError(error.message)
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    if (!loading) {
      setFormData({
        name: `${sourceEnvironment.name}-copy`,
        targetNamespace: '',
      })
      setError(null)
      setShowAdvanced(false)
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Copy className="h-5 w-5" />
              Clone Environment
            </DialogTitle>
            <DialogDescription>
              Create a copy of &quot;{sourceEnvironment.name}&quot; with all its resources including
              environment variables, secrets, configuration files, and compositions.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">New Environment Name</Label>
              <Input
                id="name"
                placeholder="my-cloned-environment"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                disabled={loading}
                required
              />
              <p className="text-muted-foreground text-xs">
                Must be lowercase alphanumeric or &quot;-&quot;, max 63 characters
              </p>
            </div>

            {/* Advanced Options */}
            <div className="pt-2">
              <button
                type="button"
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="text-muted-foreground hover:text-foreground flex items-center gap-1 text-sm transition-colors"
              >
                {showAdvanced ? (
                  <ChevronDown className="h-4 w-4" />
                ) : (
                  <ChevronRight className="h-4 w-4" />
                )}
                Advanced Options
              </button>

              {showAdvanced && (
                <div className="space-y-2 pt-3">
                  <div className="space-y-2">
                    <Label htmlFor="namespace">Target Namespace (Optional)</Label>
                    <Input
                      id="namespace"
                      placeholder={`env-${formData.name || 'environment-name'}`}
                      value={formData.targetNamespace}
                      onChange={(e) =>
                        setFormData({ ...formData, targetNamespace: e.target.value })
                      }
                      disabled={loading}
                    />
                    <p className="text-muted-foreground text-xs">
                      Leave empty to auto-generate as &quot;env-{'{name}'}&quot;. The Kubernetes
                      namespace that will be created for this environment.
                    </p>
                  </div>
                </div>
              )}
            </div>

            {error && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {loading ? 'Cloning...' : 'Clone Environment'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
