import type { SubscriptionItem } from '@/lib/console/storage'
import type { TierConfigItem } from '@/app/actions/billing/pricing'
import { formatCurrency } from '@/lib/billing-utils'

interface SubscriptionStatusProps {
  items: SubscriptionItem[]
  tierConfig: TierConfigItem[]
  currency: string
}

export function SubscriptionStatus({ items, tierConfig, currency }: SubscriptionStatusProps) {
  const total = items.reduce((sum, item) => {
    const config = tierConfig.find((t) => t.tier === item.tier)
    const unitPrice = config?.pricePerUnit ?? 0
    return sum + unitPrice * item.quantity
  }, 0)

  return (
    <div className="rounded-lg border border-foreground/10 p-4">
      <h3 className="text-sm font-medium text-foreground mb-3">Current Products</h3>
      <div className="space-y-2">
        {items.map((item) => {
          const config = tierConfig.find((t) => t.tier === item.tier)
          const unitPrice = config?.pricePerUnit ?? 0
          return (
            <div key={item.id} className="flex items-center justify-between text-sm">
              <span className="text-foreground">{item.productName}</span>
              <span className="text-muted-foreground">
                {item.quantity}&times; &nbsp; {formatCurrency(unitPrice * item.quantity, currency)}/mo
              </span>
            </div>
          )
        })}
      </div>
      <div className="mt-3 pt-3 border-t border-foreground/10 flex justify-between text-sm font-medium">
        <span>Total</span>
        <span>{formatCurrency(total, currency)}/mo</span>
      </div>
    </div>
  )
}
