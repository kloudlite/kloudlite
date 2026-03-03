'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { Button, Input } from '@kloudlite/ui'
import { Loader2, Minus, Plus } from 'lucide-react'
import { cn } from '@kloudlite/lib'
import { previewModification } from '@/app/actions/billing'
import { getCurrencySymbol } from '@/lib/billing-utils'
import { getErrorMessage } from '@/lib/errors'
import type { Plan } from '@/lib/console/storage'

type BillingPeriod = 'monthly' | 'annual'

interface SubscriptionConfiguratorProps {
  plans: Plan[]
  onSubscribe: (
    allocations: { planId: string; quantity: number }[],
    billingPeriod: BillingPeriod,
  ) => Promise<void>
  initialQuantities?: Record<string, number>
  initialBillingPeriod?: BillingPeriod
  mode?: 'subscribe' | 'modify'
  onCancel?: () => void
  installationId?: string
  scheduledBillingPeriod?: 'monthly' | 'annual' | null
  currentEnd?: string | null
}

interface ProrationPreview {
  proratedAmount: number
  isUpgrade: boolean
  remainingDays: number
  scheduledDowngrade: boolean
}

export function SubscriptionConfigurator({
  plans,
  onSubscribe,
  initialQuantities,
  initialBillingPeriod,
  mode = 'subscribe',
  onCancel,
  installationId,
  scheduledBillingPeriod,
  currentEnd,
}: SubscriptionConfiguratorProps) {
  const [quantities, setQuantities] = useState<Record<string, number>>(() => {
    if (initialQuantities) return { ...initialQuantities }
    const initial: Record<string, number> = {}
    for (const plan of plans) {
      initial[plan.id] = 0
    }
    return initial
  })
  const [billingPeriod, setBillingPeriod] = useState<BillingPeriod>(initialBillingPeriod ?? 'monthly')
  const [loading, setLoading] = useState(false)
  const [proration, setProration] = useState<ProrationPreview | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewError, setPreviewError] = useState<string | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const isModifyMode = mode === 'modify'
  const hasScheduledDowngrade = !!scheduledBillingPeriod

  const periodUnchanged = !initialBillingPeriod || billingPeriod === initialBillingPeriod
  const quantitiesUnchanged = isModifyMode && initialQuantities
    ? plans.every((p) => (quantities[p.id] || 0) === (initialQuantities[p.id] || 0)) && periodUnchanged
    : false

  const fetchPreview = useCallback(async () => {
    if (!isModifyMode || !installationId || quantitiesUnchanged) {
      setProration(null)
      setPreviewError(null)
      return
    }
    const allocations = plans
      .map((p) => ({ planId: p.id, quantity: quantities[p.id] || 0 }))
      .filter((a) => a.quantity > 0)
    if (allocations.length === 0) {
      setProration(null)
      return
    }
    setPreviewLoading(true)
    setPreviewError(null)
    try {
      const newPeriod = initialBillingPeriod && billingPeriod !== initialBillingPeriod ? billingPeriod : undefined
      const result = await previewModification(installationId, allocations, newPeriod)
      setProration({
        proratedAmount: result.proratedAmount,
        isUpgrade: result.isUpgrade,
        remainingDays: result.remainingDays,
        scheduledDowngrade: result.scheduledDowngrade,
      })
    } catch (err) {
      setProration(null)
      setPreviewError(getErrorMessage(err, 'Failed to calculate proration'))
    } finally {
      setPreviewLoading(false)
    }
  }, [isModifyMode, installationId, quantities, plans, quantitiesUnchanged, billingPeriod, initialBillingPeriod])

  useEffect(() => {
    if (!isModifyMode) return
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(fetchPreview, 500)
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [fetchPreview, isModifyMode])

  const baseFee = plans[0]?.baseFee ? plans[0].baseFee / 100 : 29
  const discountPct = plans[0]?.annualDiscountPct ?? 20
  const cs = getCurrencySymbol(plans[0]?.currency)
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
    <div>
      {/* Billing Period Toggle — left-aligned */}
      <div className="pb-4">
        <div className="inline-flex items-center rounded-lg border border-foreground/10 bg-muted/30 p-1">
          <button
            onClick={() => setBillingPeriod('monthly')}
            disabled={hasScheduledDowngrade}
            className={cn(
              'px-4 py-2 text-sm font-medium rounded-md transition-all',
              !isAnnual
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground',
              hasScheduledDowngrade && 'opacity-50 cursor-not-allowed',
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

      {/* Control Plane — flat row */}
      <div className="border-t border-foreground/10 py-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-semibold text-sm text-foreground">Control Plane</h3>
            <p className="text-sm text-muted-foreground">Dashboard, user management, billing</p>
          </div>
          <div className="text-right">
            {isAnnual ? (
              <span className="text-sm font-semibold tabular-nums text-foreground">
                <span className="text-muted-foreground line-through font-normal mr-1.5">
                  {cs}{(baseFee * 12).toFixed(0)}
                </span>
                {cs}{(baseFee * 12 * (1 - discountPct / 100)).toFixed(0)}
                <span className="text-muted-foreground font-normal">/yr</span>
              </span>
            ) : (
              <span className="text-sm font-semibold tabular-nums text-foreground">
                {cs}{baseFee}
                <span className="text-muted-foreground font-normal">/mo</span>
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Compute Sizes — bare rows */}
      <div className="border-t border-foreground/10 pt-4">
        <h3 className="text-sm font-medium text-foreground pb-2">Compute size per user</h3>
        <div className="divide-y divide-foreground/5">
          {plans.map((plan) => {
            const qty = quantities[plan.id] || 0
            const unitPrice = plan.amountPerUser / 100
            return (
              <div key={plan.id} className="flex items-center justify-between gap-4 py-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h4 className="font-semibold text-foreground text-sm">{plan.name}</h4>
                    <span className="text-xs text-muted-foreground font-mono">
                      {plan.cpu} vCPU &middot; {plan.ram}
                    </span>
                  </div>
                  <div className="mt-1 flex flex-wrap gap-x-4 gap-y-0.5 text-xs text-muted-foreground">
                    <span>{plan.storage} storage</span>
                    <span>{plan.monthlyHours} hrs/mo</span>
                    <span>{plan.autoSuspend} auto-suspend</span>
                    <span>Overage: {cs}{(plan.overageRate / 100).toFixed(2)}/hr</span>
                  </div>
                </div>

                <div className="flex items-center gap-4 shrink-0">
                  <div className="text-right tabular-nums">
                    {isAnnual ? (
                      <>
                        <span className="text-xs text-muted-foreground line-through mr-1">
                          {cs}{(unitPrice * 12).toFixed(0)}
                        </span>
                        <span className="text-sm font-semibold text-foreground">
                          {cs}{(unitPrice * 12 * (1 - discountPct / 100)).toFixed(0)}
                        </span>
                        <span className="text-muted-foreground text-xs">/user/yr</span>
                      </>
                    ) : (
                      <>
                        <span className="text-sm font-semibold text-foreground">{cs}{unitPrice}</span>
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
            )
          })}
        </div>
      </div>

      {/* Summary — subtle bg, no border */}
      <div className="border-t border-foreground/10 pt-4">
        <div className="rounded-lg bg-muted/30 p-4 space-y-3">
          <h3 className="text-sm font-medium text-foreground">
            {isAnnual ? 'Annual' : 'Monthly'} cost breakdown
          </h3>

          <div className="space-y-2 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Control Plane (base fee)</span>
              <span className="text-foreground tabular-nums">
                {cs}{isAnnual ? (baseFee * 12 * (1 - discountPct / 100)).toFixed(2) : baseFee.toFixed(2)}
              </span>
            </div>
            {sizeCosts.map(({ plan, quantity: qty, lineTotal }) => {
              const displayLine = isAnnual
                ? lineTotal * 12 * (1 - discountPct / 100)
                : lineTotal
              return (
                <div key={plan.id} className="flex items-center justify-between">
                  <span className="text-muted-foreground">
                    {plan.name} &mdash; {qty} {qty === 1 ? 'user' : 'users'} &times; {cs}
                    {isAnnual
                      ? ((plan.amountPerUser / 100) * 12 * (1 - discountPct / 100)).toFixed(0)
                      : plan.amountPerUser / 100}
                    {periodLabel}
                  </span>
                  <span className="text-foreground tabular-nums">{cs}{displayLine.toFixed(2)}</span>
                </div>
              )
            })}
            {totalUsers === 0 && (
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground italic">No users selected</span>
                <span className="text-foreground tabular-nums">{cs}0.00</span>
              </div>
            )}
          </div>

          {isAnnual && totalUsers > 0 && (
            <div className="flex items-center justify-between text-sm">
              <span className="text-green-700 dark:text-green-400">
                You save vs monthly
              </span>
              <span className="text-green-700 dark:text-green-400 font-medium tabular-nums">
                -{cs}{(monthlyTotal * 12 - annualTotal).toFixed(2)}
              </span>
            </div>
          )}

          <div className="border-t border-foreground/10 pt-3 flex items-center justify-between">
            <span className="font-medium text-foreground">
              Total {isAnnual ? 'per year' : 'per month'} ({totalUsers}{' '}
              {totalUsers === 1 ? 'user' : 'users'})
            </span>
            <span className="text-lg font-bold text-foreground tabular-nums">
              {cs}{displayTotal.toFixed(2)}
            </span>
          </div>

          {/* Proration Preview (modify mode only) */}
          {isModifyMode && !quantitiesUnchanged && (
            <div className="border-t border-foreground/10 pt-3">
              {previewLoading ? (
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Loader2 className="h-3 w-3 animate-spin" />
                  Calculating proration...
                </div>
              ) : proration ? (
                <p className="text-sm text-muted-foreground">
                  {proration.scheduledDowngrade
                    ? `Your plan will switch to Monthly billing on ${currentEnd ? new Date(currentEnd).toLocaleDateString() : 'next renewal'}. No charge now.`
                    : proration.isUpgrade
                      ? `Prorated charge: ${cs}${(proration.proratedAmount / 100).toFixed(2)} for ${proration.remainingDays} remaining days`
                      : 'Downgrade applies immediately. No charge.'}
                </p>
              ) : previewError ? (
                <p className="text-sm text-destructive">{previewError}</p>
              ) : null}
            </div>
          )}
        </div>
      </div>

      {/* Action Buttons — standalone */}
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
          className={cn(onCancel ? 'flex-1' : 'w-full')}
          size="lg"
          disabled={
            totalUsers === 0 ||
            loading ||
            (isModifyMode && quantitiesUnchanged)
          }
          onClick={handleSubscribe}
        >
          {loading ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
              Processing...
            </>
          ) : isModifyMode ? (
            proration?.scheduledDowngrade
              ? 'Schedule Switch to Monthly'
              : proration?.isUpgrade
                ? `Upgrade — ${cs}${(proration.proratedAmount / 100).toFixed(2)} prorated`
                : 'Apply Changes'
          ) : (
            `Subscribe — ${cs}${displayTotal.toFixed(2)}${periodLabel}`
          )}
        </Button>
      </div>
    </div>
  )
}
