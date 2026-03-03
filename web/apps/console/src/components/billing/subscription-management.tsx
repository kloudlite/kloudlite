'use client'

import { useMemo } from 'react'
import { SubscriptionConfigurator } from '@/components/billing/subscription-configurator'
import { SubscriptionStatus } from '@/components/billing/subscription-status'
import { SubscriptionHeader } from '@/components/billing/subscription-header'
import { PaymentDueBanner } from '@/components/billing/payment-due-banner'
import { PastSubscriptions } from '@/components/billing/past-subscriptions'
import { useSubscriptionPayments } from '@/hooks/use-subscription-payments'
import type { Plan, Subscription, Invoice } from '@/lib/console/storage'

interface SubscriptionManagementProps {
  installationId: string
  plans: Plan[]
  subscriptions: Subscription[]
  invoices: Invoice[]
  isOwner: boolean
  userEmail: string
  userName: string
}

export function SubscriptionManagement({
  installationId,
  plans,
  subscriptions,
  invoices,
  isOwner,
  userEmail,
  userName,
}: SubscriptionManagementProps) {
  const activeSubs = useMemo(() => subscriptions.filter((s) =>
    ['active', 'authenticated', 'paused'].includes(s.status),
  ), [subscriptions])
  const visibleActiveSubs = activeSubs.filter((s) => s.quantity > 0)
  const hasActiveSubs = activeSubs.length > 0
  const pendingInvoice = invoices.find((i) => i.status === 'issued')
  const primarySub = visibleActiveSubs[0]

  const initialQuantities = useMemo(() => {
    const q: Record<string, number> = {}
    for (const plan of plans) {
      const sub = activeSubs.find((s) => s.planId === plan.id)
      q[plan.id] = sub?.quantity ?? 0
    }
    return q
  }, [plans, activeSubs])

  const {
    paying,
    editing,
    setEditing,
    handleSubscribe,
    handleModify,
    handlePayNow,
    handleCancelScheduledDowngrade,
  } = useSubscriptionPayments({
    installationId,
    userEmail,
    userName,
    activeSubs,
    pendingInvoice,
  })

  return (
    <>
      {hasActiveSubs && primarySub && !editing && (
        <SubscriptionHeader
          primarySub={primarySub}
          isOwner={isOwner}
          onModify={() => setEditing(true)}
        />
      )}

      {pendingInvoice && isOwner && (
        <PaymentDueBanner
          pendingInvoice={pendingInvoice}
          paying={paying}
          onPayNow={handlePayNow}
        />
      )}

      {hasActiveSubs && !editing && (
        <div className="space-y-6">
          {visibleActiveSubs.map((sub) => {
            const plan = plans.find((p) => p.id === sub.planId)
            return (
              <SubscriptionStatus
                key={sub.id}
                subscription={sub}
                plan={plan}
                onCancelScheduledDowngrade={
                  isOwner && sub.scheduledBillingPeriod ? handleCancelScheduledDowngrade : undefined
                }
              />
            )
          })}
        </div>
      )}

      {hasActiveSubs && editing && (
        <SubscriptionConfigurator
          plans={plans}
          onSubscribe={async (allocations, billingPeriod) => {
            await handleModify(allocations, billingPeriod)
          }}
          initialQuantities={initialQuantities}
          initialBillingPeriod={activeSubs[0].billingPeriod}
          mode="modify"
          onCancel={() => setEditing(false)}
          installationId={installationId}
          scheduledBillingPeriod={activeSubs[0].scheduledBillingPeriod}
          currentEnd={activeSubs[0].currentEnd}
        />
      )}

      {isOwner && !hasActiveSubs && (
        <SubscriptionConfigurator plans={plans} onSubscribe={handleSubscribe} />
      )}

      <PastSubscriptions subscriptions={subscriptions} plans={plans} />

      {!isOwner && (
        <p className="text-muted-foreground text-sm text-center py-4">
          Only the installation owner can manage billing.
        </p>
      )}
    </>
  )
}
