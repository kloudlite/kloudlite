'use client'

import { useState } from 'react'
import { Button, Badge, Card, CardContent, Input } from '@kloudlite/ui'
import { Loader2, Minus, Plus } from 'lucide-react'
import { cn } from '@kloudlite/lib'
import type { Plan } from '@/lib/console/storage'

type BillingPeriod = 'monthly' | 'annual'

interface SubscriptionConfiguratorProps {
  plans: Plan[]
  onSubscribe: (
    allocations: { planId: string; quantity: number }[],
    billingPeriod: BillingPeriod,
  ) => Promise<void>
}

export function SubscriptionConfigurator({ plans, onSubscribe }: SubscriptionConfiguratorProps) {
  const [quantities, setQuantities] = useState<Record<string, number>>(() => {
    const initial: Record<string, number> = {}
    for (const plan of plans) {
      initial[plan.id] = 0
    }
    return initial
  })
  const [billingPeriod, setBillingPeriod] = useState<BillingPeriod>('monthly')
  const [loading, setLoading] = useState(false)

  const baseFee = plans[0]?.baseFee ? plans[0].baseFee / 100 : 29
  const discountPct = plans[0]?.annualDiscountPct ?? 20
  const totalUsers = Object.values(quantities).reduce((sum, q) => sum + q, 0)
  const isAnnual = billingPeriod === 'annual'

  const sizeCosts = plans
    .filter((plan) => (quantities[plan.id] || 0) > 0)
    .map((plan) => {
      const qty = quantities[plan.id] || 0
      return {
        plan,
        quantity: qty,
        lineTotal: (plan.amountPerUser * qty) / 100,
      }
    })

  const userTotal = sizeCosts.reduce((sum, t) => sum + t.lineTotal, 0)
  const monthlyTotal = baseFee + userTotal
  const annualTotal = monthlyTotal * 12 * (1 - discountPct / 100)
  const displayTotal = isAnnual ? annualTotal : monthlyTotal

  const setQuantity = (planId: string, value: number) => {
    setQuantities((prev) => ({ ...prev, [planId]: Math.max(0, Math.min(100, value)) }))
  }

  const handleSubscribe = async () => {
    const allocations = plans
      .filter((plan) => (quantities[plan.id] || 0) > 0)
      .map((plan) => ({
        planId: plan.id,
        quantity: quantities[plan.id],
      }))
    if (allocations.length === 0) return
    setLoading(true)
    try {
      await onSubscribe(allocations, billingPeriod)
    } finally {
      setLoading(false)
    }
  }

  const periodLabel = isAnnual ? '/yr' : '/mo'

  return (
    <div className="space-y-8">
      {/* Billing Period Toggle */}
      <div className="flex items-center justify-center">
        <div className="inline-flex items-center rounded-lg border border-foreground/10 bg-muted/30 p-1">
          <button
            onClick={() => setBillingPeriod('monthly')}
            className={cn(
              'px-4 py-2 text-sm font-medium rounded-md transition-all',
              !isAnnual
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            Monthly
          </button>
          <button
            onClick={() => setBillingPeriod('annual')}
            className={cn(
              'px-4 py-2 text-sm font-medium rounded-md transition-all flex items-center gap-2',
              isAnnual
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            Annual
            <span className="text-[10px] font-semibold uppercase tracking-wider px-1.5 py-0.5 rounded bg-green-500/10 text-green-700 dark:text-green-400">
              Save {discountPct}%
            </span>
          </button>
        </div>
      </div>

      {/* Base Fee Banner */}
      <div className="border border-foreground/10 rounded-lg p-5 bg-muted/30">
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <Badge variant="outline" className="text-xs font-medium">
                Base Fee
              </Badge>
            </div>
            <h3 className="font-semibold text-foreground">Control Plane</h3>
            <p className="text-sm text-muted-foreground">Dashboard, user management, billing</p>
          </div>
          <div className="text-right">
            {isAnnual ? (
              <>
                <span className="text-sm text-muted-foreground line-through mr-2">
                  ₹{(baseFee * 12).toFixed(0)}
                </span>
                <span className="text-3xl font-bold text-foreground">
                  ₹{(baseFee * 12 * (1 - discountPct / 100)).toFixed(0)}
                </span>
                <span className="text-muted-foreground text-sm">/yr</span>
              </>
            ) : (
              <>
                <span className="text-3xl font-bold text-foreground">₹{baseFee}</span>
                <span className="text-muted-foreground text-sm">/mo</span>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Compute Size Cards with Quantity */}
      <div>
        <h3 className="text-sm font-medium text-foreground mb-3">Choose compute size per user</h3>
        <div className="space-y-3">
          {plans.map((plan) => {
            const qty = quantities[plan.id] || 0
            const unitPrice = plan.amountPerUser / 100
            return (
              <Card key={plan.id} className="border-foreground/10">
                <CardContent className="p-4">
                  <div className="flex items-start gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <h4 className="font-semibold text-foreground text-sm">{plan.name}</h4>
                        <span className="text-xs text-muted-foreground font-mono">
                          {plan.cpu} vCPU &middot; {plan.ram}
                        </span>
                      </div>
                      <div className="mt-1.5 flex flex-wrap gap-x-4 gap-y-0.5 text-xs text-muted-foreground">
                        <span>{plan.storage} storage</span>
                        <span>{plan.monthlyHours} hrs/mo</span>
                        <span>{plan.autoSuspend} auto-suspend</span>
                        <span>Overage: ₹{(plan.overageRate / 100).toFixed(2)}/hr</span>
                      </div>
                    </div>

                    <div className="flex items-center gap-4 shrink-0">
                      <div className="text-right">
                        {isAnnual ? (
                          <>
                            <span className="text-xs text-muted-foreground line-through mr-1">
                              ₹{(unitPrice * 12).toFixed(0)}
                            </span>
                            <span className="text-lg font-bold text-foreground">
                              ₹{(unitPrice * 12 * (1 - discountPct / 100)).toFixed(0)}
                            </span>
                            <span className="text-muted-foreground text-xs">/user/yr</span>
                          </>
                        ) : (
                          <>
                            <span className="text-lg font-bold text-foreground">₹{unitPrice}</span>
                            <span className="text-muted-foreground text-xs">/user/mo</span>
                          </>
                        )}
                      </div>

                      <div className="flex items-center gap-1.5">
                        <Button
                          type="button"
                          variant="outline"
                          size="icon"
                          className="h-8 w-8"
                          aria-label={`Decrease users for ${plan.name}`}
                          disabled={qty <= 0 || loading}
                          onClick={() => setQuantity(plan.id, qty - 1)}
                        >
                          <Minus className="h-3 w-3" />
                        </Button>
                        <Input
                          type="number"
                          min={0}
                          max={100}
                          value={qty}
                          onChange={(e) => setQuantity(plan.id, parseInt(e.target.value) || 0)}
                          className="h-8 w-14 text-center font-mono text-sm [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                          disabled={loading}
                        />
                        <Button
                          type="button"
                          variant="outline"
                          size="icon"
                          className="h-8 w-8"
                          aria-label={`Increase users for ${plan.name}`}
                          disabled={qty >= 100 || loading}
                          onClick={() => setQuantity(plan.id, qty + 1)}
                        >
                          <Plus className="h-3 w-3" />
                        </Button>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      </div>

      {/* Cost Summary */}
      <div className="border border-foreground/10 rounded-lg bg-background">
        <div className="p-5 space-y-3">
          <h3 className="text-sm font-medium text-foreground">
            {isAnnual ? 'Annual' : 'Monthly'} cost breakdown
          </h3>

          <div className="space-y-2 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Control Plane (base fee)</span>
              <span className="text-foreground">
                ₹{isAnnual ? (baseFee * 12 * (1 - discountPct / 100)).toFixed(2) : baseFee.toFixed(2)}
              </span>
            </div>
            {sizeCosts.map(({ plan, quantity: qty, lineTotal }) => {
              const displayLine = isAnnual
                ? lineTotal * 12 * (1 - discountPct / 100)
                : lineTotal
              return (
                <div key={plan.id} className="flex items-center justify-between">
                  <span className="text-muted-foreground">
                    {plan.name} &mdash; {qty} {qty === 1 ? 'user' : 'users'} &times; ₹
                    {isAnnual
                      ? ((plan.amountPerUser / 100) * 12 * (1 - discountPct / 100)).toFixed(0)
                      : plan.amountPerUser / 100}
                    {periodLabel}
                  </span>
                  <span className="text-foreground">₹{displayLine.toFixed(2)}</span>
                </div>
              )
            })}
            {totalUsers === 0 && (
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground italic">No users selected</span>
                <span className="text-foreground">₹0.00</span>
              </div>
            )}
          </div>

          {isAnnual && totalUsers > 0 && (
            <div className="flex items-center justify-between text-sm">
              <span className="text-green-700 dark:text-green-400">
                You save vs monthly
              </span>
              <span className="text-green-700 dark:text-green-400 font-medium">
                -₹{(monthlyTotal * 12 - annualTotal).toFixed(2)}
              </span>
            </div>
          )}

          <div className="border-t border-foreground/10 pt-3 flex items-center justify-between">
            <span className="font-medium text-foreground">
              Total {isAnnual ? 'per year' : 'per month'} ({totalUsers}{' '}
              {totalUsers === 1 ? 'user' : 'users'})
            </span>
            <span className="text-xl font-bold text-foreground">₹{displayTotal.toFixed(2)}</span>
          </div>
        </div>

        <div className="border-t border-foreground/10 p-5">
          <Button
            className="w-full"
            size="lg"
            disabled={totalUsers === 0 || loading}
            onClick={handleSubscribe}
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
                Processing...
              </>
            ) : (
              `Subscribe — ₹${displayTotal.toFixed(2)}${periodLabel}`
            )}
          </Button>
        </div>
      </div>
    </div>
  )
}
