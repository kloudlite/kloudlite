'use client'

import { useState } from 'react'
import { Button, Badge, Card, CardContent, CardHeader, CardTitle } from '@kloudlite/ui'
import { AlertTriangle, Clock, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import type { Subscription, Plan } from '@/lib/console/storage'

interface SubscriptionStatusProps {
  subscription: Subscription
  plan: Plan | undefined
  isOwner: boolean
  onCancel: () => Promise<void>
}

type DisplayStatus =
  | 'active'
  | 'expiring_soon'
  | 'created'
  | 'authenticated'
  | 'paused'
  | 'cancelled'
  | 'expired'

const statusConfig: Record<DisplayStatus, { label: string; color: string }> = {
  active: {
    label: 'Active',
    color: 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20',
  },
  expiring_soon: {
    label: 'Expiring Soon',
    color: 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20',
  },
  created: {
    label: 'Pending Payment',
    color: 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border-yellow-500/20',
  },
  authenticated: {
    label: 'Authenticating',
    color: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20',
  },
  paused: {
    label: 'Paused',
    color: 'bg-orange-500/10 text-orange-700 dark:text-orange-400 border-orange-500/20',
  },
  cancelled: {
    label: 'Cancelled',
    color: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
  },
  expired: {
    label: 'Expired',
    color: 'bg-gray-500/10 text-gray-700 dark:text-gray-400 border-gray-500/20',
  },
}

function getDisplayStatus(subscription: Subscription): {
  status: DisplayStatus
  daysUntilEnd: number | null
} {
  if (subscription.status === 'active' && subscription.currentEnd) {
    const msUntilEnd = new Date(subscription.currentEnd).getTime() - Date.now()
    const daysUntilEnd = Math.ceil(msUntilEnd / (24 * 60 * 60 * 1000))
    if (daysUntilEnd <= 7 && daysUntilEnd > 0) {
      return { status: 'expiring_soon', daysUntilEnd }
    }
    return { status: 'active', daysUntilEnd: daysUntilEnd > 0 ? daysUntilEnd : null }
  }
  return { status: subscription.status as DisplayStatus, daysUntilEnd: null }
}

export function SubscriptionStatus({ subscription, plan, isOwner, onCancel }: SubscriptionStatusProps) {
  const [cancelling, setCancelling] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)

  const handleCancel = async () => {
    setCancelling(true)
    try {
      await onCancel()
      toast.success('Subscription cancelled')
      setShowConfirm(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to cancel subscription')
    } finally {
      setCancelling(false)
    }
  }

  const { status: displayStatus, daysUntilEnd } = getDisplayStatus(subscription)
  const config = statusConfig[displayStatus]
  const isTerminated = ['cancelled', 'expired'].includes(subscription.status)
  const isAnnual = subscription.billingPeriod === 'annual'
  const periodLabel = isAnnual ? 'Annual' : 'Monthly'

  const costDisplay = plan
    ? isAnnual
      ? `₹${((plan.amountPerUser * subscription.quantity) / 100 * 12 * (1 - (plan.annualDiscountPct ?? 20) / 100)).toFixed(2)}/yr`
      : `₹${((plan.amountPerUser * subscription.quantity) / 100).toFixed(2)}/mo`
    : '\u2014'

  return (
    <Card className={isTerminated ? 'border-foreground/5 opacity-60' : 'border-foreground/10'}>
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">
            {isTerminated ? 'Past Subscription' : 'Current Subscription'}
          </CardTitle>
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="text-xs">
              {periodLabel}
            </Badge>
            <Badge variant="outline" className={config.color}>
              {config.label}
            </Badge>
          </div>
        </div>
        {displayStatus === 'expiring_soon' && daysUntilEnd !== null && (
          <div className="flex items-center gap-1.5 mt-2 text-sm text-amber-700 dark:text-amber-400">
            <Clock className="h-3.5 w-3.5" />
            <span>Expires in {daysUntilEnd} {daysUntilEnd === 1 ? 'day' : 'days'}</span>
          </div>
        )}
      </CardHeader>
      <CardContent>
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

        {isOwner && !isTerminated && (
          <div className="mt-6 pt-4 border-t border-foreground/10">
            {!showConfirm ? (
              <Button
                variant="outline"
                size="sm"
                className="text-red-600 dark:text-red-400 border-red-500/20 hover:bg-red-500/10"
                onClick={() => setShowConfirm(true)}
              >
                Cancel Subscription
              </Button>
            ) : (
              <div className="flex items-center gap-3 p-3 bg-red-500/5 border border-red-500/20 rounded-md">
                <AlertTriangle className="h-4 w-4 text-red-500 flex-shrink-0" />
                <p className="text-sm text-red-700 dark:text-red-400 flex-1">
                  Are you sure? This will cancel your subscription at the end of the current billing period.
                </p>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" onClick={() => setShowConfirm(false)} disabled={cancelling}>
                    Keep
                  </Button>
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={handleCancel}
                    disabled={cancelling}
                  >
                    {cancelling && <Loader2 className="h-3 w-3 animate-spin mr-1" />}
                    Confirm Cancel
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
