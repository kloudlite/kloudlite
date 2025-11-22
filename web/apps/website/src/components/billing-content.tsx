'use client'

import { Card, CardContent, CardDescription, CardTitle } from '@kloudlite/ui'
import { Cloud, Rocket } from 'lucide-react'

export function BillingContent() {
  return (
    <div className="space-y-6">
      {/* Coming Soon Card */}
      <Card className="border-primary/20 border-2 border-dashed">
        <CardContent className="flex flex-col items-center justify-center py-16 text-center">
          <div className="bg-primary/10 mb-6 rounded-full p-6">
            <Cloud className="text-primary h-16 w-16" />
          </div>
          <CardTitle className="mb-3 text-3xl">Cloud Plans Coming Soon</CardTitle>
          <CardDescription className="mb-8 max-w-md text-base">
            We&apos;re working on bringing you flexible cloud hosting plans for your installations.
            Stay tuned for updates!
          </CardDescription>
          <div className="text-muted-foreground bg-muted/50 flex items-center gap-2 rounded-full px-4 py-2 text-sm">
            <Rocket className="h-4 w-4" />
            <span>In the meantime, enjoy using Kloudlite with your current setup</span>
          </div>
        </CardContent>
      </Card>

      {/* Info Note - smaller and less prominent */}
      <div className="mt-6 text-center">
        <p className="text-muted-foreground text-xs">
          Billing and payment features will be available when cloud plans launch
        </p>
      </div>
    </div>
  )
}
