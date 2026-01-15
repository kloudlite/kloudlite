'use client'

import { useState, useEffect } from 'react'
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
import { Loader2, AlertCircle, GitFork } from 'lucide-react'
import { forkEnvironment } from '@/app/actions/snapshot.actions'

interface ForkEnvironmentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
  sourceEnvironment: string
}

export function ForkEnvironmentDialog({
  open,
  onOpenChange,
  onSuccess,
  sourceEnvironment,
}: ForkEnvironmentDialogProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [name, setName] = useState('')

  // Reset form when dialog opens
  useEffect(() => {
    if (open) {
      setName(`${sourceEnvironment}-fork`)
      setError(null)
    }
  }, [open, sourceEnvironment])

  const validateNamespace = (name: string): string | null => {
    if (!name) {
      return 'Environment name is required'
    }
    if (name.length > 63) {
      return 'Environment name must be no more than 63 characters'
    }
    if (name.includes('--')) {
      return 'Environment name cannot contain "--" (double hyphens)'
    }
    const dnsLabelRegex = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
    if (!dnsLabelRegex.test(name)) {
      return 'Environment name must consist of lower case alphanumeric characters or "-", and must start and end with an alphanumeric character'
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
      setError(nameError)
      return
    }

    setLoading(true)

    try {
      const result = await forkEnvironment(sourceEnvironment, name)

      if (result.success) {
        // Reset form
        setName('')
        onOpenChange(false)

        // Call success callback
        if (onSuccess) {
          onSuccess()
        }
      } else {
        setError(result.error || 'Failed to fork environment. Please try again.')
      }
    } catch (err) {
      console.error('Failed to fork environment:', err)
      const error = err instanceof Error ? err : new Error('Failed to fork environment')
      setError(error.message)
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    if (!loading) {
      setName('')
      setError(null)
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <GitFork className="h-5 w-5" />
              Fork Environment
            </DialogTitle>
            <DialogDescription>
              Create a new environment from the latest snapshot of <strong>{sourceEnvironment}</strong>.
              You can restore to a different snapshot later if needed.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* New Environment Name */}
            <div className="space-y-2">
              <Label htmlFor="name">New Environment Name</Label>
              <Input
                id="name"
                placeholder="my-forked-environment"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={loading}
                required
                autoFocus
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
            <Button type="submit" disabled={loading || !name}>
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {loading ? 'Forking...' : 'Fork'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
