'use client'

import { Card, CardContent, CardDescription, CardTitle } from '@/components/ui/card'
import { Cloud, Rocket } from 'lucide-react'

export function BillingContent() {
  return (
    <div className="space-y-6">
      {/* Coming Soon Card */}
      <Card className="border-2 border-dashed border-primary/20">
        <CardContent className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-primary/10 p-6 mb-6">
            <Cloud className="h-16 w-16 text-primary" />
          </div>
          <CardTitle className="text-3xl mb-3">Cloud Plans Coming Soon</CardTitle>
          <CardDescription className="max-w-md text-base mb-8">
            We&apos;re working on bringing you flexible cloud hosting plans for your installations.
            Stay tuned for updates!
          </CardDescription>
          <div className="flex items-center gap-2 text-sm text-muted-foreground bg-muted/50 px-4 py-2 rounded-full">
            <Rocket className="h-4 w-4" />
            <span>In the meantime, enjoy using Kloudlite with your current setup</span>
          </div>
        </CardContent>
      </Card>

      {/* Info Note - smaller and less prominent */}
      <div className="mt-6 text-center">
        <p className="text-xs text-muted-foreground">
          Billing and payment features will be available when cloud plans launch
        </p>
      </div>
    </div>
  )
}
