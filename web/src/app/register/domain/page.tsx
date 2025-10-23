'use client'

import { useState, useEffect } from 'react'
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
import { KloudliteLogo } from '@/components/kloudlite-logo'
import { toast } from 'sonner'

interface SessionData {
  user: {
    email: string
    name: string
  }
  installationKey: string
}

const subdomainSchema = z.object({
  subdomain: z
    .string()
    .min(3, 'Subdomain must be at least 3 characters')
    .max(63, 'Subdomain must be less than 63 characters')
    .regex(
      /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/,
      'Subdomain must start and end with alphanumeric characters and can only contain lowercase letters, numbers, and hyphens'
    ),
})

type SubdomainFormData = z.infer<typeof subdomainSchema>

export default function DomainPage() {
  const router = useRouter()
  const [session, setSession] = useState<SessionData | null>(null)
  const [loading, setLoading] = useState(true)
  const [checkingSubdomain, setCheckingSubdomain] = useState(false)
  const [subdomainAvailable, setSubdomainAvailable] = useState<boolean | null>(null)
  const [saving, setSaving] = useState(false)

  const form = useForm<SubdomainFormData>({
    resolver: zodResolver(subdomainSchema),
    defaultValues: {
      subdomain: '',
    },
  })

  useEffect(() => {
    // Check for registration session cookie
    // Middleware handles all redirects based on state
    const checkSession = async () => {
      try {
        const response = await fetch('/api/register/session')
        if (response.ok) {
          const data = await response.json()
          setSession(data)
        } else {
          router.push('/register')
        }
      } catch {
        router.push('/register')
      } finally {
        setLoading(false)
      }
    }
    checkSession()
  }, [router])

  const checkSubdomainAvailability = async (subdomain: string) => {
    if (!subdomain || subdomain.length < 3) {
      setSubdomainAvailable(null)
      return
    }

    setCheckingSubdomain(true)
    try {
      const response = await fetch(`/api/register/check-subdomain?subdomain=${subdomain}`)
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
      const response = await fetch('/api/register/reserve-subdomain', {
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

      // Redirect immediately to complete page
      router.push('/register/complete')
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to reserve subdomain')
      toast.error(error.message)
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="size-8 animate-spin text-primary" />
      </div>
    )
  }

  if (!session) {
    return null
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-8 bg-background">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <KloudliteLogo className="mx-auto mb-6" />
          <h1 className="text-2xl font-medium text-foreground mb-2">
            Choose Your Domain
          </h1>
          <p className="text-sm text-muted-foreground">
            Select a subdomain for your Kloudlite workspace
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Workspace Domain</CardTitle>
            <CardDescription>
              Your team will access Kloudlite at this domain
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
                            className="h-10 px-3 text-sm pr-10"
                            onChange={(e) => {
                              field.onChange(e)
                              checkSubdomainAvailability(e.target.value)
                            }}
                          />
                          {checkingSubdomain && (
                            <div className="absolute right-3 top-1/2 -translate-y-1/2">
                              <Loader2 className="size-4 animate-spin text-muted-foreground" />
                            </div>
                          )}
                          {!checkingSubdomain && subdomainAvailable === true && (
                            <div className="absolute right-3 top-1/2 -translate-y-1/2">
                              <CheckCircle2 className="size-4 text-green-600" />
                            </div>
                          )}
                          {!checkingSubdomain && subdomainAvailable === false && (
                            <div className="absolute right-3 top-1/2 -translate-y-1/2">
                              <AlertCircle className="size-4 text-destructive" />
                            </div>
                          )}
                        </div>
                      </FormControl>
                      <FormDescription className="text-xs">
                        Your workspace will be available at {field.value || 'your-subdomain'}.kloudlite.io
                      </FormDescription>
                      {!checkingSubdomain && subdomainAvailable === false && (
                        <p className="text-sm text-destructive">This subdomain is already taken</p>
                      )}
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <Button
                  type="submit"
                  className="w-full"
                  disabled={saving || subdomainAvailable === false}
                >
                  {saving ? (
                    <>
                      <Loader2 className="size-4 animate-spin mr-2" />
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
      </div>
    </div>
  )
}
