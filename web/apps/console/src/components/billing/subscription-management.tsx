'use client'

import { useState } from 'react'
import { ExternalLink } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { SubscriptionConfigurator } from '@/components/billing/subscription-configurator'
import { SubscriptionStatus } from '@/components/billing/subscription-status'
import { PaymentWarningBanner } from '@/components/billing/payment-warning-banner'
import { useSubscriptionPayments } from '@/hooks/use-subscription-payments'
import type { StripeCustomer, SubscriptionItem } from '@/lib/console/storage'
import type { TierConfigItem } from '@/app/actions/billing/pricing'

interface SubscriptionManagementProps {
  installationId: string
  customer: StripeCustomer | null
  items: SubscriptionItem[]
  tierConfig: TierConfigItem[]
  currency: string
  isOwner: boolean
}

export function SubscriptionManagement({
  installationId,
  customer,
  items,
  tierConfig,
  currency,
  isOwner,
}: SubscriptionManagementProps) {
  const [editing, setEditing] = useState(false)
  const hasActiveSubscription = customer?.billingStatus === 'active'

  const {
    loading,
    handleSubscribe,
    handleModify,
    handleManageBilling,
  } = useSubscriptionPayments({ installationId })

  return (
    <div className="space-y-4">
      {/* Header with Manage Billing button */}
      {hasActiveSubscription && isOwner && (
        <div className="flex items-center justify-between">
          <div>
            <span className="inline-flex items-center gap-1.5 rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">
              <span className="h-1.5 w-1.5 rounded-full bg-green-500" />
              Active
            </span>
            {customer?.currentPeriodEnd && (
              <span className="ml-3 text-sm text-muted-foreground">
                Next billing: {new Date(customer.currentPeriodEnd).toLocaleDateString()}
              </span>
            )}
          </div>
          <div className="flex gap-2">
            {!editing && (
              <Button variant="outline" size="sm" onClick={() => setEditing(true)}>
                Modify Plan
              </Button>
            )}
            <Button variant="outline" size="sm" onClick={handleManageBilling} disabled={loading}>
              Manage Billing <ExternalLink className="ml-1.5 h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      )}

      {/* Payment warning */}
      {customer?.paymentIssue && isOwner && (
        <PaymentWarningBanner onManageBilling={handleManageBilling} />
      )}

      {/* Current products */}
      {hasActiveSubscription && !editing && items.length > 0 && (
        <SubscriptionStatus items={items} tierConfig={tierConfig} currency={currency} />
      )}

      {/* Modify plan */}
      {hasActiveSubscription && editing && (
        <SubscriptionConfigurator
          items={items}
          tierConfig={tierConfig}
          currency={currency}
          onSave={async (modifications) => {
            await handleModify(modifications)
            setEditing(false)
          }}
          onCancel={() => setEditing(false)}
          loading={loading}
          mode="modify"
        />
      )}

      {/* No subscription — show subscribe form */}
      {isOwner && !hasActiveSubscription && (
        <SubscriptionConfigurator
          items={[]}
          tierConfig={tierConfig}
          currency={currency}
          onSave={handleSubscribe}
          loading={loading}
          mode="subscribe"
        />
      )}

      {!isOwner && (
        <p className="text-muted-foreground text-sm text-center py-4">
          Only the installation owner can manage billing.
        </p>
      )}
    </div>
  )
}
