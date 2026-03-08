'use client'

import { useState } from 'react'
import { Button, Input } from '@kloudlite/ui'
import { Loader2, Minus, Plus } from 'lucide-react'
import { formatCurrency } from '@/lib/billing-utils'
import type { SubscriptionItem } from '@/lib/console/storage'
import type { TierConfigItem } from '@/app/actions/billing/pricing'

interface SubscriptionConfiguratorProps {
  items: SubscriptionItem[]
  tierConfig: TierConfigItem[]
  currency: string
  onSave: (allocations: { priceId: string; quantity: number }[]) => Promise<void>
  onCancel?: () => void
  loading: boolean
  mode?: 'subscribe' | 'modify'
}

export function SubscriptionConfigurator({
  items,
  tierConfig,
  currency,
  onSave,
  onCancel,
  loading,
  mode = 'subscribe',
}: SubscriptionConfiguratorProps) {
  const isModify = mode === 'modify'

  const [quantities, setQuantities] = useState<Record<string, number>>(() => {
    const initial: Record<string, number> = {}
    for (const tier of tierConfig) {
      if (tier.fixed) {
        initial[tier.priceId] = 1
      } else {
        const existing = items.find((i) => i.tier === tier.tier)
        initial[tier.priceId] = existing?.quantity ?? 0
      }
    }
    return initial
  })

  const setQuantity = (priceId: string, value: number) => {
    setQuantities((prev) => ({ ...prev, [priceId]: Math.max(0, Math.min(100, value)) }))
  }

  const fixedTiers = tierConfig.filter((t) => t.fixed)
  const seatTiers = tierConfig.filter((t) => !t.fixed)
  const totalUsers = seatTiers.reduce((sum, t) => sum + (quantities[t.priceId] || 0), 0)

  const monthlyTotal = tierConfig.reduce((sum, t) => {
    return sum + t.pricePerUnit * (quantities[t.priceId] || 0)
  }, 0)

  const hasChanges = isModify
    ? seatTiers.some((t) => {
        const existing = items.find((i) => i.tier === t.tier)
        return (quantities[t.priceId] || 0) !== (existing?.quantity ?? 0)
      })
    : totalUsers > 0

  const handleSave = async () => {
    const allocations = tierConfig
      .filter((t) => {
        const qty = quantities[t.priceId] || 0
        if (qty > 0) return true
        // In modify mode, also include tiers set to 0 that previously had items
        // so the server knows to remove them from the subscription
        if (isModify && items.find((i) => i.tier === t.tier)) return true
        return false
      })
      .map((t) => ({
        priceId: t.priceId,
        quantity: quantities[t.priceId] || 0,
      }))
    if (allocations.length === 0) return
    await onSave(allocations)
  }

  return (
    <div>
      {/* Fixed items (Control Plane) */}
      {fixedTiers.map((tier) => (
        <div key={tier.priceId} className="border-b border-foreground/10 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-semibold text-sm text-foreground">{tier.name}</h3>
              <p className="text-sm text-muted-foreground">{tier.description}</p>
            </div>
            <span className="text-sm font-semibold tabular-nums text-foreground">
              {formatCurrency(tier.pricePerUnit, currency)}/mo
            </span>
          </div>
        </div>
      ))}

      {/* Seat tiers */}
      <div className="pt-4">
        <h3 className="text-sm font-medium text-foreground pb-2">Compute size per user</h3>
        <div className="divide-y divide-foreground/5">
          {seatTiers.map((tier) => {
            const qty = quantities[tier.priceId] || 0
            return (
              <div key={tier.priceId} className="flex items-center justify-between gap-4 py-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h4 className="font-semibold text-foreground text-sm">{tier.name}</h4>
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {formatCurrency(tier.pricePerUnit, currency)}/user/mo
                  </p>
                </div>

                <div className="flex items-center gap-1.5 shrink-0">
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    className="h-8 w-8"
                    aria-label={`Decrease users for ${tier.name}`}
                    disabled={qty <= 0 || loading}
                    onClick={() => setQuantity(tier.priceId, qty - 1)}
                  >
                    <Minus className="h-3 w-3" />
                  </Button>
                  <Input
                    type="number"
                    min={0}
                    max={100}
                    value={qty}
                    onChange={(e) => setQuantity(tier.priceId, parseInt(e.target.value) || 0)}
                    className="h-8 w-14 text-center font-mono text-sm [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                    disabled={loading}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    className="h-8 w-8"
                    aria-label={`Increase users for ${tier.name}`}
                    disabled={qty >= 100 || loading}
                    onClick={() => setQuantity(tier.priceId, qty + 1)}
                  >
                    <Plus className="h-3 w-3" />
                  </Button>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Summary */}
      <div className="border-t border-foreground/10 pt-4">
        <div className="rounded-lg bg-muted/30 p-4 space-y-3">
          <h3 className="text-sm font-medium text-foreground">Monthly cost breakdown</h3>

          <div className="space-y-2 text-sm">
            {fixedTiers.map((tier) => (
              <div key={tier.priceId} className="flex items-center justify-between">
                <span className="text-muted-foreground">{tier.name} (base fee)</span>
                <span className="text-foreground tabular-nums">
                  {formatCurrency(tier.pricePerUnit, currency)}
                </span>
              </div>
            ))}
            {seatTiers
              .filter((t) => (quantities[t.priceId] || 0) > 0)
              .map((tier) => {
                const qty = quantities[tier.priceId] || 0
                return (
                  <div key={tier.priceId} className="flex items-center justify-between">
                    <span className="text-muted-foreground">
                      {tier.name} &mdash; {qty} {qty === 1 ? 'user' : 'users'} &times;{' '}
                      {formatCurrency(tier.pricePerUnit, currency)}
                    </span>
                    <span className="text-foreground tabular-nums">
                      {formatCurrency(tier.pricePerUnit * qty, currency)}
                    </span>
                  </div>
                )
              })}
            {totalUsers === 0 && (
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground italic">No users selected</span>
                <span className="text-foreground tabular-nums">{formatCurrency(0, currency)}</span>
              </div>
            )}
          </div>

          <div className="border-t border-foreground/10 pt-3 flex items-center justify-between">
            <span className="font-medium text-foreground">
              Total per month ({totalUsers} {totalUsers === 1 ? 'user' : 'users'})
            </span>
            <span className="text-lg font-bold text-foreground tabular-nums">
              {formatCurrency(monthlyTotal, currency)}/mo
            </span>
          </div>

          {isModify && (
            <p className="text-xs text-muted-foreground">
              Prorated charges applied automatically by Stripe.
            </p>
          )}
        </div>
      </div>

      {/* Action Buttons */}
      <div className="pt-4 flex gap-3">
        {onCancel && (
          <Button
            variant="outline"
            size="lg"
            className="flex-1"
            onClick={onCancel}
            disabled={loading}
          >
            Cancel
          </Button>
        )}
        <Button
          className={onCancel ? 'flex-1' : 'w-full'}
          size="lg"
          disabled={!hasChanges || loading}
          onClick={handleSave}
        >
          {loading ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
              Processing...
            </>
          ) : isModify ? (
            'Apply Changes'
          ) : (
            `Subscribe — ${formatCurrency(monthlyTotal, currency)}/mo`
          )}
        </Button>
      </div>
    </div>
  )
}
