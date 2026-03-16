'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Form, Tabs, TabsList, TabsTrigger, TabsContent } from '@kloudlite/ui'
import { toast } from 'sonner'
import { getErrorMessage } from '@/lib/errors'
import { useSubdomainCheck } from '@/hooks/use-subdomain-check'
import { InstallationFields } from '@/components/installation-fields'
import { KlCloudTab } from '@/components/kl-cloud-tab'
import { ByocTab } from '@/components/byoc-tab'

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

interface CreateInstallationPageProps {
  orgId: string
}

export function CreateInstallationPage({ orgId }: CreateInstallationPageProps) {
  const searchParams = useSearchParams()
  const [creating, setCreating] = useState(false)
  const [activeTab, setActiveTab] = useState('kloudlite-cloud')

  const {
    checking: checkingSubdomain,
    available: subdomainAvailable,
    check: checkSubdomainAvailability,
  } = useSubdomainCheck({ endpoint: '/api/installations/check-domain-kli' })

  // Restore form values from URL params (after payment redirect)
  const restoredName = searchParams.get('name') || ''
  const restoredSubdomain = searchParams.get('subdomain') || ''

  const form = useForm<InstallationFormData>({
    resolver: zodResolver(installationSchema),
    defaultValues: {
      name: restoredName,
      subdomain: restoredSubdomain,
    },
  })

  // On return from payment, verify checkout session and restore subdomain check
  useEffect(() => {
    if (restoredSubdomain) {
      checkSubdomainAvailability(restoredSubdomain)
    }

  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const checkoutSessionId = searchParams.get('checkout_session') || undefined
  const paymentSuccess = searchParams.get('payment') === 'success'

  const createInstallation = async (hostingType: 'kloudlite' | 'byoc') => {
    const valid = await form.trigger()
    if (!valid) return

    if (subdomainAvailable !== true) {
      toast.error('Please choose an available subdomain')
      return
    }

    setCreating(true)
    const data = form.getValues()

    try {
      const response = await fetch('/api/installations/create-installation', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          orgId,
          name: data.name,
          subdomain: data.subdomain,
          hostingType,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to create installation')
      }

      const result = await response.json()
      window.location.href = `/api/installations/${result.installationId}/continue`
    } catch (err) {
      toast.error(getErrorMessage(err, 'Failed to create installation'))
      setCreating(false)
    }
  }

  return (
    <div className="lg:flex lg:gap-12">
      {/* Left Column - Steps & Tips */}
      <div className="hidden lg:block lg:w-[400px] lg:flex-shrink-0">
        <div className="sticky top-6 space-y-6">
          <div className="border border-foreground/10 rounded-lg p-6 bg-muted/20">
            <h3 className="text-sm font-semibold text-foreground mb-3">What happens next?</h3>
            <div className="space-y-3">
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-semibold">1</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Configure installation</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Name, domain & hosting type</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-semibold">2</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Deploy & verify</p>
                  <p className="text-xs text-muted-foreground mt-0.5">We&apos;ll deploy and verify for you</p>
                </div>
              </div>
            </div>
          </div>

          <div className="border border-foreground/10 rounded-lg p-6 bg-background">
            <h3 className="text-sm font-semibold text-foreground mb-3">Quick Tips</h3>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&bull;</span>
                <span>Choose a memorable name for easy identification</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&bull;</span>
                <span>Your domain will be accessible at subdomain.khost.dev</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      {/* Right Column - Form */}
      <div className="space-y-6 lg:flex-1 lg:min-w-0">
        <div>
          <h1 className="text-foreground text-2xl font-semibold tracking-tight">Create Installation</h1>
          <p className="text-muted-foreground mt-1 text-sm">Set up a new Kloudlite installation</p>
        </div>

        <Form {...form}>
          <form onSubmit={(e) => e.preventDefault()} className="space-y-6">
            {/* Shared fields */}
            <InstallationFields
              control={form.control}
              creating={creating}
              checkingSubdomain={checkingSubdomain}
              subdomainAvailable={subdomainAvailable}
              onSubdomainChange={checkSubdomainAvailability}
            />

            {/* Type tabs */}
            <Tabs value={activeTab} onValueChange={setActiveTab}>
              <TabsList className="inline-flex gap-1 rounded-lg bg-muted/50 p-1">
                <TabsTrigger
                  value="kloudlite-cloud"
                  className="rounded-md px-4 py-2 text-sm data-[state=active]:bg-background data-[state=active]:shadow-sm"
                >
                  Kloudlite Cloud
                </TabsTrigger>
                <TabsTrigger
                  value="byoc"
                  className="rounded-md px-4 py-2 text-sm data-[state=active]:bg-background data-[state=active]:shadow-sm"
                >
                  Bring your own Cloud
                </TabsTrigger>
              </TabsList>

              <TabsContent value="kloudlite-cloud" className="mt-4">
                <KlCloudTab
                  orgId={orgId}
                  creating={creating}
                  subdomainAvailable={subdomainAvailable}
                  onSubmit={() => createInstallation('kloudlite')}
                  getFormValues={() => form.getValues()}
                  checkoutSessionId={paymentSuccess ? checkoutSessionId : undefined}
                />
              </TabsContent>

              <TabsContent value="byoc" className="mt-4">
                <ByocTab
                  creating={creating}
                  subdomainAvailable={subdomainAvailable}
                  onSubmit={() => createInstallation('byoc')}
                />
              </TabsContent>
            </Tabs>
          </form>
        </Form>

        {/* Help text */}
        <div className="flex items-start gap-3 text-sm text-muted-foreground">
          <svg className="h-4 w-4 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p>
            Need help getting started?{' '}
            <a href="https://docs.kloudlite.io" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
              View our documentation
            </a>{' '}
            for detailed installation guides.
          </p>
        </div>
      </div>
    </div>
  )
}
