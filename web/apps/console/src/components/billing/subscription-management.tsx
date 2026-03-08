'use client'

import { useState } from 'react'
import { ExternalLink } from 'lucide-react'
import { Button } from '@kloudlite/ui'
import { SubscriptionConfigurator } from '@/components/billing/subscription-configurator'
import { SubscriptionStatus } from '@/components/billing/subscription-status'
import { PaymentWarningBanner } from '@/components/billing/payment-warning-banner'
import { useSubscriptionPayments } from '@/hooks/use-subscription-payments'
import type { BillingAccount, SubscriptionItem } from '@/lib/console/storage'
import type { TierConfigItem } from '@/app/actions/billing/pricing'

interface SubscriptionManagementProps {
  orgId: string
  customer: BillingAccount | null
  items: SubscriptionItem[]
  tierConfig: TierConfigItem[]
  currency: string
  isOwner: boolean
}

export function SubscriptionManagement({
  orgId,
  customer,
  items,
  tierConfig,
  currency,
  isOwner,
}: SubscriptionManagementProps) {
  const [editing, setEditing] = useState(false)
  const [confirmCancel, setConfirmCancel] = useState(false)
  const hasActiveSubscription = customer?.billingStatus === 'active'
  const isCancelled = customer?.billingStatus === 'cancelled'

  const {
    loading,
    handleSubscribe,
    handleModify,
    handleCancel,
    handleManageBilling,
  } = useSubscriptionPayments({ orgId })

  return (
    <div className="space-y-4">
      {/* Active subscription header */}
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

      {/* Cancelled subscription header */}
      {isCancelled && isOwner && (
        <div className="flex items-center justify-between">
          <div>
            <span className="inline-flex items-center gap-1.5 rounded-full bg-yellow-100 px-2.5 py-0.5 text-xs font-medium text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400">
              <span className="h-1.5 w-1.5 rounded-full bg-yellow-500" />
              Cancelling
            </span>
            {customer?.currentPeriodEnd && (
              <span className="ml-3 text-sm text-muted-foreground">
                Active until: {new Date(customer.currentPeriodEnd).toLocaleDateString()}
              </span>
            )}
          </div>
          <Button variant="outline" size="sm" onClick={handleManageBilling} disabled={loading}>
            Manage Billing <ExternalLink className="ml-1.5 h-3.5 w-3.5" />
          </Button>
        </div>
      )}

      {/* Payment warning */}
      {customer?.hasPaymentIssue && isOwner && (
        <PaymentWarningBanner onManageBilling={handleManageBilling} />
      )}

      {/* Current products */}
      {(hasActiveSubscription || isCancelled) && !editing && items.length > 0 && (
        <SubscriptionStatus items={items} tierConfig={tierConfig} currency={currency} />
      )}

      {/* Active subscription but no items synced yet */}
      {(hasActiveSubscription || isCancelled) && !editing && items.length === 0 && (
        <div className="rounded-lg border border-foreground/10 p-4">
          <p className="text-sm text-muted-foreground">
            Your subscription is active. Subscription details will appear here once they are synced.
          </p>
        </div>
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

      {/* Cancel subscription */}
      {hasActiveSubscription && isOwner && !editing && (
        <div className="border-t border-foreground/10 pt-4">
          {!confirmCancel ? (
            <Button
              variant="outline"
              size="sm"
              className="text-destructive hover:text-destructive"
              onClick={() => setConfirmCancel(true)}
              disabled={loading}
            >
              Cancel Subscription
            </Button>
          ) : (
            <div className="flex items-center gap-3">
              <p className="text-sm text-muted-foreground">
                Your subscription will remain active until the end of the current billing period.
              </p>
              <Button
                variant="destructive"
                size="sm"
                onClick={async () => {
                  await handleCancel()
                  setConfirmCancel(false)
                }}
                disabled={loading}
              >
                Confirm Cancel
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setConfirmCancel(false)}
                disabled={loading}
              >
                Keep Plan
              </Button>
            </div>
          )}
        </div>
      )}

      {/* No subscription — show subscribe form */}
      {isOwner && !hasActiveSubscription && !isCancelled && (
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
          Only the organization owner can manage billing.
        </p>
      )}
    </div>
  )
}
