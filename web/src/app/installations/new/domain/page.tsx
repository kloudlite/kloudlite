'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from '@/components/ui/form'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { CheckCircle2, Loader2, AlertCircle } from 'lucide-react'
import { InstallationProgress } from '@/components/installation-progress'
import { toast } from 'sonner'

const subdomainSchema = z.object({
  subdomain: z
    .string()
    .min(3, 'Subdomain must be at least 3 characters')
    .max(63, 'Subdomain must be less than 63 characters')
    .regex(
      /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/,
      'Subdomain must start and end with alphanumeric characters and can only contain lowercase letters, numbers, and hyphens',
    ),
})

type SubdomainFormData = z.infer<typeof subdomainSchema>

export default function ConfigureDomainPage() {
  const router = useRouter()
  const [checkingSubdomain, setCheckingSubdomain] = useState(false)
  const [subdomainAvailable, setSubdomainAvailable] = useState<boolean | null>(null)
  const [saving, setSaving] = useState(false)

  const form = useForm<SubdomainFormData>({
    resolver: zodResolver(subdomainSchema),
    defaultValues: {
      subdomain: '',
    },
  })

  const checkSubdomainAvailability = async (subdomain: string) => {
    if (!subdomain || subdomain.length < 3) {
      setSubdomainAvailable(null)
      return
    }

    setCheckingSubdomain(true)
    try {
      const response = await fetch(`/api/installations/check-subdomain?subdomain=${subdomain}`)
      const data = await response.json()
      setSubdomainAvailable(data.available)
    } catch (err) {
      console.error('Error checking subdomain:', err)
      setSubdomainAvailable(false)
    } finally {
      setCheckingSubdomain(false)
    }
  }

  const onSubmit = async (data: SubdomainFormData) => {
    setSaving(true)

    try {
      const response = await fetch('/api/installations/reserve-subdomain', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          subdomain: data.subdomain,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to reserve subdomain')
      }

      await response.json()
      toast.success('Domain configured successfully!')

      // Redirect to complete step
      router.push('/installations/new/complete')
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to reserve subdomain')
      toast.error(error.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      {/* Header */}
      <div className="mb-12 text-center">
        <h1 className="text-foreground mb-3 text-4xl font-bold tracking-tight">
          Configure Your Domain
        </h1>
        <p className="text-muted-foreground text-lg">Choose a domain name for your installation</p>
      </div>

      <InstallationProgress currentStep={3} />

      <Card>
        <CardHeader>
          <CardTitle className="text-xl">Choose Your Installation Domain</CardTitle>
          <CardDescription>
            Select a subdomain for your Kloudlite installation. Your team will access Kloudlite at
            this domain.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="subdomain"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Subdomain</FormLabel>
                    <FormControl>
                      <div className="relative">
                        <Input
                          placeholder="your-company"
                          {...field}
                          disabled={saving}
                          className="h-11 px-4 pr-10 text-base"
                          onChange={(e) => {
                            field.onChange(e)
                            checkSubdomainAvailability(e.target.value)
                          }}
                        />
                        {checkingSubdomain && (
                          <div className="absolute top-1/2 right-3 -translate-y-1/2">
                            <Loader2 className="text-muted-foreground size-4 animate-spin" />
                          </div>
                        )}
                        {!checkingSubdomain && subdomainAvailable === true && (
                          <div className="absolute top-1/2 right-3 -translate-y-1/2">
                            <CheckCircle2 className="size-5 text-green-600" />
                          </div>
                        )}
                        {!checkingSubdomain && subdomainAvailable === false && (
                          <div className="absolute top-1/2 right-3 -translate-y-1/2">
                            <AlertCircle className="text-destructive size-5" />
                          </div>
                        )}
                      </div>
                    </FormControl>
                    <FormDescription>
                      Your installation will be available at{' '}
                      <span className="font-mono font-medium">
                        {field.value || 'your-subdomain'}.
                        {process.env.NEXT_PUBLIC_INSTALLATION_DOMAIN || 'khost.dev'}
                      </span>
                    </FormDescription>
                    {!checkingSubdomain && subdomainAvailable === false && (
                      <p className="text-destructive text-sm font-medium">
                        This subdomain is already taken. Please choose another.
                      </p>
                    )}
                    {!checkingSubdomain && subdomainAvailable === true && (
                      <p className="text-sm font-medium text-green-600">
                        This subdomain is available!
                      </p>
                    )}
                    <FormMessage />
                  </FormItem>
                )}
              />

              <Button
                type="submit"
                className="w-full"
                size="lg"
                disabled={saving || subdomainAvailable !== true}
              >
                {saving ? (
                  <>
                    <Loader2 className="mr-2 size-4 animate-spin" />
                    Reserving domain...
                  </>
                ) : (
                  'Complete Setup'
                )}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>
    </>
  )
}
