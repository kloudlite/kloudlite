'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button, Input, Textarea, Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@kloudlite/ui'
import { Loader2, CheckCircle2, AlertCircle } from 'lucide-react'
import { InstallationProgress } from '@/components/installation-progress'
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
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to create installation')
      }

      await response.json()
      toast.success('Installation created successfully!')

      // Redirect to install step
      router.push('/installations/new/install')
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to create installation')
      toast.error(error.message)
    } finally {
      setCreating(false)
    }
  }

  return (
    <>
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-foreground text-2xl font-semibold tracking-tight">
          Create Installation
        </h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Deploy Kloudlite in your cloud account
        </p>
      </div>

      <InstallationProgress currentStep={1} />

      <div className="mt-6">
        <div>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
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
                    <FormDescription>
                      <span className="font-mono">
                        {field.value || 'your-subdomain'}.{process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                      </span>
                    </FormDescription>
                    {!checkingSubdomain && subdomainAvailable === false && (
                      <p className="text-[0.8rem] font-medium text-destructive">
                        This domain is already taken
                      </p>
                    )}
                    {!checkingSubdomain && subdomainAvailable === true && (
                      <p className="text-[0.8rem] font-medium text-green-600 dark:text-green-500">
                        Domain is available
                      </p>
                    )}
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="flex justify-end pt-2">
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
    </>
  )
}
