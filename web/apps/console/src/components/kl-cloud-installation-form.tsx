'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import {
  Button,
  Input,
  Textarea,
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from '@kloudlite/ui'
import { Loader2, CheckCircle2, AlertCircle, Minus, Plus, Cpu, HardDrive, Clock, Zap } from 'lucide-react'
import { toast } from 'sonner'
import {
  getRazorpayKey,
  createInstallationOrder,
  verifyPaymentAndActivate,
} from '@/app/actions/billing'
import { useRazorpay } from '@/components/razorpay-provider'
import type { Plan } from '@/lib/console/storage'

const installationSchema = z.object({
  name: z
    .string()
    .min(3, 'Name must be at least 3 characters')
    .max(50, 'Name must be less than 50 characters')
    .regex(/^[a-zA-Z0-9\s-]+$/, 'Name can only contain letters, numbers, spaces, and hyphens'),
  description: z.string().max(200, 'Description must be less than 200 characters').optional(),
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

interface KlCloudInstallationFormProps {
  plans: Plan[]
  existingInstallationId?: string
}

export function KlCloudInstallationForm({
  plans,
  existingInstallationId,
}: KlCloudInstallationFormProps) {
  const isSubscribeOnly = !!existingInstallationId
  const router = useRouter()
  const { isLoaded: razorpayLoaded, openCheckout } = useRazorpay()
  const [creating, setCreating] = useState(false)
  const [checkingSubdomain, setCheckingSubdomain] = useState(false)
  const [subdomainAvailable, setSubdomainAvailable] = useState<boolean | null>(null)
  const [razorpayKey, setRazorpayKey] = useState<string | null>(null)
  const [isLoadingKey, setIsLoadingKey] = useState(false)

  const loadRazorpayKey = async () => {
    if (razorpayKey || isLoadingKey) return
    setIsLoadingKey(true)
    try {
      const key = await getRazorpayKey()
      setRazorpayKey(key)
    } catch {
      toast.error('Failed to load payment configuration')
    } finally {
      setIsLoadingKey(false)
    }
  }

  // Per-tier quantities
  const [quantities, setQuantities] = useState<Record<string, number>>(() => {
    const initial: Record<string, number> = {}
    for (const plan of plans) {
      initial[plan.id] = 0
    }
    return initial
  })

  const baseFee = plans[0]?.baseFee ? plans[0].baseFee / 100 : 29
  const currencySymbol = plans[0]?.currency === 'INR' ? '₹' : '$'
  const totalUsers = Object.values(quantities).reduce((sum, q) => sum + q, 0)

  // Calculate cost breakdown per tier
  const tierCosts = plans
    .filter((plan) => (quantities[plan.id] || 0) > 0)
    .map((plan) => {
      const qty = quantities[plan.id] || 0
      return {
        plan,
        quantity: qty,
        lineTotal: (plan.amountPerUser * qty) / 100,
      }
    })

  const userTotal = tierCosts.reduce((sum, t) => sum + t.lineTotal, 0)
  const monthlyTotal = baseFee + userTotal

  const form = useForm<InstallationFormData>({
    resolver: zodResolver(installationSchema),
    defaultValues: {
      name: '',
      description: '',
      subdomain: '',
    },
  })

  const setQuantity = (planId: string, value: number) => {
    setQuantities((prev) => ({ ...prev, [planId]: Math.max(0, Math.min(100, value)) }))
  }

  const checkSubdomainAvailability = async (subdomain: string) => {
    if (!subdomain || subdomain.length < 3) {
      setSubdomainAvailable(null)
      return
    }

    const subdomainRegex = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/
    if (!subdomainRegex.test(subdomain)) {
      setSubdomainAvailable(null)
      return
    }

    setCheckingSubdomain(true)
    try {
      const response = await fetch(`/api/installations/check-domain-kli?subdomain=${subdomain}`)
      const data = await response.json()
      setSubdomainAvailable(data.available)
    } catch (err) {
      console.error('Error checking subdomain:', err)
      setSubdomainAvailable(false)
    } finally {
      setCheckingSubdomain(false)
    }
  }

  const onSubmit = async (data: InstallationFormData) => {
    if (!isSubscribeOnly && subdomainAvailable !== true) {
      toast.error('Please choose an available subdomain')
      return
    }
    if (totalUsers === 0) {
      toast.error('Please add at least one user')
      return
    }

    setCreating(true)

    try {
      let installationId: string

      if (isSubscribeOnly) {
        // Existing installation — skip creation
        installationId = existingInstallationId!
      } else {
        // Step 1: Create the installation
        const response = await fetch('/api/installations/create-installation', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            name: data.name,
            description: data.description || undefined,
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

      // Step 2: Build tier allocations (only tiers with users > 0)
      const tierAllocations = plans
        .filter((plan) => (quantities[plan.id] || 0) > 0)
        .map((plan) => ({
          planId: plan.id,
          quantity: quantities[plan.id],
        }))

      // Step 3: Create Razorpay order for the total amount
      const order = await createInstallationOrder(installationId, tierAllocations)

      // Step 4: Load Razorpay key and open checkout
      if (!razorpayKey) {
        await loadRazorpayKey()
        return
      }

      // Step 5: Open Razorpay Checkout for payment
      const options = {
        key: razorpayKey,
        order_id: order.razorpayOrderId,
        amount: order.amount,
        currency: order.currency,
        name: 'Kloudlite',
        description: `${totalUsers} ${totalUsers === 1 ? 'user' : 'users'} — Kloudlite Cloud`,
        theme: {
          color: '#3B82F6',
        },
        handler: async (response: Record<string, string>) => {
          try {
            // Verify payment server-side and activate subscriptions
            await verifyPaymentAndActivate(
              installationId,
              response.razorpay_order_id,
              response.razorpay_payment_id,
              response.razorpay_signature,
            )
            toast.success('Payment successful! Starting deployment...')
            router.push('/installations/new/kloudlite-cloud')
          } catch {
            toast.error('Payment verification failed. Please contact support.')
          }
        },
        modal: {
          ondismiss: () => {
            toast.info(
              'Payment cancelled. Your installation has been created — you can subscribe from billing settings.',
            )
            router.push('/installations/new/kloudlite-cloud')
          },
        },
      }

      openCheckout(options)
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to create installation')
      toast.error(error.message)
    } finally {
      setCreating(false)
    }
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
          {/* Section 1: Installation Details (hidden when subscribing to existing) */}
          {!isSubscribeOnly && (
          <div className="border border-foreground/10 rounded-lg bg-background">
            <div className="px-6 py-4 border-b border-foreground/10">
              <h3 className="font-medium text-foreground">Installation Details</h3>
            </div>
            <div className="p-6 space-y-5">
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
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      Description{' '}
                      <span className="text-muted-foreground font-normal">(optional)</span>
                    </FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="Production deployment for our platform"
                        {...field}
                        disabled={creating}
                        rows={3}
                        className="resize-none"
                      />
                    </FormControl>
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
                      <span className="text-xs font-medium whitespace-nowrap">
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

          {/* Section 2: Compute Sizes & Users */}
          <div className="border border-foreground/10 rounded-lg bg-background">
            <div className="px-6 py-4 border-b border-foreground/10">
              <h3 className="font-medium text-foreground">Compute & Users</h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                Select a compute size and number of users for each
              </p>
            </div>
            <div className="p-6 space-y-4">
              {/* Base Fee — compact inline */}
              <div className="flex items-center justify-between rounded-md bg-muted/40 px-4 py-3">
                <div className="flex items-center gap-2">
                  <div className="flex size-7 items-center justify-center rounded-md bg-primary/10">
                    <Zap className="size-3.5 text-primary" />
                  </div>
                  <div>
                    <span className="text-sm font-medium text-foreground">Control Plane</span>
                    <span className="text-xs text-muted-foreground ml-2">
                      Dashboard, user management, billing
                    </span>
                  </div>
                </div>
                <span className="text-sm font-semibold text-foreground tabular-nums">
                  {currencySymbol}{baseFee}/mo
                </span>
              </div>

              {/* Compute Size Cards */}
              <div className="space-y-3">
                {plans.map((plan) => {
                  const qty = quantities[plan.id] || 0
                  const isActive = qty > 0
                  return (
                    <div
                      key={plan.id}
                      className={`rounded-lg border transition-colors ${
                        isActive
                          ? 'border-primary/40 bg-primary/[0.03]'
                          : 'border-foreground/10 bg-background'
                      }`}
                    >
                      <div className="px-4 py-4">
                        {/* Top row: Name + Price + Stepper */}
                        <div className="flex items-center justify-between gap-4">
                          <div className="flex items-center gap-3 min-w-0">
                            <div
                              className={`flex size-9 items-center justify-center rounded-lg shrink-0 ${
                                isActive
                                  ? 'bg-primary/10 text-primary'
                                  : 'bg-foreground/[0.06] text-muted-foreground'
                              }`}
                            >
                              <Cpu className="size-4" />
                            </div>
                            <div className="min-w-0">
                              <div className="flex items-baseline gap-2">
                                <h4 className="text-sm font-semibold text-foreground">
                                  {plan.name}
                                </h4>
                                <span className="text-xs text-muted-foreground">
                                  {currencySymbol}{plan.amountPerUser / 100}/user/mo
                                </span>
                              </div>
                              <p className="text-xs text-muted-foreground mt-0.5">
                                {plan.cpu} vCPU &middot; {plan.ram} RAM &middot; {plan.storage}
                              </p>
                            </div>
                          </div>

                          {/* Quantity stepper */}
                          <div className="flex items-center gap-0 shrink-0">
                            <button
                              type="button"
                              className="flex size-8 items-center justify-center rounded-l-md border border-foreground/10 bg-background text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:opacity-40 disabled:pointer-events-none"
                              disabled={qty <= 0 || creating}
                              onClick={() => setQuantity(plan.id, qty - 1)}
                            >
                              <Minus className="size-3" />
                            </button>
                            <input
                              type="number"
                              min={0}
                              max={100}
                              value={qty}
                              onChange={(e) =>
                                setQuantity(plan.id, parseInt(e.target.value) || 0)
                              }
                              className="h-8 w-12 border-y border-foreground/10 bg-background text-center font-mono text-sm text-foreground outline-none [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                              disabled={creating}
                            />
                            <button
                              type="button"
                              className="flex size-8 items-center justify-center rounded-r-md border border-foreground/10 bg-background text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:opacity-40 disabled:pointer-events-none"
                              disabled={qty >= 100 || creating}
                              onClick={() => setQuantity(plan.id, qty + 1)}
                            >
                              <Plus className="size-3" />
                            </button>
                          </div>
                        </div>

                        {/* Spec chips row */}
                        <div className="flex flex-wrap gap-x-1.5 gap-y-1 mt-3 ml-12">
                          <span className="inline-flex items-center gap-1 rounded-full bg-foreground/[0.05] px-2 py-0.5 text-[11px] text-muted-foreground">
                            <Clock className="size-2.5" />
                            {plan.monthlyHours} hrs/mo
                          </span>
                          <span className="inline-flex items-center gap-1 rounded-full bg-foreground/[0.05] px-2 py-0.5 text-[11px] text-muted-foreground">
                            <HardDrive className="size-2.5" />
                            {plan.storage}
                          </span>
                          <span className="inline-flex items-center gap-1 rounded-full bg-foreground/[0.05] px-2 py-0.5 text-[11px] text-muted-foreground">
                            {plan.autoSuspend} suspend
                          </span>
                          <span className="inline-flex items-center gap-1 rounded-full bg-foreground/[0.05] px-2 py-0.5 text-[11px] text-muted-foreground">
                            +{currencySymbol}{(plan.overageRate / 100).toFixed(2)}/hr overage
                          </span>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          </div>

          {/* Section 3: Cost Summary & Submit */}
          <div className="border border-foreground/10 rounded-lg bg-background">
            <div className="px-6 py-4 border-b border-foreground/10">
              <h3 className="font-medium text-foreground">Summary</h3>
            </div>
            <div className="px-6 py-4">
              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Control Plane</span>
                  <span className="text-foreground tabular-nums">{currencySymbol}{baseFee.toFixed(2)}</span>
                </div>
                {tierCosts.map(({ plan, quantity: qty, lineTotal }) => (
                  <div key={plan.id} className="flex items-center justify-between">
                    <span className="text-muted-foreground">
                      {plan.name} ({plan.cpu} vCPU) &times; {qty}{' '}
                      {qty === 1 ? 'user' : 'users'}
                    </span>
                    <span className="text-foreground tabular-nums">{currencySymbol}{lineTotal.toFixed(2)}</span>
                  </div>
                ))}
                {totalUsers === 0 && (
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground italic">No users added yet</span>
                    <span className="text-foreground tabular-nums">{currencySymbol}0.00</span>
                  </div>
                )}
              </div>

              <div className="border-t border-foreground/10 mt-3 pt-3 flex items-center justify-between">
                <span className="text-sm font-medium text-foreground">
                  Monthly total ({totalUsers} {totalUsers === 1 ? 'user' : 'users'})
                </span>
                <span className="text-lg font-bold text-foreground tabular-nums">
                  {currencySymbol}{monthlyTotal.toFixed(2)}
                </span>
              </div>
            </div>

            <div className="border-t border-foreground/10 px-6 py-4">
              <Button
                type="submit"
                className="w-full"
                size="lg"
                disabled={
                  creating ||
                  !razorpayLoaded ||
                  (!isSubscribeOnly && subdomainAvailable !== true) ||
                  totalUsers === 0
                }
              >
                {creating ? (
                  <>
                    <Loader2 className="mr-2 size-4 animate-spin" />
                    {isSubscribeOnly ? 'Subscribing...' : 'Creating...'}
                  </>
                ) : isSubscribeOnly ? (
                  `Subscribe — ${currencySymbol}${monthlyTotal.toFixed(2)}/mo`
                ) : (
                  `Create & Subscribe — ${currencySymbol}${monthlyTotal.toFixed(2)}/mo`
                )}
              </Button>
            </div>
          </div>
        </form>
      </Form>
  )
}
