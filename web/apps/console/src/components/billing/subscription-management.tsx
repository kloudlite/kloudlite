'use client'

import { useCallback, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Button, Badge } from '@kloudlite/ui'
import { SubscriptionConfigurator } from '@/components/billing/subscription-configurator'
import { SubscriptionStatus } from '@/components/billing/subscription-status'
import { formatCurrency } from '@/lib/billing-utils'
import { AlertTriangle, Loader2, Pencil } from 'lucide-react'
import {
  getRazorpayKey,
  createInstallationOrder,
  verifyPaymentAndActivate,
  modifySubscriptionQuantities,
  verifyModificationAndApply,
  cancelScheduledDowngrade,
} from '@/app/actions/billing'
import { useRazorpay } from '@/components/razorpay-provider'
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

const statusColors: Record<string, string> = {
  active: 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20',
  expiring_soon: 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20',
  created: 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border-yellow-500/20',
  authenticated: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20',
  paused: 'bg-orange-500/10 text-orange-700 dark:text-orange-400 border-orange-500/20',
  cancelled: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
  expired: 'bg-gray-500/10 text-gray-700 dark:text-gray-400 border-gray-500/20',
}

const statusLabels: Record<string, string> = {
  active: 'Active',
  expiring_soon: 'Expiring Soon',
  created: 'Pending Payment',
  authenticated: 'Authenticating',
  paused: 'Paused',
  cancelled: 'Cancelled',
  expired: 'Expired',
}

