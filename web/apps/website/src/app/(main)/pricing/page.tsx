import { Button } from '@kloudlite/ui'
import Link from 'next/link'
import { Check, Server, Users } from 'lucide-react'
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
              Choose the plan that fits your needs
            </p>
          </div>

          {/* BYOC Section */}
          <div className="mt-16">
            <h2 className="text-foreground mb-4 text-center text-2xl font-bold">
              BYOC (Bring Your Own Compute)
            </h2>
            <p className="text-muted-foreground mx-auto mb-8 max-w-2xl text-center">
              Deploy on your own infrastructure - AWS, Azure, or GCP
            </p>

            {/* BYOC Pricing Cards */}
            <div className="mx-auto grid max-w-4xl gap-8 lg:grid-cols-2">
              {/* Free */}
              <div className="bg-card flex flex-col rounded-2xl border p-8">
                <div className="flex-1">
                  <h3 className="text-card-foreground text-2xl font-bold">Free</h3>
                  <p className="text-muted-foreground mt-2 text-sm">For individual developers</p>
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
                      <span className="text-card-foreground text-sm">Everything in Free</span>
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
                      <span className="text-card-foreground text-sm">Dedicated support & SLA</span>
                    </li>
                  </ul>
                </div>
                <Button asChild variant="outline" size="lg" className="mt-8 w-full">
                  <Link href="/contact">Contact Sales</Link>
                </Button>
              </div>
            </div>
          </div>

          {/* Cloud Section */}
          <div className="mt-24">
            <h2 className="text-foreground mb-4 text-center text-2xl font-bold">Cloud</h2>
            <p className="text-muted-foreground mx-auto mb-8 max-w-2xl text-center">
              Fully managed infrastructure - base fee + per-user pricing
            </p>

            {/* Control Plane Card */}
            <div className="mx-auto mb-12 max-w-2xl">
              <div className="bg-card border-primary/50 flex flex-col rounded-2xl border-2 p-8">
                <div className="flex items-center gap-4">
                  <div className="bg-primary/10 rounded-full p-3">
                    <Server className="text-primary h-6 w-6" />
                  </div>
                  <div className="flex-1">
                    <h3 className="text-card-foreground text-2xl font-bold">Control Plane</h3>
                    <p className="text-muted-foreground text-sm">
                      Base fee for your team&apos;s managed infrastructure
                    </p>
                  </div>
                  <div className="text-right">
                    <span className="text-card-foreground text-4xl font-bold">$29</span>
                    <span className="text-muted-foreground text-lg">/month</span>
                  </div>
                </div>
                <div className="mt-6 flex flex-wrap gap-4">
                  <div className="text-muted-foreground flex items-center gap-2 text-sm">
                    <Check className="text-success h-4 w-4" />
                    Team dashboard
                  </div>
                  <div className="text-muted-foreground flex items-center gap-2 text-sm">
                    <Check className="text-success h-4 w-4" />
                    User management
                  </div>
                  <div className="text-muted-foreground flex items-center gap-2 text-sm">
                    <Check className="text-success h-4 w-4" />
                    Usage analytics
                  </div>
                  <div className="text-muted-foreground flex items-center gap-2 text-sm">
                    <Check className="text-success h-4 w-4" />
                    Billing management
                  </div>
                </div>
              </div>
            </div>

            {/* Per-User Tiers Header */}
            <div className="mb-8 flex items-center justify-center gap-3">
              <Users className="text-muted-foreground h-5 w-5" />
              <h3 className="text-foreground text-xl font-semibold">
                Add users at any tier
              </h3>
            </div>

            {/* User Tier Cards */}
            <div className="grid gap-8 lg:grid-cols-3">
              {/* Tier 1 */}
              <div className="bg-card flex flex-col rounded-2xl border p-8">
                <div className="flex-1">
                  <h3 className="text-card-foreground text-2xl font-bold">Tier 1</h3>
                  <p className="text-muted-foreground mt-2 text-sm">For standard workloads</p>
                  <div className="mt-6">
                    <span className="text-card-foreground text-5xl font-bold">$49</span>
                    <span className="text-muted-foreground text-lg">/user/month</span>
                  </div>
                  <ul className="mt-8 space-y-4">
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">4 OCPU, 32GB RAM</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">160 hours/month included</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">200GB storage</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">Auto-suspend: 30 minutes</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">All core features</span>
                    </li>
                  </ul>
                </div>
                <Button size="lg" className="mt-8 w-full" disabled>
                  Coming Soon
                </Button>
              </div>

              {/* Tier 2 */}
              <div className="bg-card border-primary relative flex flex-col rounded-2xl border-2 p-8">
                <div className="bg-primary text-primary-foreground absolute -top-3 left-1/2 -translate-x-1/2 rounded-full px-4 py-1 text-sm font-medium">
                  Most Popular
                </div>
                <div className="flex-1">
                  <h3 className="text-card-foreground text-2xl font-bold">Tier 2</h3>
                  <p className="text-muted-foreground mt-2 text-sm">For power users</p>
                  <div className="mt-6">
                    <span className="text-card-foreground text-5xl font-bold">$89</span>
                    <span className="text-muted-foreground text-lg">/user/month</span>
                  </div>
                  <ul className="mt-8 space-y-4">
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">8 OCPU, 64GB RAM</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">160 hours/month included</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">500GB storage</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">Auto-suspend: 1 hour</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">Priority support</span>
                    </li>
                  </ul>
                </div>
                <Button size="lg" className="mt-8 w-full" disabled>
                  Coming Soon
                </Button>
              </div>

              {/* Tier 3 */}
              <div className="bg-card flex flex-col rounded-2xl border p-8">
                <div className="flex-1">
                  <h3 className="text-card-foreground text-2xl font-bold">Tier 3</h3>
                  <p className="text-muted-foreground mt-2 text-sm">For heavy workloads</p>
                  <div className="mt-6">
                    <span className="text-card-foreground text-5xl font-bold">$169</span>
                    <span className="text-muted-foreground text-lg">/user/month</span>
                  </div>
                  <ul className="mt-8 space-y-4">
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">16 OCPU, 128GB RAM</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">320 hours/month included</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">1TB storage</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">Auto-suspend: 2 hours</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <Check className="text-success h-5 w-5 flex-shrink-0" />
                      <span className="text-card-foreground text-sm">Dedicated support</span>
                    </li>
                  </ul>
                </div>
                <Button size="lg" className="mt-8 w-full" disabled>
                  Coming Soon
                </Button>
              </div>
            </div>

            {/* Pricing Example */}
            <div className="bg-muted/50 mx-auto mt-12 max-w-xl rounded-xl border p-6">
              <h3 className="text-foreground text-center text-lg font-semibold">
                Example: Team of 3 users
              </h3>
              <div className="mt-4 space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Control Plane</span>
                  <span className="text-foreground">$29</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">2 x Tier 1 users</span>
                  <span className="text-foreground">$98</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">1 x Tier 2 user</span>
                  <span className="text-foreground">$89</span>
                </div>
                <div className="bg-border my-2 h-px" />
                <div className="flex justify-between font-semibold">
                  <span className="text-foreground">Total</span>
                  <span className="text-foreground">$216/month</span>
                </div>
              </div>
            </div>

            {/* Add-ons */}
            <div className="bg-muted/50 mx-auto mt-8 max-w-2xl rounded-xl border p-6">
              <h3 className="text-foreground text-center text-lg font-semibold">Add-ons</h3>
              <div className="mt-4 flex flex-col items-center justify-center gap-6 sm:flex-row">
                <div className="text-center">
                  <span className="text-foreground text-2xl font-bold">$0.20</span>
                  <span className="text-muted-foreground">/hour</span>
                  <p className="text-muted-foreground mt-1 text-sm">Extra usage hours</p>
                </div>
                <div className="bg-border hidden h-12 w-px sm:block" />
                <div className="text-center">
                  <span className="text-foreground text-2xl font-bold">$5</span>
                  <span className="text-muted-foreground">/100GB</span>
                  <p className="text-muted-foreground mt-1 text-sm">Extra storage</p>
                </div>
              </div>
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
                  What&apos;s the difference between BYOC and Cloud?
                </h3>
                <p className="text-muted-foreground mt-2">
                  BYOC (Bring Your Own Compute) lets you deploy Kloudlite on your own AWS, Azure, or
                  GCP infrastructure. You manage the infrastructure and pay your cloud provider
                  directly. Cloud is our fully managed offering where we handle all infrastructure
                  for you.
                </p>
              </div>
              <div>
                <h3 className="text-foreground text-lg font-semibold">
                  How does per-user pricing work?
                </h3>
                <p className="text-muted-foreground mt-2">
                  With Cloud pricing, you pay a base fee of $29/month for the control plane, then
                  add each user at their chosen tier. Users can be on different tiers - for example,
                  some developers might need Tier 1 while others need Tier 3 for heavier workloads.
                  You can change a user&apos;s tier at any time.
                </p>
              </div>
              <div>
                <h3 className="text-foreground text-lg font-semibold">
                  What are OCPU and how do they compare to vCPUs?
                </h3>
                <p className="text-muted-foreground mt-2">
                  OCPU (Oracle CPU) represents a physical CPU core with hyper-threading. 1 OCPU is
                  roughly equivalent to 2 vCPUs. So Tier 1 with 4 OCPU is comparable to 8 vCPUs.
                </p>
              </div>
              <div>
                <h3 className="text-foreground text-lg font-semibold">
                  What happens when a user exceeds their included hours?
                </h3>
                <p className="text-muted-foreground mt-2">
                  Any usage beyond the included hours is billed at $0.20 per hour. You can monitor
                  usage in the dashboard. Auto-suspend helps save hours by automatically stopping
                  idle workspaces.
                </p>
              </div>
              <div>
                <h3 className="text-foreground text-lg font-semibold">How does auto-suspend work?</h3>
                <p className="text-muted-foreground mt-2">
                  Auto-suspend automatically pauses a workspace after a period of inactivity, saving
                  usage hours. Tier 1 suspends after 30 minutes, Tier 2 after 1 hour, and Tier 3
                  after 2 hours of inactivity.
                </p>
              </div>
              <div>
                <h3 className="text-foreground text-lg font-semibold">
                  Can users change their tier?
                </h3>
                <p className="text-muted-foreground mt-2">
                  Yes, you can change any user&apos;s tier at any time. Upgrades take effect
                  immediately, and downgrades take effect at the start of the next billing cycle.
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
  return <PricingPage />
}
