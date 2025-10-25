import { Button } from '@/components/ui/button'
import { KloudliteLogo } from '@/components/kloudlite-logo'
import Link from 'next/link'
import { APP_MODE } from '@/lib/app-mode'
import { Check } from 'lucide-react'
import { ThemeSwitcherServer } from '@/components/theme-switcher-server'
import { GetStartedButton } from '@/components/get-started-button'

// Pricing page for website mode
function PricingPage() {
  return (
    <div className="flex min-h-screen flex-col bg-background">
      {/* Navigation Header */}
      <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <nav className="mx-auto flex h-16 max-w-[90rem] items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-6 lg:gap-8">
            <KloudliteLogo showText={true} linkToHome={true} />
            <div className="hidden items-center gap-6 md:flex">
              <Link
                href="/docs"
                className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                Docs
              </Link>
              <Link
                href="/pricing"
                className="text-sm font-medium text-foreground"
              >
                Pricing
              </Link>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <GetStartedButton size="sm" className="hidden sm:flex" />
          </div>
        </nav>
      </header>

      {/* Pricing Section */}
      <main className="flex-1 px-4 py-16 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-[90rem]">
          {/* Header */}
          <div className="text-center">
            <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
              Simple, Transparent Pricing
            </h1>
            <p className="mx-auto mt-4 max-w-2xl text-lg text-muted-foreground">
              Choose the plan that fits your team&apos;s needs
            </p>
          </div>

          {/* Pricing Cards */}
          <div className="mt-16 grid gap-8 lg:grid-cols-3">
            {/* BYOC */}
            <div className="flex flex-col rounded-2xl border bg-card p-8">
              <div className="flex-1">
                <h3 className="text-2xl font-bold text-card-foreground">BYOC</h3>
                <p className="mt-2 text-sm text-muted-foreground">
                  Bring Your Own Cloud
                </p>
                <div className="mt-6">
                  <span className="text-5xl font-bold text-card-foreground">Free</span>
                </div>
                <ul className="mt-8 space-y-4">
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Deploy on AWS, Azure, or GCP
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Unlimited workspaces
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Unlimited environments
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      All core features
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Community support
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Open source
                    </span>
                  </li>
                </ul>
              </div>
              <GetStartedButton size="lg" className="mt-8 w-full" />
            </div>

            {/* Cloud */}
            <div className="flex flex-col rounded-2xl border bg-card p-8">
              <div className="flex-1">
                <h3 className="text-2xl font-bold text-card-foreground">Cloud</h3>
                <p className="mt-2 text-sm text-muted-foreground">
                  Managed infrastructure for you
                </p>
                <div className="mt-6 space-y-3">
                  <div className="rounded-lg border bg-muted p-3">
                    <div className="flex items-baseline gap-2">
                      <span className="text-2xl font-bold text-foreground">$29</span>
                      <span className="text-sm font-medium text-muted-foreground">
                        /team/month
                      </span>
                    </div>
                    <p className="mt-1 text-xs text-muted-foreground">Base fee</p>
                  </div>
                  <div className="flex items-center justify-center">
                    <span className="text-2xl font-bold text-muted-foreground">+</span>
                  </div>
                  <div className="rounded-lg border-2 border-info bg-info/10 p-3">
                    <div className="flex items-baseline gap-2">
                      <span className="text-3xl font-bold text-foreground">$49</span>
                      <span className="text-base font-medium text-muted-foreground">
                        /user/month
                      </span>
                    </div>
                    <p className="mt-1 text-xs text-muted-foreground">Per active user</p>
                  </div>
                </div>
                <ul className="mt-8 space-y-4">
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Everything in BYOC
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      160 hours usage per user per month
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Workmachine: 12 vCPU and 16GB RAM
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Workmachine: 200GB storage
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      $0.30/hour beyond 160 hours
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Managed infrastructure
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Priority support & 99.9% SLA
                    </span>
                  </li>
                </ul>
              </div>
              <Button size="lg" className="mt-8 w-full" disabled>
                Coming Soon
              </Button>
            </div>

            {/* Enterprise */}
            <div className="flex flex-col rounded-2xl border bg-card p-8">
              <div className="flex-1">
                <h3 className="text-2xl font-bold text-card-foreground">Enterprise</h3>
                <p className="mt-2 text-sm text-muted-foreground">
                  Advanced features for large teams
                </p>
                <div className="mt-6">
                  <span className="text-5xl font-bold text-card-foreground">Custom</span>
                </div>
                <ul className="mt-8 space-y-4">
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Everything in Cloud
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      On-premise deployment
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Organisation Management
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      SSO & advanced auth
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Custom integrations
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Dedicated support
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="h-5 w-5 flex-shrink-0 text-success" />
                    <span className="text-sm text-card-foreground">
                      Custom SLA
                    </span>
                  </li>
                </ul>
              </div>
              <Button asChild variant="outline" size="lg" className="mt-8 w-full">
                <Link href="/contact">Contact Sales</Link>
              </Button>
            </div>
          </div>

          {/* FAQ Section */}
          <div className="mt-24">
            <h2 className="text-center text-3xl font-bold text-foreground">
              Frequently Asked Questions
            </h2>
            <div className="mx-auto mt-12 max-w-3xl space-y-8">
              <div>
                <h3 className="text-lg font-semibold text-foreground">
                  What&apos;s included in the free BYOC plan?
                </h3>
                <p className="mt-2 text-muted-foreground">
                  The BYOC (Bring Your Own Cloud) plan includes all core features of Kloudlite with no
                  limitations on workspaces or environments. You deploy it on your own AWS, Azure, or
                  GCP infrastructure.
                </p>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-foreground">
                  Can I try the Cloud plan before committing?
                </h3>
                <p className="mt-2 text-muted-foreground">
                  Yes! We&apos;ll give you 15 days of access for one developer to try out the Cloud plan.
                </p>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-foreground">
                  How does billing work?
                </h3>
                <p className="mt-2 text-muted-foreground">
                  Pricing is $49 per user per month, plus a $29 base fee per team. For example, a team
                  with 5 users pays $274/month ($29 base + 5 × $49). Each user gets 160 hours of usage,
                  a workmachine with 12 vCPU and 16GB RAM, and 200GB storage. Usage beyond 160 hours per
                  user is billed at $0.30/hour.
                </p>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-foreground">
                  What happens if I exceed 160 hours?
                </h3>
                <p className="mt-2 text-muted-foreground">
                  Any usage beyond 160 hours per user per month is automatically billed at $0.30 per
                  hour. You can monitor your usage in the dashboard to track consumption.
                </p>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-foreground">
                  What&apos;s the difference between Cloud and Enterprise?
                </h3>
                <p className="mt-2 text-muted-foreground">
                  Enterprise includes on-premise deployment, advanced authentication options (SSO,
                  SAML), custom integrations, dedicated support, and custom SLAs for large
                  organizations.
                </p>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t bg-muted">
        <div className="mx-auto max-w-[90rem] px-4 py-12 sm:px-6 lg:px-8">
          <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
            <p className="text-sm text-muted-foreground">© 2024 Kloudlite. All rights reserved.</p>
            <div className="flex items-center gap-6">
              <Link href="/docs" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
                Docs
              </Link>
              <Link href="/pricing" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
                Pricing
              </Link>
              <Link href="/contact" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
                Contact
              </Link>
              <Link href="https://github.com/kloudlite/kloudlite" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
                GitHub
              </Link>
              <ThemeSwitcherServer />
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default function Page() {
  // Only show pricing page in website mode
  if (APP_MODE === 'website') {
    return <PricingPage />
  }

  // Redirect to home for other modes
  return (
    <div className="flex min-h-screen items-center justify-center">
      <p className="text-muted-foreground">Pricing page is only available in website mode.</p>
    </div>
  )
}
