'use client'

import { AlertTriangle } from 'lucide-react'
import { Button } from '@kloudlite/ui'

interface PaymentWarningBannerProps {
  onManageBilling: () => void
}

export function PaymentWarningBanner({ onManageBilling }: PaymentWarningBannerProps) {
  return (
    <div className="rounded-lg border border-amber-200 bg-amber-50 p-4 dark:border-amber-800 dark:bg-amber-950">
      <div className="flex items-center gap-3">
        <AlertTriangle className="h-5 w-5 shrink-0 text-amber-600 dark:text-amber-400" />
        <div className="flex-1">
          <p className="text-sm font-medium text-amber-800 dark:text-amber-200">
            Payment issue detected
          </p>
          <p className="text-sm text-amber-700 dark:text-amber-300">
            Please update your payment method to avoid service interruption.
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={onManageBilling}>
          Update Payment Method
        </Button>
      </div>
    </div>
  )
}
