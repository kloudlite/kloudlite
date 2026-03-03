'use client'

import { Badge, Button } from '@kloudlite/ui'
import { Pencil } from 'lucide-react'
import type { Subscription } from '@/lib/console/storage'

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

interface SubscriptionHeaderProps {
  primarySub: Subscription
  isOwner: boolean
  onModify: () => void
}

export function SubscriptionHeader({ primarySub, isOwner, onModify }: SubscriptionHeaderProps) {
  const status = getDisplayStatus(primarySub)

  return (
    <div className="flex items-center justify-between mb-6">
      <div className="flex items-center gap-2">
        <Badge variant="outline" className="text-xs">
          {primarySub.billingPeriod === 'annual' ? 'Annual' : 'Monthly'}
        </Badge>
        <Badge variant="outline" className={statusColors[status]}>
          {statusLabels[status]}
        </Badge>
      </div>
      {isOwner && (
        <Button variant="outline" size="sm" onClick={onModify} className="gap-2">
          <Pencil className="h-3.5 w-3.5" />
          Modify Plan
        </Button>
      )}
    </div>
  )
}
