'use client'

import { Button } from '@kloudlite/ui'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'
import { getCurrencySymbol } from '@/lib/billing-utils'
import type { Subscription, Plan } from '@/lib/console/storage'

interface SubscriptionStatusProps {
  subscription: Subscription
  plan: Plan | undefined
  onCancelScheduledDowngrade?: () => Promise<void>
}

export function SubscriptionStatus({ subscription, plan, onCancelScheduledDowngrade }: SubscriptionStatusProps) {
  const [cancelling, setCancelling] = useState(false)
  const isTerminated = ['cancelled', 'expired'].includes(subscription.status)
  const isAnnual = subscription.billingPeriod === 'annual'

  const cs = getCurrencySymbol(plan?.currency)
  const costDisplay = plan
    ? isAnnual
      ? `${cs}${((plan.amountPerUser * subscription.quantity) / 100 * 12 * (1 - (plan.annualDiscountPct ?? 20) / 100)).toFixed(2)}/yr`
      : `${cs}${((plan.amountPerUser * subscription.quantity) / 100).toFixed(2)}/mo`
    : '\u2014'

  const handleCancelScheduled = async () => {
    if (!onCancelScheduledDowngrade) return
    setCancelling(true)
    try {
      await onCancelScheduledDowngrade()
    } finally {
      setCancelling(false)
    }
  }

  return (
    <div className={isTerminated ? 'opacity-60' : ''}>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div>
          <p className="text-muted-foreground">Plan</p>
          <p className="font-medium">{plan?.name || 'Unknown'}</p>
        </div>
        <div>
          <p className="text-muted-foreground">Users</p>
          <p className="font-medium">{subscription.quantity}</p>
        </div>
        <div>
          <p className="text-muted-foreground">{isAnnual ? 'Annual Cost' : 'Monthly Cost'}</p>
          <p className="font-medium">{costDisplay}</p>
        </div>
        <div>
          <p className="text-muted-foreground">
            {isTerminated ? 'Ended' : 'Next Billing'}
          </p>
          <p className="font-medium">
            {subscription.currentEnd
              ? new Date(subscription.currentEnd).toLocaleDateString()
              : '\u2014'}
          </p>
        </div>
      </div>

      {subscription.scheduledBillingPeriod && !isTerminated && (
        <div className="mt-4 flex items-center justify-between gap-4 rounded-lg border border-blue-500/20 bg-blue-500/10 p-3">
          <p className="text-sm text-blue-800 dark:text-blue-200">
            Switching to Monthly billing on{' '}
            {subscription.currentEnd
              ? new Date(subscription.currentEnd).toLocaleDateString()
              : 'next renewal'}
            {!onCancelScheduledDowngrade && (
              <span className="text-blue-600 dark:text-blue-300"> — managed by the installation owner</span>
            )}
          </p>
          {onCancelScheduledDowngrade && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleCancelScheduled}
              disabled={cancelling}
              className="shrink-0"
            >
              {cancelling ? (
                <>
                  <Loader2 className="h-3 w-3 animate-spin mr-1" />
                  Cancelling...
                </>
              ) : (
                'Cancel'
              )}
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
