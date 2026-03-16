'use client'

import { Button } from '@kloudlite/ui'
import { Wallet, ArrowRight } from 'lucide-react'
import { useRouter } from 'next/navigation'

interface CreditTopupPromptProps {
  installationId: string
}

export function CreditTopupPrompt({ installationId }: CreditTopupPromptProps) {
  const router = useRouter()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-foreground text-2xl font-semibold tracking-tight">
          Add Credits to Continue
        </h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Your account needs credits before we can deploy this installation
        </p>
      </div>

      <div className="rounded-lg border border-foreground/10 bg-background">
        <div className="border-b border-foreground/10 px-6 py-4">
          <h3 className="font-medium text-foreground">Insufficient Credits</h3>
        </div>
        <div className="px-6 py-6 space-y-4">
          <div className="flex items-start gap-3">
            <Wallet className="size-5 text-muted-foreground mt-0.5 shrink-0" />
            <div>
              <p className="text-sm text-foreground">
                Kloudlite Cloud installations require credits to cover infrastructure costs.
                Add credits to your account to start the deployment.
              </p>
            </div>
          </div>

          <div className="flex gap-3">
            <Button
              onClick={() => router.push('/installations/settings/billing')}
            >
              <Wallet className="size-4 mr-2" />
              Add Credits
            </Button>
            <Button
              variant="outline"
              onClick={() => router.push(`/api/installations/${installationId}/continue`)}
            >
              Retry
              <ArrowRight className="size-4 ml-2" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
