'use client'

import { useState, useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'
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
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  KloudliteLogo,
} from '@kloudlite/ui'
import { CheckCircle2, Loader2, AlertCircle, ArrowLeft, AlertTriangle } from 'lucide-react'
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

interface InstallationInfo {
  name: string
  description: string | null
  oldSubdomain: string
  claimedByEmail?: string
  claimedByName?: string
}

export default function ReselectDomainPage() {
  const router = useRouter()
  const params = useParams()
  const installationId = params.id as string

  const [checkingSubdomain, setCheckingSubdomain] = useState(false)
  const [subdomainAvailable, setSubdomainAvailable] = useState<boolean | null>(null)
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState(true)
  const [installationInfo, setInstallationInfo] = useState<InstallationInfo | null>(null)
  const [error, setError] = useState<string | null>(null)

  const form = useForm<SubdomainFormData>({
    resolver: zodResolver(subdomainSchema),
    defaultValues: {
      subdomain: '',
    },
  })

  useEffect(() => {
    async function fetchInstallationInfo() {
      try {
        const response = await fetch(`/api/installations/${installationId}/domain-status`)
        if (!response.ok) {
          if (response.status === 404) {
            router.push('/installations')
            return
          }
          throw new Error('Failed to fetch installation info')
        }
        const data = await response.json()

        if (!data.needsReselection) {
          // Domain is still valid, redirect back to installation page
          router.push(`/installations/${installationId}`)
          return
        }

        setInstallationInfo({
          name: data.name,
          description: data.description,
          oldSubdomain: data.oldSubdomain,
          claimedByEmail: data.claimedByEmail,
          claimedByName: data.claimedByName,
        })
      } catch (err) {
        console.error('Error fetching installation info:', err)
        setError('Failed to load installation information')
      } finally {
        setLoading(false)
      }
    }

    fetchInstallationInfo()
  }, [installationId, router])

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
      const response = await fetch(`/api/installations/${installationId}/re-reserve-domain`, {
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
      toast.success('Domain updated successfully!')

      // Redirect back to installation page
      router.push(`/installations/${installationId}`)
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to reserve subdomain')
      toast.error(error.message)
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="bg-background min-h-screen p-8">
        <div className="mx-auto w-full max-w-3xl">
          <div className="flex items-center justify-center py-20">
            <Loader2 className="text-muted-foreground size-8 animate-spin" />
          </div>
        </div>
      </div>
    )
  }

  if (error || !installationInfo) {
    return (
      <div className="bg-background min-h-screen p-8">
        <div className="mx-auto w-full max-w-3xl">
          <div className="flex flex-col items-center justify-center py-20">
            <AlertCircle className="text-destructive mb-4 size-12" />
            <p className="text-destructive text-lg">{error || 'Installation not found'}</p>
            <Button variant="outline" className="mt-4" onClick={() => router.push('/installations')}>
              Back to Installations
            </Button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-background min-h-screen p-8">
      <div className="mx-auto w-full max-w-3xl">
        {/* Back button */}
        <div className="mb-6">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.push(`/installations/${installationId}`)}
            className="gap-2"
          >
            <ArrowLeft className="size-4" />
            Back to Installation
          </Button>
        </div>

        {/* Logo */}
        <div className="mb-8 flex items-center justify-center">
          <KloudliteLogo />
        </div>

        {/* Header */}
        <div className="mb-8 text-center">
          <h1 className="text-foreground mb-3 text-4xl font-bold tracking-tight">
            Update Your Domain
          </h1>
          <p className="text-muted-foreground text-lg">
            Your previous domain is no longer available. Please choose a new one.
          </p>
        </div>

        {/* Warning about claimed domain */}
        <Card className="mb-6 border-amber-200 bg-amber-50 dark:border-amber-900 dark:bg-amber-950">
          <CardContent className="pt-6">
            <div className="flex items-start gap-3">
              <AlertTriangle className="mt-0.5 size-5 shrink-0 text-amber-600 dark:text-amber-400" />
              <div>
                <p className="font-semibold text-amber-900 dark:text-amber-200">
                  Domain No Longer Available
                </p>
                <p className="mt-1 text-sm text-amber-800 dark:text-amber-300">
                  Your previously reserved domain{' '}
                  <span className="font-mono font-medium">{installationInfo.oldSubdomain}</span> has
                  expired and been claimed by another user
                  {installationInfo.claimedByName && (
                    <> ({installationInfo.claimedByName})</>
                  )}
                  . Please choose a new subdomain to continue with your installation.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Installation Info (read-only) */}
        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="text-lg">Installation Details</CardTitle>
            <CardDescription>These details cannot be changed</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="text-muted-foreground text-sm font-medium">Name</label>
              <p className="text-foreground mt-1">{installationInfo.name}</p>
            </div>
            {installationInfo.description && (
              <div>
                <label className="text-muted-foreground text-sm font-medium">Description</label>
                <p className="text-foreground mt-1">{installationInfo.description}</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Domain Selection */}
        <Card>
          <CardHeader>
            <CardTitle className="text-xl">Choose a New Domain</CardTitle>
            <CardDescription>
              Select a new subdomain for your Kloudlite installation. Your team will access
              Kloudlite at this domain.
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
                      Updating domain...
                    </>
                  ) : (
                    'Update Domain'
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
