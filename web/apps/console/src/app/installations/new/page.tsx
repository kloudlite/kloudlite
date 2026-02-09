'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button, Input, Textarea, Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription, RadioGroup, RadioGroupItem, Label } from '@kloudlite/ui'
import { Loader2, CheckCircle2, AlertCircle } from 'lucide-react'
import { toast } from 'sonner'

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
  hostingType: z.enum(['kloudlite', 'byoc']),
})

type InstallationFormData = z.infer<typeof installationSchema>

export default function NewInstallationPage() {
  const router = useRouter()
  const [creating, setCreating] = useState(false)
  const [checkingSubdomain, setCheckingSubdomain] = useState(false)
  const [subdomainAvailable, setSubdomainAvailable] = useState<boolean | null>(null)

  const form = useForm<InstallationFormData>({
    resolver: zodResolver(installationSchema),
    defaultValues: {
      name: '',
      description: '',
      subdomain: '',
      hostingType: 'kloudlite',
    },
  })

  const checkSubdomainAvailability = async (subdomain: string) => {
    if (!subdomain || subdomain.length < 3) {
      setSubdomainAvailable(null)
      return
    }

    // Check if subdomain format is valid before making API call
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
    if (subdomainAvailable !== true) {
      toast.error('Please choose an available subdomain')
      return
    }

    setCreating(true)

    try {
      const response = await fetch('/api/installations/create-installation', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: data.name,
          description: data.description || undefined,
          subdomain: data.subdomain,
          hostingType: data.hostingType,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to create installation')
      }

      await response.json()
      toast.success('Installation created successfully!')

      // Redirect based on hosting type
      if (data.hostingType === 'kloudlite') {
        router.push('/installations/new/kloudlite-cloud')
      } else {
        router.push('/installations/new/install')
      }
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to create installation')
      toast.error(error.message)
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="lg:flex lg:gap-12">
      {/* Left Column - Information */}
      <div className="hidden lg:block lg:w-[400px] lg:flex-shrink-0">
        <div className="sticky top-6 space-y-6">
          {/* Background Icon */}
          <div className="absolute -top-4 -right-4 -z-10 opacity-5">
            <svg width="300" height="300" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14h-2v-2h2v2zm0-4h-2V7h2v6z"/>
            </svg>
          </div>

          {/* What's Next Card */}
          <div className="border border-foreground/10 rounded-lg p-6 bg-muted/20">
            <h3 className="text-sm font-semibold text-foreground mb-3">What happens next?</h3>
            <div className="space-y-3">
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary/10 text-primary flex items-center justify-center text-xs font-semibold">1</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Create installation</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Set up your installation details</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-semibold">2</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Deploy to cloud</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Install Kloudlite in your infrastructure</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-semibold">3</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Verify & complete</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Confirm your installation is ready</p>
                </div>
              </div>
            </div>
          </div>

          {/* Quick Tips Card */}
          <div className="border border-foreground/10 rounded-lg p-6 bg-background">
            <h3 className="text-sm font-semibold text-foreground mb-3">Quick Tips</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">•</span>
                <span>Choose a memorable name for easy identification</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">•</span>
                <span>Your domain will be accessible at subdomain.khost.dev</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">•</span>
                <span>You can always update these details later</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      {/* Right Column - Form */}
      <div className="space-y-6 lg:flex-1 lg:min-w-0">
        {/* Header */}
        <div>
          <h1 className="text-foreground text-2xl font-semibold tracking-tight">
            Create Installation
          </h1>
          <p className="text-muted-foreground mt-1 text-sm">
            Deploy Kloudlite in your cloud account
          </p>
        </div>

        {/* Form Card */}
        <div className="border border-foreground/10 rounded-lg bg-background">
          <div className="p-8">
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Name</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="e.g., Production"
                          {...field}
                          disabled={creating}
                        />
                      </FormControl>
                      <FormDescription>
                        A friendly name for this installation
                      </FormDescription>
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
                        Description <span className="text-muted-foreground font-normal">(optional)</span>
                      </FormLabel>
                      <FormControl>
                        <Textarea
                          placeholder="Production deployment for our platform"
                          {...field}
                          disabled={creating}
                          rows={4}
                          className="resize-none"
                        />
                      </FormControl>
                      <FormDescription>
                        Optional context about this installation
                      </FormDescription>
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
                          {field.value || 'your-subdomain'}.{process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                        </span>
                        <span className="text-xs font-medium whitespace-nowrap">
                          {!checkingSubdomain && subdomainAvailable === false && (
                            <span className="text-destructive">
                              This domain is already taken
                            </span>
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

                <FormField
                  control={form.control}
                  name="hostingType"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Hosting Type</FormLabel>
                      <FormControl>
                        <RadioGroup
                          onValueChange={field.onChange}
                          defaultValue={field.value}
                          className="grid gap-3"
                          disabled={creating}
                        >
                          <Label
                            htmlFor="hosting-kloudlite"
                            className={`flex items-start gap-3 rounded-lg border p-4 cursor-pointer transition-colors ${
                              field.value === 'kloudlite'
                                ? 'border-primary bg-primary/5'
                                : 'border-foreground/10 hover:border-foreground/20'
                            }`}
                          >
                            <RadioGroupItem value="kloudlite" id="hosting-kloudlite" className="mt-0.5" />
                            <div>
                              <p className="text-sm font-medium text-foreground">Kloudlite Cloud</p>
                              <p className="text-xs text-muted-foreground mt-0.5">
                                We manage the infrastructure for you. No CLI required.
                              </p>
                            </div>
                          </Label>
                          <Label
                            htmlFor="hosting-byoc"
                            className={`flex items-start gap-3 rounded-lg border p-4 cursor-pointer transition-colors ${
                              field.value === 'byoc'
                                ? 'border-primary bg-primary/5'
                                : 'border-foreground/10 hover:border-foreground/20'
                            }`}
                          >
                            <RadioGroupItem value="byoc" id="hosting-byoc" className="mt-0.5" />
                            <div>
                              <p className="text-sm font-medium text-foreground">Bring your own Cloud</p>
                              <p className="text-xs text-muted-foreground mt-0.5">
                                Deploy to your own AWS, GCP, or Azure account using CLI commands.
                              </p>
                            </div>
                          </Label>
                        </RadioGroup>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="flex justify-end pt-4">
                  <Button
                    type="submit"
                    size="default"
                    disabled={creating || subdomainAvailable !== true}
                  >
                    {creating ? (
                      <>
                        <Loader2 className="mr-2 size-4 animate-spin" />
                        Creating...
                      </>
                    ) : (
                      'Continue to Installation'
                    )}
                  </Button>
                </div>
              </form>
            </Form>
          </div>
        </div>

        {/* Help text */}
        <div className="flex items-start gap-3 text-sm text-muted-foreground">
          <div className="flex-shrink-0 mt-0.5">
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <p>
              Need help getting started?{' '}
              <a
                href="https://docs.kloudlite.io"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary relative inline-block"
              >
                <span className="relative">
                  View our documentation
                  <span className="absolute -bottom-0.5 left-0 right-0 h-0.5 bg-primary scale-x-0 hover:scale-x-100 transition-transform duration-300 origin-left" />
                </span>
              </a>
              {' '}for detailed installation guides.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
