'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button, Input, Textarea, Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@kloudlite/ui'
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
})

type InstallationFormData = z.infer<typeof installationSchema>

interface InstallationFormProps {
  hostingType: 'kloudlite' | 'byoc'
  redirectTo: string
}

export function InstallationForm({ hostingType, redirectTo }: InstallationFormProps) {
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
          hostingType,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to create installation')
      }

      await response.json()
      toast.success('Installation created successfully!')
      router.push(redirectTo)
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to create installation')
      toast.error(error.message)
    } finally {
      setCreating(false)
    }
  }

  return (
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
  )
}
