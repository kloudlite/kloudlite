'use client'

import { SubscriptionStatus } from '@/components/billing/subscription-status'
import type { Plan, Subscription } from '@/lib/console/storage'

interface PastSubscriptionsProps {
  subscriptions: Subscription[]
  plans: Plan[]
}

export function PastSubscriptions({ subscriptions, plans }: PastSubscriptionsProps) {
  const pastSubs = subscriptions.filter((s) => ['cancelled', 'expired'].includes(s.status))

  if (pastSubs.length === 0) return null

  return (
    <div className="mt-6 pt-6 border-t border-foreground/10 space-y-4">
      <h3 className="text-sm font-medium text-muted-foreground">Past Subscriptions</h3>
      {pastSubs.map((sub) => {
        const plan = plans.find((p) => p.id === sub.planId)
        return (
          <SubscriptionStatus
            key={sub.id}
            subscription={sub}
            plan={plan}
          />
        )
      })}
    </div>
  )
}
