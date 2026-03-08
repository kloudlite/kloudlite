'use client'

import { useState, useEffect, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import {
  Button,
  Input,
  Slider,
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
  Label,
  RadioGroup,
  RadioGroupItem,
} from '@kloudlite/ui'
import { cn } from '@kloudlite/lib'
import { Loader2, CheckCircle2, AlertCircle, Cpu, HardDrive, Wallet, Server } from 'lucide-react'
import { toast } from 'sonner'
import { getErrorMessage } from '@/lib/errors'
import { useSubdomainCheck } from '@/hooks/use-subdomain-check'
import { useCredits } from '@/hooks/use-credits'
import type { PricingTier } from '@/lib/console/storage/credits-types'

// --- Pure helpers (no server imports) ---

function calculateProjectedMonthlyCost(
  tiers: PricingTier[],
  selectedResources: Array<{
    resourceType: string
    quantity?: number
    sizeGb?: number
  }>,
): number {
  let total = 0
  for (const resource of selectedResources) {
    const tier = tiers.find((t) => t.resourceType === resource.resourceType)
    if (!tier) continue
    if (tier.category === 'storage') {
      total += tier.hourlyRate * (resource.sizeGb ?? 0) * 24 * 30
    } else {
      total += tier.hourlyRate * (resource.quantity ?? 1) * 24 * 30
    }
  }
  return total
}

function calculateMinimumTopup(projectedMonthlyCost: number, currentBalance: number): number {
  const needed = projectedMonthlyCost - currentBalance
  return Math.max(needed, 5)
}

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

  // Derived tier lists
  const controlPlaneTier = useMemo(
    () => pricingTiers.find((t) => t.category === 'compute' && t.resourceType.includes('controlplane')),
    [pricingTiers],
  )
  const computeTiers = useMemo(
    () =>
      pricingTiers.filter(
        (t) => t.category === 'compute' && !t.resourceType.includes('controlplane'),
      ),
    [pricingTiers],
  )
  const storageTier = useMemo(
    () => pricingTiers.find((t) => t.category === 'storage'),
    [pricingTiers],
  )

  // Form state for new fields
  const [selectedComputeType, setSelectedComputeType] = useState<string>('')
  const [storageSize, setStorageSize] = useState(50)

  // Set default compute selection when tiers load
  useEffect(() => {
    if (computeTiers.length > 0 && !selectedComputeType) {
      setSelectedComputeType(computeTiers[0].resourceType)
    }
  }, [computeTiers, selectedComputeType])

  const selectedComputeTier = useMemo(
    () => computeTiers.find((t) => t.resourceType === selectedComputeType),
    [computeTiers, selectedComputeType],
  )

  // Cost calculation
  const projectedMonthlyCost = useMemo(() => {
    const resources: Array<{ resourceType: string; quantity?: number; sizeGb?: number }> = []
    if (controlPlaneTier) {
      resources.push({ resourceType: controlPlaneTier.resourceType, quantity: 1 })
    }
    if (selectedComputeTier) {
      resources.push({ resourceType: selectedComputeTier.resourceType, quantity: 1 })
    }
    if (storageTier) {
      resources.push({ resourceType: storageTier.resourceType, sizeGb: storageSize })
    }
    return calculateProjectedMonthlyCost(pricingTiers, resources)
  }, [pricingTiers, controlPlaneTier, selectedComputeTier, storageTier, storageSize])

  const balance = creditsData?.account?.balance ?? 0
  const hasEnoughCredits = balance >= projectedMonthlyCost
  const minimumTopup = calculateMinimumTopup(projectedMonthlyCost, balance)

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

        {/* Section 2: WorkMachine Configuration */}
        <div className="rounded-lg border border-foreground/10 bg-background">
          <div className="border-b border-foreground/10 px-6 py-4">
            <h3 className="font-medium text-foreground">WorkMachine Configuration</h3>
            <p className="mt-0.5 text-xs text-muted-foreground">
              Select a compute size for your WorkMachines
            </p>
          </div>
          <div className="p-6">
            {computeTiers.length === 0 ? (
              <p className="text-sm text-muted-foreground">No compute tiers available</p>
            ) : (
              <RadioGroup
                value={selectedComputeType}
                onValueChange={setSelectedComputeType}
                className="space-y-3"
              >
                {computeTiers.map((tier) => {
                  const specs = tier.specs as { vcpu?: number; ram_gb?: number }
                  const isSelected = selectedComputeType === tier.resourceType
                  return (
                    <label
                      key={tier.id}
                      className={cn(
                        'flex cursor-pointer items-center gap-4 rounded-lg border px-4 py-4 transition-colors',
                        isSelected
                          ? 'border-primary/40 bg-primary/[0.03]'
                          : 'border-foreground/10 bg-background hover:border-foreground/20',
                      )}
                    >
                      <RadioGroupItem value={tier.resourceType} />
                      <div
                        className={cn(
                          'flex size-9 shrink-0 items-center justify-center rounded-lg',
                          isSelected
                            ? 'bg-primary/10 text-primary'
                            : 'bg-foreground/[0.06] text-muted-foreground',
                        )}
                      >
                        <Cpu className="size-4" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-baseline gap-2">
                          <span className="text-sm font-semibold text-foreground">
                            {tier.displayName}
                          </span>
                          {specs.vcpu && specs.ram_gb && (
                            <span className="text-xs text-muted-foreground">
                              {specs.vcpu} vCPU / {specs.ram_gb} GB RAM
                            </span>
                          )}
                        </div>
                      </div>
                      <span className="shrink-0 text-sm font-semibold tabular-nums text-foreground">
                        {formatDollars(tier.hourlyRate)}/hr
                      </span>
                    </label>
                  )
                })}
              </RadioGroup>
            )}
          </div>
        </div>

        {/* Section 3: Storage Size */}
        {storageTier && (
          <div className="rounded-lg border border-foreground/10 bg-background">
            <div className="border-b border-foreground/10 px-6 py-4">
              <h3 className="font-medium text-foreground">Storage</h3>
              <p className="mt-0.5 text-xs text-muted-foreground">
                Configure initial volume size
              </p>
            </div>
            <div className="space-y-4 p-6">
              <div className="flex items-center justify-between">
                <Label className="text-sm font-medium text-foreground">Volume Size</Label>
                <span className="text-sm tabular-nums text-muted-foreground">
                  {formatDollars(storageTier.hourlyRate * 24 * 30)}/GB/mo
                </span>
              </div>
              <div className="flex items-center gap-4">
                <Slider
                  value={[storageSize]}
                  onValueChange={([val]) => setStorageSize(val)}
                  min={50}
                  max={1000}
                  step={10}
                  className="flex-1"
                  disabled={creating}
                />
                <div className="flex items-center gap-1.5">
                  <HardDrive className="size-3.5 text-muted-foreground" />
                  <span className="w-20 text-right text-sm font-semibold tabular-nums text-foreground">
                    {storageSize} GB
                  </span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Section 4: Cost Summary */}
        <div className="rounded-lg border border-foreground/10 bg-background">
          <div className="border-b border-foreground/10 px-6 py-4">
            <h3 className="font-medium text-foreground">Estimated Cost</h3>
          </div>
          <div className="px-6 py-4">
            <div className="space-y-2 text-sm">
              {/* Control Plane */}
              {controlPlaneTier && (
                <div className="flex items-center justify-between">
                  <span className="flex items-center gap-2 text-muted-foreground">
                    <Server className="size-3.5" />
                    Control Plane
                  </span>
                  <span className="tabular-nums text-foreground">
                    {formatDollars(controlPlaneTier.hourlyRate)}/hr{' '}
                    <span className="text-muted-foreground">
                      (~{formatDollars(hourlyToMonthly(controlPlaneTier.hourlyRate))}/mo)
                    </span>
                  </span>
                </div>
              )}

              {/* Selected WorkMachine */}
              {selectedComputeTier && (
                <div className="flex items-center justify-between">
                  <span className="flex items-center gap-2 text-muted-foreground">
                    <Cpu className="size-3.5" />
                    1&times; {selectedComputeTier.displayName} WorkMachine
                  </span>
                  <span className="tabular-nums text-foreground">
                    {formatDollars(selectedComputeTier.hourlyRate)}/hr{' '}
                    <span className="text-muted-foreground">
                      (~{formatDollars(hourlyToMonthly(selectedComputeTier.hourlyRate))}/mo)
                    </span>
                  </span>
                </div>
              )}

              {/* Storage */}
              {storageTier && (
                <div className="flex items-center justify-between">
                  <span className="flex items-center gap-2 text-muted-foreground">
                    <HardDrive className="size-3.5" />
                    Storage ({storageSize} GB)
                  </span>
                  <span className="tabular-nums text-foreground">
                    ~{formatDollars(storageTier.hourlyRate * storageSize * 24 * 30)}/mo
                  </span>
                </div>
              )}
            </div>

            {/* Total */}
            <div className="mt-3 flex items-center justify-between border-t border-foreground/10 pt-3">
              <span className="text-sm font-medium text-foreground">Estimated total</span>
              <span className="text-lg font-bold tabular-nums text-foreground">
                ~{formatDollars(projectedMonthlyCost)}/mo
              </span>
            </div>

            <p className="mt-2 text-xs text-muted-foreground">
              Billed by actual usage. WorkMachines only charged while running.
            </p>
          </div>

          {/* Section 5: Balance Gate + Submit */}
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
                    You need at least {formatDollars(projectedMonthlyCost)} to cover 30 days of
                    estimated usage. Current balance: {formatDollars(balance)}.
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
