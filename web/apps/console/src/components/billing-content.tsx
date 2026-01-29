'use client'

import { Cloud, Rocket } from 'lucide-react'

export function BillingContent() {
  return (
    <div className="space-y-6">
      {/* Coming Soon Section */}
      <div>
        <div className="mb-5">
          <h2 className="text-xl font-semibold">Billing & Plans</h2>
          <p className="text-muted-foreground mt-1 text-base">Manage your subscription and payment methods</p>
        </div>

        <div className="flex flex-col items-center justify-center py-12 text-center border border-foreground/10 border-dashed">
          <div className="bg-primary/10 mb-5 p-5">
            <Cloud className="text-primary h-12 w-12" />
          </div>
          <h3 className="mb-2 text-2xl font-semibold">Cloud Plans Coming Soon</h3>
          <p className="text-muted-foreground mb-5 max-w-md text-base">
            We&apos;re working on bringing you flexible cloud hosting plans for your installations.
            Stay tuned for updates!
          </p>
          <div className="text-muted-foreground bg-muted/30 border border-foreground/10 flex items-center gap-2 px-4 py-2 text-sm">
            <Rocket className="h-4 w-4" />
            <span>In the meantime, enjoy using Kloudlite with your current setup</span>
          </div>
        </div>
      </div>
    </div>
  )
}
