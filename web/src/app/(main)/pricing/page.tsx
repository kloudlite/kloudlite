import { Button } from '@/components/ui/button'
import Link from 'next/link'
import { APP_MODE } from '@/lib/app-mode'
import { Check } from 'lucide-react'
import { GetStartedButton } from '@/components/get-started-button'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'

// Pricing page for website mode
function PricingPage() {
  return (
    <div className="bg-background flex min-h-screen flex-col">
      <WebsiteHeader currentPage="pricing" />

      {/* Pricing Section */}
      <main className="flex-1 px-4 py-16 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-[90rem]">
          {/* Header */}
          <div className="text-center">
            <h1 className="text-foreground text-4xl font-bold tracking-tight sm:text-5xl">
              Simple, Transparent Pricing
            </h1>
            <p className="text-muted-foreground mx-auto mt-4 max-w-2xl text-lg">
              Choose the plan that fits your team&apos;s needs
            </p>
          </div>

          {/* Pricing Cards */}
          <div className="mt-16 grid gap-8 lg:grid-cols-3">
            {/* BYOC */}
            <div className="bg-card flex flex-col rounded-2xl border p-8">
              <div className="flex-1">
                <h3 className="text-card-foreground text-2xl font-bold">BYOC</h3>
                <p className="text-muted-foreground mt-2 text-sm">Bring Your Own Cloud</p>
                <div className="mt-6">
                  <span className="text-card-foreground text-5xl font-bold">Free</span>
                </div>
                <ul className="mt-8 space-y-4">
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">
                      Deploy on AWS, Azure, or GCP
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Unlimited workspaces</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Unlimited environments</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">All core features</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Community support</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Open source</span>
                  </li>
                </ul>
              </div>
              <GetStartedButton size="lg" className="mt-8 w-full" />
            </div>

            {/* Cloud */}
            <div className="bg-card flex flex-col rounded-2xl border p-8">
              <div className="flex-1">
                <h3 className="text-card-foreground text-2xl font-bold">Cloud</h3>
                <p className="text-muted-foreground mt-2 text-sm">Managed infrastructure for you</p>
                <div className="mt-6 space-y-3">
                  <div className="bg-muted rounded-lg border p-3">
                    <div className="flex items-baseline gap-2">
                      <span className="text-foreground text-2xl font-bold">$29</span>
                      <span className="text-muted-foreground text-sm font-medium">/team/month</span>
                    </div>
                    <p className="text-muted-foreground mt-1 text-xs">Base fee</p>
                  </div>
                  <div className="flex items-center justify-center">
                    <span className="text-muted-foreground text-2xl font-bold">+</span>
                  </div>
                  <div className="border-info bg-info/10 rounded-lg border-2 p-3">
                    <div className="flex items-baseline gap-2">
                      <span className="text-foreground text-3xl font-bold">$49</span>
                      <span className="text-muted-foreground text-base font-medium">
                        /user/month
                      </span>
                    </div>
                    <p className="text-muted-foreground mt-1 text-xs">Per active user</p>
                  </div>
                </div>
                <ul className="mt-8 space-y-4">
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Everything in BYOC</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">
                      160 hours usage per user per month
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">
                      Workmachine: 12 vCPU and 16GB RAM
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Workmachine: 200GB storage</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">
                      $0.30/hour beyond 160 hours
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Managed infrastructure</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">
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
            <div className="bg-card flex flex-col rounded-2xl border p-8">
              <div className="flex-1">
                <h3 className="text-card-foreground text-2xl font-bold">Enterprise</h3>
                <p className="text-muted-foreground mt-2 text-sm">
                  Advanced features for large teams
                </p>
                <div className="mt-6">
                  <span className="text-card-foreground text-5xl font-bold">Custom</span>
                </div>
                <ul className="mt-8 space-y-4">
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Everything in Cloud</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">On-premise deployment</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Organisation Management</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">SSO & advanced auth</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Custom integrations</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Dedicated support</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="text-success h-5 w-5 flex-shrink-0" />
                    <span className="text-card-foreground text-sm">Custom SLA</span>
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
            <h2 className="text-foreground text-center text-3xl font-bold">
              Frequently Asked Questions
            </h2>
            <div className="mx-auto mt-12 max-w-3xl space-y-8">
              <div>
                <h3 className="text-foreground text-lg font-semibold">
                  What&apos;s included in the free BYOC plan?
                </h3>
                <p className="text-muted-foreground mt-2">
                  The BYOC (Bring Your Own Cloud) plan includes all core features of Kloudlite with
                  no limitations on workspaces or environments. You deploy it on your own AWS,
                  Azure, or GCP infrastructure.
                </p>
              </div>
              <div>
                <h3 className="text-foreground text-lg font-semibold">
                  Can I try the Cloud plan before committing?
                </h3>
                <p className="text-muted-foreground mt-2">
                  Yes! We&apos;ll give you 15 days of access for one developer to try out the Cloud
                  plan.
                </p>
              </div>
              <div>
                <h3 className="text-foreground text-lg font-semibold">How does billing work?</h3>
                <p className="text-muted-foreground mt-2">
                  Pricing is $49 per user per month, plus a $29 base fee per team. For example, a
                  team with 5 users pays $274/month ($29 base + 5 × $49). Each user gets 160 hours
                  of usage, a workmachine with 12 vCPU and 16GB RAM, and 200GB storage. Usage beyond
                  160 hours per user is billed at $0.30/hour.
                </p>
              </div>
              <div>
                <h3 className="text-foreground text-lg font-semibold">
                  What happens if I exceed 160 hours?
                </h3>
                <p className="text-muted-foreground mt-2">
                  Any usage beyond 160 hours per user per month is automatically billed at $0.30 per
                  hour. You can monitor your usage in the dashboard to track consumption.
                </p>
              </div>
              <div>
                <h3 className="text-foreground text-lg font-semibold">
                  What&apos;s the difference between Cloud and Enterprise?
                </h3>
                <p className="text-muted-foreground mt-2">
                  Enterprise includes on-premise deployment, advanced authentication options (SSO,
                  SAML), custom integrations, dedicated support, and custom SLAs for large
                  organizations.
                </p>
              </div>
            </div>
          </div>
        </div>
      </main>

      <WebsiteFooter />
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
