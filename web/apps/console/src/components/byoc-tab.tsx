'use client'

import { Button } from '@kloudlite/ui'
import { Loader2, Terminal, Cloud, Shield } from 'lucide-react'

interface ByocTabProps {
  creating: boolean
  subdomainAvailable: boolean | null
  onSubmit: () => void
}

export function ByocTab({ creating, subdomainAvailable, onSubmit }: ByocTabProps) {
  return (
    <div className="rounded-lg border border-foreground/10 bg-background">
      <div className="border-b border-foreground/10 px-6 py-4">
        <h3 className="font-medium text-foreground">What you&apos;ll need</h3>
      </div>
      <div className="px-6 py-5 space-y-4">
        <div className="space-y-3">
          <div className="flex items-start gap-3">
            <Cloud className="size-4 text-primary mt-0.5 shrink-0" />
            <div>
              <p className="text-sm font-medium text-foreground">Cloud provider account</p>
              <p className="text-xs text-muted-foreground">AWS, GCP, or Azure with permissions to create resources</p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <Terminal className="size-4 text-primary mt-0.5 shrink-0" />
            <div>
              <p className="text-sm font-medium text-foreground">CLI access</p>
              <p className="text-xs text-muted-foreground">Cloud provider CLI configured and authenticated</p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <Shield className="size-4 text-primary mt-0.5 shrink-0" />
            <div>
              <p className="text-sm font-medium text-foreground">No billing required</p>
              <p className="text-xs text-muted-foreground">Resources run in your own cloud account — you pay your cloud provider directly</p>
            </div>
          </div>
        </div>

        <div className="pt-2">
          <Button
            type="button"
            className="w-full"
            size="lg"
            disabled={creating || subdomainAvailable !== true}
            onClick={onSubmit}
          >
            {creating ? (
              <>
                <Loader2 className="mr-2 size-4 animate-spin" />
                Creating...
              </>
            ) : (
              'Continue to Setup'
            )}
          </Button>
        </div>
      </div>
    </div>
  )
}
