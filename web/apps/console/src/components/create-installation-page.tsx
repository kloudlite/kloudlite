'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Button, Input } from '@kloudlite/ui'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { getErrorMessage } from '@/lib/errors'

interface CreateInstallationPageProps {
  orgId: string
}

export function CreateInstallationPage({ orgId }: CreateInstallationPageProps) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [name, setName] = useState(searchParams.get('name') || '')
  const [creating, setCreating] = useState(false)
  const [hostingType, setHostingType] = useState<'kloudlite' | 'byoc'>('byoc')
  const checkoutSessionId = searchParams.get('checkout_session') || undefined
  const paymentSuccess = searchParams.get('payment') === 'success'

  useEffect(() => {
    if (paymentSuccess && checkoutSessionId) {
      toast.success('Payment successful! Credits added to your account.')
    }
  }, [paymentSuccess, checkoutSessionId])

  const handleCreate = async () => {
    if (!name.trim()) {
      toast.error('Please enter an installation name')
      return
    }

    setCreating(true)
    try {
      const response = await fetch('/api/installations/create-installation', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: name.trim(),
          orgId,
          hostingType,
        }),
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Failed to create installation')
      }

      const data = await response.json()
      toast.success('Installation created')

      if (hostingType === 'kloudlite') {
        router.push(`/installations/${data.installationId}/install`)
      } else {
        router.push(`/installations/${data.installationId}`)
      }
    } catch (err) {
      toast.error(getErrorMessage(err, 'Failed to create installation'))
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="max-w-lg mx-auto space-y-8">
      <div>
        <h1 className="text-foreground text-2xl font-semibold">New Installation</h1>
        <p className="text-muted-foreground mt-1 text-sm">Create a new Kloudlite installation</p>
      </div>

      <div className="space-y-4">
        <div>
          <label className="text-sm font-medium text-foreground">Name</label>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="My Installation"
            className="mt-1.5"
          />
        </div>

        <div>
          <label className="text-sm font-medium text-foreground">Hosting Type</label>
          <div className="mt-2 flex gap-2">
            <button
              type="button"
              onClick={() => setHostingType('byoc')}
              className={`flex-1 rounded-md border px-4 py-3 text-sm font-medium transition-colors ${
                hostingType === 'byoc'
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-foreground/10 text-muted-foreground hover:border-foreground/30'
              }`}
            >
              <div className="font-medium">Self-Hosted</div>
              <div className="text-[11px] opacity-70 mt-0.5">Use your own cloud account</div>
            </button>
            <button
              type="button"
              onClick={() => setHostingType('kloudlite')}
              className={`flex-1 rounded-md border px-4 py-3 text-sm font-medium transition-colors ${
                hostingType === 'kloudlite'
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-foreground/10 text-muted-foreground hover:border-foreground/30'
              }`}
            >
              <div className="font-medium">Kloudlite Cloud</div>
              <div className="text-[11px] opacity-70 mt-0.5">Managed by us on OCI</div>
            </button>
          </div>
        </div>

        <Button
          onClick={handleCreate}
          disabled={creating || !name.trim()}
          className="w-full"
          size="lg"
        >
          {creating ? <Loader2 className="size-4 animate-spin mr-2" /> : null}
          Create Installation
        </Button>
      </div>
    </div>
  )
}
