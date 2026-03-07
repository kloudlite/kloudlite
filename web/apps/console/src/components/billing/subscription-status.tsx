import type { SubscriptionItem } from '@/lib/console/storage'
import { formatCurrency } from '@/lib/billing-utils'

// Price per unit per month in cents (matching Stripe prices)
const TIER_PRICES: Record<number, number> = {
  0: 2900, // Control Plane $29
  1: 2900, // Tier 1 $29/user
  2: 4900, // Tier 2 $49/user
  3: 8900, // Tier 3 $89/user
}

interface SubscriptionStatusProps {
  items: SubscriptionItem[]
}

export function SubscriptionStatus({ items }: SubscriptionStatusProps) {
  const total = items.reduce((sum, item) => {
    const unitPrice = TIER_PRICES[item.tier] ?? 0
    return sum + unitPrice * item.quantity
  }, 0)

  return (
    <div className="rounded-lg border border-foreground/10 p-4">
      <h3 className="text-sm font-medium text-foreground mb-3">Current Products</h3>
      <div className="space-y-2">
        {items.map((item) => {
          const unitPrice = TIER_PRICES[item.tier] ?? 0
          return (
            <div key={item.id} className="flex items-center justify-between text-sm">
              <span className="text-foreground">{item.productName}</span>
              <span className="text-muted-foreground">
                {item.quantity}&times; &nbsp; {formatCurrency(unitPrice * item.quantity, 'USD')}/mo
              </span>
            </div>
          )
        })}
      </div>
      <div className="mt-3 pt-3 border-t border-foreground/10 flex justify-between text-sm font-medium">
        <span>Total</span>
        <span>{formatCurrency(total, 'USD')}/mo</span>
      </div>
    </div>
  )
}
