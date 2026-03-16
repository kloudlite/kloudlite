'use client'

import { useState, useEffect, useMemo, useRef } from 'react'
import {
  Button,
  Collapsible,
  CollapsibleTrigger,
  CollapsibleContent,
  Label,
  Slider,
} from '@kloudlite/ui'
import { cn } from '@kloudlite/lib'
import { Loader2, Wallet, Server, Calculator, ChevronDown, Minus, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { useCredits } from '@/hooks/use-credits'
import type { PricingTier } from '@/lib/console/storage/credits-types'

function formatDollars(amount: number): string {
  return `$${amount.toFixed(2)}`
}

function hourlyToMonthly(hourly: number): number {
  return hourly * 24 * 30
}

interface KlCloudTabProps {
  orgId: string
  creating: boolean
  subdomainAvailable: boolean | null
  onSubmit: () => void
  existingInstallationId?: string
  getFormValues?: () => { name: string; subdomain: string }
  checkoutSessionId?: string
}

export function KlCloudTab({
  orgId,
  creating,
  subdomainAvailable,
  onSubmit,
  existingInstallationId,
  getFormValues,
  checkoutSessionId,
}: KlCloudTabProps) {
  const isSubscribeOnly = !!existingInstallationId
  const [pricingTiers, setPricingTiers] = useState<PricingTier[]>([])
  const [pricingLoading, setPricingLoading] = useState(true)
  const [calculatorOpen, setCalculatorOpen] = useState(false)
  const [userCounts, setUserCounts] = useState<Record<string, number>>({})
  const [workingHoursPerDay, setWorkingHoursPerDay] = useState(8)

  const { loading: creditsLoading, data: creditsData, handleTopup, refresh: refreshCredits } = useCredits(orgId, { skipInitialFetch: !!checkoutSessionId })

  // Verify checkout session and credit account on return from Stripe
  const verified = useRef(false)
  useEffect(() => {
    if (!checkoutSessionId || verified.current) return
    verified.current = true

    fetch(`/api/orgs/${orgId}/billing/topup?session_id=${checkoutSessionId}`)
      .then((res) => res.json())
      .then((data) => {
        if (data.status === 'credited') {
          toast.success(`$${data.amount.toFixed(2)} credits added to your account`)
        }
        refreshCredits()
      })
      .catch(() => refreshCredits())
  }, [checkoutSessionId, orgId, refreshCredits])

  useEffect(() => {
    async function fetchPricing() {
      try {
        const res = await fetch('/api/pricing')
        if (!res.ok) throw new Error('Failed to fetch pricing')
        const json = await res.json()
        setPricingTiers(json.tiers ?? [])
      } catch (err) {
        console.error('Failed to fetch pricing tiers:', err)
        toast.error('Failed to load pricing information')
      } finally {
        setPricingLoading(false)
      }
    }
    fetchPricing()
  }, [])

  const controlPlaneTier = useMemo(
    () => pricingTiers.find((t) => t.resourceType === 'controlplane'),
    [pricingTiers],
  )

  const computeTiers = useMemo(
    () =>
      pricingTiers
        .filter((t) => t.category === 'compute' && t.resourceType !== 'controlplane')
        .sort((a, b) => a.hourlyRate - b.hourlyRate),
    [pricingTiers],
  )

  const getUserCount = (resourceType: string) => userCounts[resourceType] ?? 0
  const setUserCount = (resourceType: string, count: number) => {
    setUserCounts((prev) => ({ ...prev, [resourceType]: Math.max(0, count) }))
  }

  const controlPlaneMonthlyCost = controlPlaneTier ? hourlyToMonthly(controlPlaneTier.hourlyRate) : 0
  const estimatedComputeMonthlyCost = computeTiers.reduce(
    (sum, tier) => sum + getUserCount(tier.resourceType) * tier.hourlyRate * workingHoursPerDay * 30,
    0,
  )
  const totalUsers = computeTiers.reduce((sum, tier) => sum + getUserCount(tier.resourceType), 0)
  const totalEstimatedMonthlyCost = controlPlaneMonthlyCost + estimatedComputeMonthlyCost

  const balance = creditsData?.account?.balance ?? 0
  const hasEnoughCredits = balance >= controlPlaneMonthlyCost
  const minimumTopup = Math.max(controlPlaneMonthlyCost - balance, 5)

  const isLoading = pricingLoading || creditsLoading

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="text-muted-foreground size-6 animate-spin" />
      </div>
    )
  }

  return (
    <div className="rounded-lg border border-foreground/10 bg-background">
      <div className="border-b border-foreground/10 px-6 py-4">
        <h3 className="font-medium text-foreground">Billing</h3>
      </div>
      <div className="px-6 py-4">
        {/* Control Plane */}
        <div className="space-y-2 text-sm">
          {controlPlaneTier && (
            <div className="flex items-center justify-between">
              <span className="flex items-center gap-2 text-muted-foreground">
                <Server className="size-3.5" />
                Control Plane (runs 24/7)
              </span>
              <span className="tabular-nums text-foreground">
                {formatDollars(controlPlaneTier.hourlyRate)}/hr{' '}
                <span className="text-muted-foreground">
                  (~{formatDollars(controlPlaneMonthlyCost)}/mo)
                </span>
              </span>
            </div>
          )}
        </div>

        {/* Expandable cost calculator */}
        <Collapsible open={calculatorOpen} onOpenChange={setCalculatorOpen} className="mt-4">
          <CollapsibleTrigger className="flex w-full items-center gap-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">
            <Calculator className="size-3.5" />
            Estimate full costs
            <ChevronDown className={cn('size-3.5 transition-transform', calculatorOpen && 'rotate-180')} />
          </CollapsibleTrigger>
          <CollapsibleContent className="mt-3 space-y-4">
            {computeTiers.length > 0 && (
              <div className="space-y-2">
                <Label className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  Users by tier
                </Label>
                <div className="grid gap-1.5">
                  {computeTiers.map((tier) => {
                    const count = getUserCount(tier.resourceType)
                    const tierCost = count * tier.hourlyRate * workingHoursPerDay * 30
                    return (
                      <div
                        key={tier.resourceType}
                        className={cn(
                          'flex items-center rounded-md border px-3 py-2 transition-colors',
                          count > 0 ? 'border-primary/30 bg-primary/5' : 'border-foreground/10',
                        )}
                      >
                        <div className="flex min-w-0 flex-1 flex-col">
                          <div className="flex items-baseline gap-1.5">
                            <span className="text-sm font-medium text-foreground">{tier.displayName}</span>
                            <span className="text-xs text-muted-foreground">{formatDollars(tier.hourlyRate)}/hr</span>
                          </div>
                          <span className="text-[11px] leading-tight text-muted-foreground/70">
                            {[
                              tier.specs?.vcpu && `${tier.specs.vcpu} vCPU`,
                              tier.specs?.memory_gb && `${tier.specs.memory_gb}GB RAM`,
                              tier.specs?.storage_gb && `${tier.specs.storage_gb}GB storage`,
                            ].filter(Boolean).join(' · ')}
                          </span>
                        </div>
                        <div className="flex items-center gap-1.5">
                          {count > 0 && (
                            <span className="text-xs tabular-nums text-muted-foreground">{formatDollars(tierCost)}</span>
                          )}
                          <div className="flex items-center rounded-md border border-foreground/10">
                            <Button type="button" variant="ghost" size="icon" className="size-7 rounded-r-none" onClick={() => setUserCount(tier.resourceType, count - 1)} disabled={count === 0}>
                              <Minus className="size-3" />
                            </Button>
                            <span className="w-7 border-x border-foreground/10 text-center text-xs font-semibold tabular-nums leading-7">{count}</span>
                            <Button type="button" variant="ghost" size="icon" className="size-7 rounded-l-none" onClick={() => setUserCount(tier.resourceType, count + 1)}>
                              <Plus className="size-3" />
                            </Button>
                          </div>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            )}

            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <Label className="text-xs font-medium uppercase tracking-wide text-muted-foreground">Avg. working hours / day</Label>
                <span className="text-sm font-semibold tabular-nums text-foreground">{workingHoursPerDay}h</span>
              </div>
              <Slider value={[workingHoursPerDay]} onValueChange={([v]) => setWorkingHoursPerDay(v)} min={1} max={24} step={1} className="w-full" />
              <div className="flex justify-between text-[11px] text-muted-foreground/60">
                <span>1h</span>
                <span>24h</span>
              </div>
            </div>

            <div className="space-y-1 rounded-md bg-muted/30 px-3 py-2.5">
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>Control Plane (24/7)</span>
                <span className="tabular-nums">{formatDollars(controlPlaneMonthlyCost)}/mo</span>
              </div>
              {computeTiers.map((tier) => {
                const count = getUserCount(tier.resourceType)
                if (count === 0) return null
                const cost = count * tier.hourlyRate * workingHoursPerDay * 30
                return (
                  <div key={tier.resourceType} className="flex items-center justify-between text-muted-foreground">
                    <span className="text-xs">{tier.displayName} ({count} {count === 1 ? 'user' : 'users'} x {workingHoursPerDay}h/day)</span>
                    <span className="text-xs tabular-nums">~{formatDollars(cost)}/mo</span>
                  </div>
                )
              })}
              <div className="flex items-center justify-between border-t border-foreground/10 pt-1.5 text-sm font-medium text-foreground">
                <span>Estimated total{totalUsers > 0 ? ` (${totalUsers} ${totalUsers === 1 ? 'user' : 'users'})` : ''}</span>
                <span className="tabular-nums">~{formatDollars(totalEstimatedMonthlyCost)}/mo</span>
              </div>
            </div>

            <p className="text-[11px] text-muted-foreground/70">
              This is an estimate. You are billed only for actual usage. Storage is billed separately when provisioned.
            </p>
          </CollapsibleContent>
        </Collapsible>

        {!calculatorOpen && (
          <div className="mt-3 flex items-center justify-between border-t border-foreground/10 pt-3">
            <span className="text-sm font-medium text-foreground">Control plane (30 days)</span>
            <span className="text-lg font-bold tabular-nums text-foreground">
              ~{formatDollars(controlPlaneMonthlyCost)}/mo
            </span>
          </div>
        )}
      </div>

      {/* Balance Gate + Submit */}
      <div className="border-t border-foreground/10 px-6 py-4">
        <div className="mb-4 flex items-center gap-2 text-sm">
          <Wallet className="size-4 text-muted-foreground" />
          <span className="text-muted-foreground">Current balance:</span>
          <span className={cn('font-semibold tabular-nums', hasEnoughCredits ? 'text-green-600 dark:text-green-500' : 'text-destructive')}>
            {formatDollars(balance)}
          </span>
        </div>

        {hasEnoughCredits ? (
          <Button
            type="button"
            className="w-full"
            size="lg"
            disabled={creating || (!isSubscribeOnly && subdomainAvailable !== true)}
            onClick={onSubmit}
          >
            {creating ? (
              <>
                <Loader2 className="mr-2 size-4 animate-spin" />
                Creating...
              </>
            ) : (
              'Create Installation'
            )}
          </Button>
        ) : (
          <div className="space-y-3">
            <div className="rounded-md border border-yellow-500/30 bg-yellow-500/5 px-4 py-3">
              <p className="text-sm font-medium text-foreground">Insufficient credits</p>
              <p className="mt-1 text-xs text-muted-foreground">
                You need at least {formatDollars(controlPlaneMonthlyCost)} to cover 30 days of control plane. Current balance: {formatDollars(balance)}.
              </p>
            </div>
            <Button type="button" className="w-full" size="lg" onClick={() => {
              const values = getFormValues?.()
              const params = new URLSearchParams()
              if (values?.name) params.set('name', values.name)
              if (values?.subdomain) params.set('subdomain', values.subdomain)
              const returnUrl = `/installations/new?${params.toString()}`
              handleTopup(minimumTopup, returnUrl)
            }}>
              Add Credits ({formatDollars(minimumTopup)} minimum)
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
