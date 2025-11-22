'use client'

import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@kloudlite/ui'
import { Button } from '@kloudlite/ui'
import { Input } from '@kloudlite/ui'
import { Label } from '@kloudlite/ui'
import { Alert, AlertDescription } from '@kloudlite/ui'
import { Loader2, AlertCircle, Copy } from 'lucide-react'
import { cloneEnvironment } from '@/app/actions/environment.actions'
import type { EnvironmentUIModel } from '@kloudlite/types'

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
  // Remove username prefix from source environment name if present
  const getDefaultName = () => {
    const sourceName = sourceEnvironment.name
    const username = currentUser.includes('@') ? currentUser.split('@')[0] : currentUser
    const prefix = `${username}--`
    const nameWithoutPrefix = sourceName.startsWith(prefix)
      ? sourceName.substring(prefix.length)
      : sourceName
    return `${nameWithoutPrefix}-copy`
  }
  const [name, setName] = useState(getDefaultName())

  const validateNamespace = (name: string): string | null => {
    if (!name) {
      return 'Namespace name is required'
    }
    if (name.length > 63) {
      return 'Namespace name must be no more than 63 characters'
    }
    if (name.includes('--')) {
      return 'Environment name cannot contain "--" (double hyphens)'
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
    const nameError = validateNamespace(name)
    if (nameError) {
      setError(`Environment name: ${nameError}`)
      return
    }

    setLoading(true)

    try {
      const result = await cloneEnvironment(
        sourceEnvironment.name,
        name,
        '', // targetNamespace - always empty, let webhook auto-generate
        true, // cloneEnvVars - always true, controller handles all resources
        true, // cloneFiles - always true, controller handles all resources
        currentUser,
      )

      if (result.success) {
        // Reset form
        setName(getDefaultName())
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
      setName(getDefaultName())
      setError(null)
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
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={loading}
                required
              />
              <p className="text-muted-foreground text-xs">
                Must be lowercase alphanumeric or &quot;-&quot;, max 63 characters
              </p>
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
