'use client'

import { useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { PlanCards } from '@/components/billing/plan-cards'
import { SubscriptionStatus } from '@/components/billing/subscription-status'
import { InvoiceHistory } from '@/components/billing/invoice-history'
import { createNewSubscription, cancelExistingSubscription } from '@/app/actions/billing'
import type { Plan, Subscription, Invoice } from '@/lib/console/storage'

declare global {
  interface Window {
    Razorpay: new (options: Record<string, unknown>) => { open: () => void }
  }
}

interface BillingContentProps {
  installationId: string
  plans: Plan[]
  subscription: Subscription | null
  invoices: Invoice[]
  isOwner: boolean
  userEmail: string
  userName: string
}

export function BillingContent({
  installationId,
  plans,
  subscription,
  invoices,
  isOwner,
  userEmail,
  userName,
}: BillingContentProps) {
  const router = useRouter()

  const handleSubscribe = useCallback(
    async (planId: string) => {
      try {
        const result = await createNewSubscription(installationId, planId, 1)

        const options = {
          key: process.env.NEXT_PUBLIC_RAZORPAY_KEY_ID,
          subscription_id: result.razorpaySubscriptionId,
          name: 'Kloudlite',
          description: 'Cloud Workspace Subscription',
          prefill: {
            email: userEmail,
            name: userName,
          },
          theme: {
            color: '#3B82F6',
          },
          handler: () => {
            toast.success('Subscription activated! Payment processing...')
            router.refresh()
          },
          modal: {
            ondismiss: () => {
              toast.info('Payment cancelled. You can try again anytime.')
            },
          },
        }

        const rzp = new window.Razorpay(options)
        rzp.open()
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to create subscription')
      }
    },
    [installationId, userEmail, userName, router],
  )

  const handleCancel = useCallback(async () => {
    await cancelExistingSubscription(installationId)
    router.refresh()
  }, [installationId, router])

  const activePlan = subscription ? plans.find((p) => p.id === subscription.planId) : undefined

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-xl font-semibold">Billing & Plans</h2>
        <p className="text-muted-foreground mt-1 text-base">
          Manage your subscription and payment methods
        </p>
      </div>

      {/* Active Subscription */}
      {subscription && (
        <SubscriptionStatus
          subscription={subscription}
          plan={activePlan}
          isOwner={isOwner}
          onCancel={handleCancel}
        />
      )}

      {/* Plan Selection */}
      <div>
        <h3 className="text-base font-medium mb-4">
          {subscription && !['cancelled', 'expired'].includes(subscription.status)
            ? 'Available Plans'
            : 'Choose a Plan'}
        </h3>
        <PlanCards
          plans={plans}
          subscription={subscription}
          installationId={installationId}
          isOwner={isOwner}
          onSubscribe={handleSubscribe}
        />
        <p className="text-muted-foreground text-xs mt-3 text-center">
          All plans include a ${plans[0] ? plans[0].baseFee / 100 : 29}/month base fee for the control plane
        </p>
      </div>

      {/* Invoice History */}
      <InvoiceHistory invoices={invoices} />

      {!isOwner && (
        <p className="text-muted-foreground text-sm text-center py-4">
          Only the installation owner can manage billing.
        </p>
      )}
    </div>
  )
}
