'use client'

import { useState, useEffect, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import {
  Button,
  Input,
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
  Collapsible,
  CollapsibleTrigger,
  CollapsibleContent,
  Label,
  Slider,
} from '@kloudlite/ui'
import { cn } from '@kloudlite/lib'
import { Loader2, CheckCircle2, AlertCircle, Wallet, Server, Calculator, ChevronDown, Minus, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { getErrorMessage } from '@/lib/errors'
import { useSubdomainCheck } from '@/hooks/use-subdomain-check'
import { useCredits } from '@/hooks/use-credits'
import type { PricingTier } from '@/lib/console/storage/credits-types'

// --- Pure helpers (no server imports) ---

function formatDollars(amount: number): string {
  return `$${amount.toFixed(2)}`
}

function hourlyToMonthly(hourly: number): number {
  return hourly * 24 * 30
}

// --- Schema ---

const installationSchema = z.object({
  name: z
    .string()
    .min(3, 'Name must be at least 3 characters')
    .max(50, 'Name must be less than 50 characters')
    .regex(/^[a-zA-Z0-9\s-]+$/, 'Name can only contain letters, numbers, spaces, and hyphens'),
  subdomain: z
    .string()
    .min(3, 'Subdomain must be at least 3 characters')
    .max(63, 'Subdomain must be less than 63 characters')
    .regex(
      /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/,
      'Subdomain must start and end with alphanumeric characters and can only contain lowercase letters, numbers, and hyphens',
    ),
})

type InstallationFormData = z.infer<typeof installationSchema>

// --- Component ---

interface KlCloudInstallationFormProps {
  orgId: string
  existingInstallationId?: string
}

export function KlCloudInstallationForm({
  orgId,
  existingInstallationId,
}: KlCloudInstallationFormProps) {
  const isSubscribeOnly = !!existingInstallationId
  const [creating, setCreating] = useState(false)
  const [pricingTiers, setPricingTiers] = useState<PricingTier[]>([])
  const [pricingLoading, setPricingLoading] = useState(true)

  const {
    checking: checkingSubdomain,
    available: subdomainAvailable,
    check: checkSubdomainAvailability,
  } = useSubdomainCheck({ endpoint: '/api/installations/check-domain-kli' })

  const { loading: creditsLoading, data: creditsData, handleTopup } = useCredits(orgId)

  // Fetch pricing tiers
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

  const [calculatorOpen, setCalculatorOpen] = useState(false)
  const [userCounts, setUserCounts] = useState<Record<string, number>>({})
  const [workingHoursPerDay, setWorkingHoursPerDay] = useState(8)

  // Pricing tiers by type
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

  // Control plane runs 24/7, so 30-day cost is the gate
  const controlPlaneMonthlyCost = controlPlaneTier ? hourlyToMonthly(controlPlaneTier.hourlyRate) : 0

  // Estimated WorkMachine cost: sum of (users × hourlyRate × hours/day × 30) per tier
  const estimatedComputeMonthlyCost = computeTiers.reduce(
    (sum, tier) => sum + getUserCount(tier.resourceType) * tier.hourlyRate * workingHoursPerDay * 30,
    0,
  )

  const totalUsers = computeTiers.reduce((sum, tier) => sum + getUserCount(tier.resourceType), 0)

  // Total estimate (for display only — not used for balance gate)
  const totalEstimatedMonthlyCost = controlPlaneMonthlyCost + estimatedComputeMonthlyCost

  // Balance gate: only require 30 days of control plane
  const balance = creditsData?.account?.balance ?? 0
  const hasEnoughCredits = balance >= controlPlaneMonthlyCost
  const minimumTopup = Math.max(controlPlaneMonthlyCost - balance, 5)

  const form = useForm<InstallationFormData>({
    resolver: zodResolver(installationSchema),
    defaultValues: {
      name: '',
      subdomain: '',
    },
  })

  const onSubmit = async (data: InstallationFormData) => {
    if (!isSubscribeOnly && subdomainAvailable !== true) {
      toast.error('Please choose an available subdomain')
      return
    }

    setCreating(true)

    try {
      let installationId: string

      if (isSubscribeOnly) {
        installationId = existingInstallationId!
      } else {
        const response = await fetch('/api/installations/create-installation', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            orgId,
            name: data.name,
            subdomain: data.subdomain,
            hostingType: 'kloudlite',
          }),
        })

        if (!response.ok) {
          const errorData = await response.json()
          throw new Error(errorData.error || 'Failed to create installation')
        }

        const result = await response.json()
        installationId = result.installationId
      }

      // Redirect to deploy page — no subscription needed
      window.location.href = `/api/installations/${installationId}/continue`
    } catch (err) {
      toast.error(getErrorMessage(err, 'Failed to create installation'))
      setCreating(false)
    }
  }

  const isLoading = pricingLoading || creditsLoading

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="text-muted-foreground size-6 animate-spin" />
      </div>
    )
  }

  return (
    <Form {...form}>
      <form
        onSubmit={
          isSubscribeOnly
            ? (e) => {
                e.preventDefault()
                onSubmit({ name: '', subdomain: '' })
              }
            : form.handleSubmit(onSubmit)
        }
        className="space-y-8"
      >
        {/* Section 1: Installation Details */}
        {!isSubscribeOnly && (
          <div className="rounded-lg border border-foreground/10 bg-background">
            <div className="border-b border-foreground/10 px-6 py-4">
              <h3 className="font-medium text-foreground">Installation Details</h3>
            </div>
            <div className="space-y-5 p-6">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input placeholder="e.g., Production" {...field} disabled={creating} />
                    </FormControl>
                    <FormDescription>A friendly name for this installation</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="subdomain"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Domain</FormLabel>
                    <FormControl>
                      <div className="relative">
                        <Input
                          placeholder="your-company"
                          {...field}
                          disabled={creating}
                          className="font-mono"
                          onChange={(e) => {
                            const value = e.target.value.toLowerCase()
                            field.onChange(value)
                            checkSubdomainAvailability(value)
                          }}
                        />
                        {checkingSubdomain && (
                          <div className="absolute top-1/2 right-3 -translate-y-1/2">
                            <Loader2 className="text-muted-foreground size-4 animate-spin" />
                          </div>
                        )}
                        {!checkingSubdomain && subdomainAvailable === true && (
                          <div className="absolute top-1/2 right-3 -translate-y-1/2">
                            <CheckCircle2 className="size-4 text-green-600" />
                          </div>
                        )}
                        {!checkingSubdomain && subdomainAvailable === false && (
                          <div className="absolute top-1/2 right-3 -translate-y-1/2">
                            <AlertCircle className="text-destructive size-4" />
                          </div>
                        )}
                      </div>
                    </FormControl>
                    <FormDescription className="flex items-center justify-between gap-3">
                      <span className="font-mono">
                        {field.value || 'your-subdomain'}.
                        {process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                      </span>
                      <span className="whitespace-nowrap text-xs font-medium">
                        {!checkingSubdomain && subdomainAvailable === false && (
                          <span className="text-destructive">This domain is already taken</span>
                        )}
                        {!checkingSubdomain && subdomainAvailable === true && (
                          <span className="text-green-600 dark:text-green-500">
                            Domain is available
                          </span>
                        )}
                      </span>
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>
        )}

        {/* Section 2: Billing */}
        <div className="rounded-lg border border-foreground/10 bg-background">
          <div className="border-b border-foreground/10 px-6 py-4">
            <h3 className="font-medium text-foreground">Billing</h3>
          </div>
          <div className="px-6 py-4">
            {/* Control Plane — always shown */}
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
                <ChevronDown
                  className={cn(
                    'size-3.5 transition-transform',
                    calculatorOpen && 'rotate-180',
                  )}
                />
              </CollapsibleTrigger>
              <CollapsibleContent className="mt-3 space-y-4">
                {/* Users per tier */}
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
                              count > 0
                                ? 'border-primary/30 bg-primary/5'
                                : 'border-foreground/10',
                            )}
                          >
                            <div className="flex min-w-0 flex-1 flex-col">
                              <div className="flex items-baseline gap-1.5">
                                <span className="text-sm font-medium text-foreground">
                                  {tier.displayName}
                                </span>
                                <span className="text-xs text-muted-foreground">
                                  {formatDollars(tier.hourlyRate)}/hr
                                </span>
                              </div>
                              <span className="text-[11px] leading-tight text-muted-foreground/70">
                                {[
                                  tier.specs?.vcpu && `${tier.specs.vcpu} vCPU`,
                                  tier.specs?.memory_gb && `${tier.specs.memory_gb}GB RAM`,
                                  tier.specs?.storage_gb && `${tier.specs.storage_gb}GB storage`,
                                ]
                                  .filter(Boolean)
                                  .join(' · ')}
                              </span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              {count > 0 && (
                                <span className="text-xs tabular-nums text-muted-foreground">
                                  {formatDollars(tierCost)}
                                </span>
                              )}
                              <div className="flex items-center rounded-md border border-foreground/10">
                                <Button
                                  type="button"
                                  variant="ghost"
                                  size="icon"
                                  className="size-7 rounded-r-none"
                                  onClick={() => setUserCount(tier.resourceType, count - 1)}
                                  disabled={count === 0}
                                >
                                  <Minus className="size-3" />
                                </Button>
                                <span className="w-7 border-x border-foreground/10 text-center text-xs font-semibold tabular-nums leading-7">
                                  {count}
                                </span>
                                <Button
                                  type="button"
                                  variant="ghost"
                                  size="icon"
                                  className="size-7 rounded-l-none"
                                  onClick={() => setUserCount(tier.resourceType, count + 1)}
                                >
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

                {/* Working hours per day */}
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between">
                    <Label className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                      Avg. working hours / day
                    </Label>
                    <span className="text-sm font-semibold tabular-nums text-foreground">
                      {workingHoursPerDay}h
                    </span>
                  </div>
                  <Slider
                    value={[workingHoursPerDay]}
                    onValueChange={([v]) => setWorkingHoursPerDay(v)}
                    min={1}
                    max={24}
                    step={1}
                    className="w-full"
                  />
                  <div className="flex justify-between text-[11px] text-muted-foreground/60">
                    <span>1h</span>
                    <span>24h</span>
                  </div>
                </div>

                {/* Estimated breakdown */}
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
                        <span className="text-xs">
                          {tier.displayName} ({count} {count === 1 ? 'user' : 'users'} x {workingHoursPerDay}h/day)
                        </span>
                        <span className="text-xs tabular-nums">
                          ~{formatDollars(cost)}/mo
                        </span>
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

            {/* Minimum cost line (when calculator is closed) */}
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
              <span
                className={cn(
                  'font-semibold tabular-nums',
                  hasEnoughCredits ? 'text-green-600 dark:text-green-500' : 'text-destructive',
                )}
              >
                {formatDollars(balance)}
              </span>
            </div>

            {hasEnoughCredits ? (
              <Button
                type="submit"
                className="w-full"
                size="lg"
                disabled={creating || (!isSubscribeOnly && subdomainAvailable !== true)}
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
                    You need at least {formatDollars(controlPlaneMonthlyCost)} to cover 30 days of
                    control plane. Current balance: {formatDollars(balance)}.
                  </p>
                </div>
                <Button
                  type="button"
                  className="w-full"
                  size="lg"
                  onClick={() => handleTopup(minimumTopup, existingInstallationId)}
                >
                  Add Credits ({formatDollars(minimumTopup)} minimum)
                </Button>
              </div>
            )}
          </div>
        </div>
      </form>
    </Form>
  )
}
