'use client'

import { useState } from 'react'
import { Button, Badge, Card, CardContent, CardHeader, CardTitle } from '@kloudlite/ui'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import type { Subscription, Plan } from '@/lib/console/storage'

interface SubscriptionStatusProps {
  subscription: Subscription
  plan: Plan | undefined
  isOwner: boolean
  onCancel: () => Promise<void>
}

const statusColors: Record<string, string> = {
  active: 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20',
  created: 'bg-yellow-500/10 text-yellow-700 dark:text-yellow-400 border-yellow-500/20',
  authenticated: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20',
  paused: 'bg-orange-500/10 text-orange-700 dark:text-orange-400 border-orange-500/20',
  cancelled: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
  expired: 'bg-gray-500/10 text-gray-700 dark:text-gray-400 border-gray-500/20',
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

  const isTerminated = ['cancelled', 'expired'].includes(subscription.status)
  if (isTerminated) return null

  return (
    <Card className="border-foreground/10">
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Current Subscription</CardTitle>
          <Badge variant="outline" className={statusColors[subscription.status]}>
            {subscription.status}
          </Badge>
        </div>
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
            <p className="text-muted-foreground">Monthly Cost</p>
            <p className="font-medium">
              {plan
                ? `$${((plan.baseFee + plan.amountPerUser * subscription.quantity) / 100).toFixed(2)}`
                : '\u2014'}
            </p>
          </div>
          <div>
            <p className="text-muted-foreground">Next Billing</p>
            <p className="font-medium">
              {subscription.currentEnd
                ? new Date(subscription.currentEnd).toLocaleDateString()
                : '\u2014'}
            </p>
          </div>
        </div>

        {isOwner && (
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
