import { getInstallationById } from '@/lib/console/storage'
import { KlCloudInstallationForm } from '@/components/kl-cloud-installation-form'
import { getTierConfig } from '@/lib/stripe-bootstrap'
import { CheckCircle2 } from 'lucide-react'

interface NewKlCloudPageProps {
  searchParams: Promise<{ installation?: string }>
}

export default async function NewKlCloudPage({ searchParams }: NewKlCloudPageProps) {
  const params = await searchParams

  // If continuing an existing installation, fetch it
  const existingInstallation = params.installation
    ? await getInstallationById(params.installation)
    : null

  return (
    <div className="lg:flex lg:gap-12">
      {/* Left Column - Information */}
      <div className="hidden lg:block lg:w-[400px] lg:flex-shrink-0">
        <div className="sticky top-6 space-y-6">
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
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary/10 text-primary flex items-center justify-center">
                  <CheckCircle2 className="w-3 h-3" />
                </div>
                <div>
                  <p className="text-sm font-medium text-foreground">Choose hosting type</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Kloudlite Cloud selected</p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-semibold">2</div>
                <div>
                  <p className="text-sm font-medium text-foreground">
                    {existingInstallation ? 'Subscribe & pay' : 'Configure installation'}
                  </p>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {existingInstallation ? 'Choose compute sizes & payment' : 'Name, domain, plan & payment'}
                  </p>
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-shrink-0 w-5 h-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-semibold">3</div>
                <div>
                  <p className="text-sm font-medium text-foreground">Auto-deploy & verify</p>
                  <p className="text-xs text-muted-foreground mt-0.5">We&apos;ll deploy automatically for you</p>
                </div>
              </div>
            </div>
          </div>

          {/* Quick Tips Card */}
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
              <li className="flex items-start gap-2">
                <span className="text-primary mt-0.5">&bull;</span>
                <span>No CLI or cloud credentials needed</span>
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
            {existingInstallation ? 'Complete Subscription' : 'Create Installation'}
          </h1>
          <p className="text-muted-foreground mt-1 text-sm">
            {existingInstallation
              ? `Choose compute sizes and subscribe for "${existingInstallation.name}"`
              : 'Set up your Kloudlite Cloud installation'}
          </p>
        </div>

        <KlCloudInstallationForm
          existingInstallationId={existingInstallation?.id}
          tierConfig={getTierConfig('usd')}
          currency="usd"
        />

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