function getDisplayStatus(sub: Subscription) {
  if (sub.status === 'active' && sub.currentEnd) {
    const days = Math.ceil((new Date(sub.currentEnd).getTime() - Date.now()) / 86400000)
    if (days <= 7 && days > 0) return 'expiring_soon'
  }
  return sub.status
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
  const router = useRouter()
  const { openCheckout } = useRazorpay()
  const [paying, setPaying] = useState(false)
  const [editing, setEditing] = useState(false)

  const activeSubs = subscriptions.filter((s) =>
    ['active', 'authenticated', 'paused'].includes(s.status),
  )
  const visibleActiveSubs = activeSubs.filter((s) => s.quantity > 0)
  const pastSubs = subscriptions.filter((s) => ['cancelled', 'expired'].includes(s.status))
  const hasActiveSubs = activeSubs.length > 0
  const pendingInvoice = invoices.find((i) => i.status === 'issued')

  // For the header badges, use the first visible active sub
  const primarySub = visibleActiveSubs[0]
  const primaryStatus = primarySub ? getDisplayStatus(primarySub) : null

  const initialQuantities = useMemo(() => {
    const q: Record<string, number> = {}
    for (const plan of plans) {
      const sub = activeSubs.find((s) => s.planId === plan.id)
      q[plan.id] = sub?.quantity ?? 0
    }
    return q
  }, [plans, activeSubs])

  const handleCancelScheduledDowngrade = useCallback(async () => {
    try {
      await cancelScheduledDowngrade(installationId)
      toast.success('Scheduled downgrade cancelled.')
      router.refresh()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to cancel scheduled downgrade')
    }
  }, [installationId, router])

  const handleModify = useCallback(
    async (allocations: { planId: string; quantity: number }[], billingPeriod: 'monthly' | 'annual') => {
      try {
        const currentPeriod = activeSubs[0]?.billingPeriod ?? 'monthly'
        const newPeriod = billingPeriod !== currentPeriod ? billingPeriod : undefined
        const result = await modifySubscriptionQuantities(installationId, allocations, newPeriod)

        if (result.applied) {
          const message = 'scheduled' in result && result.scheduled
            ? 'Monthly billing scheduled for end of current period.'
            : 'Subscription updated successfully.'
          toast.success(message)
          setEditing(false)
          router.refresh()
          return
        }

        const key = await getRazorpayKey()
        const options = {
          key,
          order_id: result.razorpayOrderId,
          amount: result.amount,
          currency: result.currency,
          name: 'Kloudlite',
          description: 'Subscription Upgrade (prorated)',
          prefill: { email: userEmail, name: userName },
          theme: { color: '#3B82F6' },
          handler: async (response: Record<string, string>) => {
            try {
              await verifyModificationAndApply(
                installationId,
                response.razorpay_order_id,
                response.razorpay_payment_id,
                response.razorpay_signature,
              )
              toast.success('Upgrade successful! Quantities updated.')
              setEditing(false)
              router.refresh()
            } catch {
              toast.error('Payment verification failed. Please contact support.')
            }
          },
          modal: {
            ondismiss: () => {
              toast.info('Payment cancelled. No changes were made.')
            },
          },
        }
        openCheckout(options)
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to modify subscription')
      }
    },
    [installationId, userEmail, userName, router, openCheckout, activeSubs],
  )

  const handlePayNow = useCallback(async () => {
    if (!pendingInvoice?.razorpayInvoiceId || paying) return
    setPaying(true)

    const key = await getRazorpayKey()
    const options = {
      key,
      order_id: pendingInvoice.razorpayInvoiceId,
      amount: pendingInvoice.amount,
      currency: pendingInvoice.currency,
      name: 'Kloudlite',
      description: 'Subscription Renewal',
      prefill: { email: userEmail, name: userName },
      theme: { color: '#3B82F6' },
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
        const key = await getRazorpayKey()

        const totalUsers = allocations.reduce((sum, a) => sum + a.quantity, 0)
        const periodLabel = billingPeriod === 'annual' ? 'Annual' : 'Monthly'
        const options = {
          key,
          order_id: order.razorpayOrderId,
          amount: order.amount,
          currency: order.currency,
          name: 'Kloudlite',
          description: `${totalUsers} ${totalUsers === 1 ? 'user' : 'users'} — ${periodLabel} — Kloudlite Cloud`,
          prefill: { email: userEmail, name: userName },
          theme: { color: '#3B82F6' },
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

  return (
    <>
      {/* Card header — rendered by parent, but we add the action row here */}
      {hasActiveSubs && primarySub && primaryStatus && !editing && (
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="text-xs">
              {primarySub.billingPeriod === 'annual' ? 'Annual' : 'Monthly'}
            </Badge>
            <Badge variant="outline" className={statusColors[primaryStatus]}>
              {statusLabels[primaryStatus]}
            </Badge>
          </div>
          {isOwner && (
            <Button variant="outline" size="sm" onClick={() => setEditing(true)} className="gap-2">
              <Pencil className="h-3.5 w-3.5" />
              Modify Plan
            </Button>
          )}
        </div>
      )}

      {/* Payment Due Banner */}
      {pendingInvoice && isOwner && (
        <div className="flex items-center justify-between gap-4 rounded-lg border border-amber-500/20 bg-amber-500/10 p-4 mb-6">
          <div className="flex items-center gap-3">
            <AlertTriangle className="h-5 w-5 text-amber-600 dark:text-amber-400 shrink-0" />
            <div>
              <p className="text-sm font-medium text-amber-800 dark:text-amber-200">Payment Due</p>
              <p className="text-xs text-amber-700 dark:text-amber-300 mt-0.5">
                {formatCurrency(pendingInvoice.amount, pendingInvoice.currency)} due for your subscription renewal
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

      {/* Active Subscriptions — details grid */}
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

      {/* Edit Mode — Modify Quantities */}
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

      {/* Subscribe / Add Compute */}
      {isOwner && !hasActiveSubs && (
        <SubscriptionConfigurator plans={plans} onSubscribe={handleSubscribe} />
      )}

      {/* Past Subscriptions */}
      {pastSubs.length > 0 && (
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
      )}

      {!isOwner && (
        <p className="text-muted-foreground text-sm text-center py-4">
          Only the installation owner can manage billing.
        </p>
      )}
    </>
  )
}
