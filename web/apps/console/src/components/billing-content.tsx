'use client'

import { useCallback, useState } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { SubscriptionConfigurator } from '@/components/billing/subscription-configurator'
import { SubscriptionStatus } from '@/components/billing/subscription-status'
import { InvoiceHistory } from '@/components/billing/invoice-history'
import { AlertTriangle, Loader2 } from 'lucide-react'
import {
  getRazorpayKey,
  createInstallationOrder,
  verifyPaymentAndActivate,
  cancelExistingSubscription,
} from '@/app/actions/billing'
import { useRazorpay } from '@/components/razorpay-provider'
import type { Plan, Subscription, Invoice } from '@/lib/console/storage'

interface BillingContentProps {
  installationId: string
  plans: Plan[]
  subscriptions: Subscription[]
  invoices: Invoice[]
  isOwner: boolean
  userEmail: string
  userName: string
}

export function BillingContent({
  installationId,
  plans,
  subscriptions,
  invoices,
  isOwner,
  userEmail,
  userName,
}: BillingContentProps) {
  const router = useRouter()
  const { openCheckout } = useRazorpay()
  const [paying, setPaying] = useState(false)

  const activeSubs = subscriptions.filter(
    (s) => ['active', 'authenticated', 'paused'].includes(s.status),
  )
  const pastSubs = subscriptions.filter((s) =>
    ['cancelled', 'expired'].includes(s.status),
  )
  const hasActiveSubs = activeSubs.length > 0
  const pendingInvoice = invoices.find((i) => i.status === 'issued')

  const handlePayNow = useCallback(async () => {
    if (!pendingInvoice?.razorpayInvoiceId || paying) return
    setPaying(true)

    const key = await getRazorpayKey()

    const options = {
      key: key,
      order_id: pendingInvoice.razorpayInvoiceId,
      amount: pendingInvoice.amount,
      currency: pendingInvoice.currency,
      name: 'Kloudlite',
      description: 'Subscription Renewal',
      prefill: {
        email: userEmail,
        name: userName,
      },
      theme: {
        color: '#3B82F6',
      },
      handler: async (response: Record<string, string>) => {
        try {
          await verifyPaymentAndActivate(
            installationId,
            response.razorpay_order_id,
            response.razorpay_payment_id,
            response.razorpay_signature,
          )
          toast.success('Payment successful! Subscription renewed.')
          router.refresh()
        } catch {
          toast.error('Payment verification failed. Please contact support.')
        }
      },
      modal: {
        ondismiss: () => {
          setPaying(false)
          toast.info('Payment cancelled. You can try again anytime.')
        },
      },
    }

    openCheckout(options)
  }, [pendingInvoice, installationId, userEmail, userName, router, openCheckout, paying])

  const handleSubscribe = useCallback(
    async (
      allocations: { planId: string; quantity: number }[],
      billingPeriod: 'monthly' | 'annual',
    ) => {
      try {
        const order = await createInstallationOrder(installationId, allocations, billingPeriod)

        // Load Razorpay key
        const key = await getRazorpayKey()

        const totalUsers = allocations.reduce((sum, a) => sum + a.quantity, 0)
        const periodLabel = billingPeriod === 'annual' ? 'Annual' : 'Monthly'
        const options = {
          key: key,
          order_id: order.razorpayOrderId,
          amount: order.amount,
          currency: order.currency,
          name: 'Kloudlite',
          description: `${totalUsers} ${totalUsers === 1 ? 'user' : 'users'} — ${periodLabel} — Kloudlite Cloud`,
          prefill: {
            email: userEmail,
            name: userName,
          },
          theme: {
            color: '#3B82F6',
          },
          handler: async (response: Record<string, string>) => {
            try {
              await verifyPaymentAndActivate(
                installationId,
                response.razorpay_order_id,
                response.razorpay_payment_id,
                response.razorpay_signature,
              )
              toast.success('Payment successful!')
              router.refresh()
            } catch {
              toast.error('Payment verification failed. Please contact support.')
            }
          },
          modal: {
            ondismiss: () => {
              toast.info('Payment cancelled. You can try again anytime.')
            },
          },
        }

        openCheckout(options)
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to create order')
      }
    },
    [installationId, userEmail, userName, router, openCheckout],
  )

  const handleCancel = useCallback(async () => {
    await cancelExistingSubscription(installationId)
    router.refresh()
  }, [installationId, router])

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-xl font-semibold">Billing & Compute</h2>
        <p className="text-muted-foreground mt-1 text-base">
          Manage your subscriptions and payment methods
        </p>
      </div>

      {/* Payment Due Banner */}
      {pendingInvoice && isOwner && (
        <div className="flex items-center justify-between gap-4 rounded-lg border border-amber-500/20 bg-amber-500/10 p-4">
          <div className="flex items-center gap-3">
            <AlertTriangle className="h-5 w-5 text-amber-600 dark:text-amber-400 shrink-0" />
            <div>
              <p className="text-sm font-medium text-amber-800 dark:text-amber-200">
                Payment Due
              </p>
              <p className="text-xs text-amber-700 dark:text-amber-300 mt-0.5">
                ₹{(pendingInvoice.amount / 100).toFixed(2)}{' '}
                due for your subscription renewal
              </p>
            </div>
          </div>
          <button
            onClick={handlePayNow}
            disabled={paying}
            className="shrink-0 rounded-md bg-amber-600 px-4 py-2 text-sm font-medium text-white hover:bg-amber-700 transition-colors disabled:opacity-50"
          >
            {paying ? (
              <span className="flex items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                Processing...
              </span>
            ) : (
              'Pay Now'
            )}
          </button>
        </div>
      )}

      {/* Active Subscriptions */}
      {hasActiveSubs &&
        activeSubs.map((sub) => {
          const plan = plans.find((p) => p.id === sub.planId)
          return (
            <SubscriptionStatus
              key={sub.id}
              subscription={sub}
              plan={plan}
              isOwner={isOwner}
              onCancel={handleCancel}
            />
          )
        })}

      {/* Subscribe / Add Compute */}
      {isOwner && !hasActiveSubs && (
        <SubscriptionConfigurator plans={plans} onSubscribe={handleSubscribe} />
      )}

      {/* Past Subscriptions */}
      {pastSubs.length > 0 && (
        <div className="space-y-4">
          <h3 className="text-sm font-medium text-muted-foreground">Past Subscriptions</h3>
          {pastSubs.map((sub) => {
            const plan = plans.find((p) => p.id === sub.planId)
            return (
              <SubscriptionStatus
                key={sub.id}
                subscription={sub}
                plan={plan}
                isOwner={isOwner}
                onCancel={handleCancel}
              />
            )
          })}
        </div>
      )}

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
