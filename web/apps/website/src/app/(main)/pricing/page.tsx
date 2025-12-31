'use client'

import { useState } from 'react'
import { Button, ScrollArea } from '@kloudlite/ui'
import Link from 'next/link'
import { ArrowRight, Check } from 'lucide-react'
import { GetStartedButton } from '@/components/get-started-button'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'

// Cross marker component
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      {/* Vertical line */}
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      {/* Horizontal line */}
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
    </div>
  )
}

// Grid container like Vercel
function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      {/* Grid lines */}
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        {/* Vertical lines */}
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />

        {/* Horizontal lines */}
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />

        {/* Corner markers */}
        <CrossMarker className="top-0 left-0 -translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="top-0 right-0 translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 left-0 -translate-x-1/2 translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 right-0 translate-x-1/2 translate-y-1/2 w-5 h-5" />
      </div>

      {/* Content */}
      <div className="relative">
        {children}
      </div>
    </div>
  )
}

function PricingPage() {
  const [activeTab, setActiveTab] = useState<'byoc' | 'cloud'>('byoc')

  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="pricing" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-24">
                <div className="text-center">
                  <h1 className="text-[2.5rem] font-bold leading-[1.08] tracking-[-0.035em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground/40">S</span><span className="text-foreground">imple</span>{' '}
                    <span className="text-foreground/40">P</span><span className="text-foreground">ricing</span>
                    <br />
                    <span className="text-foreground/40">for every team.</span>
                  </h1>

                  <p className="text-foreground/55 mx-auto mt-6 max-w-md text-lg leading-relaxed">
                    Start free with your own infrastructure,
                    <br />
                    or let us manage everything for you.
                  </p>

                  {/* Tab Switcher */}
                  <div className="mt-10 inline-flex p-1 bg-foreground/5 border border-foreground/10">
                    <button
                      onClick={() => setActiveTab('byoc')}
                      className={cn(
                        'relative px-6 py-2.5 text-sm font-medium transition-all duration-200',
                        activeTab === 'byoc'
                          ? 'bg-background text-foreground shadow-sm'
                          : 'text-foreground/50 hover:text-foreground/70'
                      )}
                    >
                      Self-Hosted
                    </button>
                    <button
                      onClick={() => setActiveTab('cloud')}
                      className={cn(
                        'relative px-6 py-2.5 text-sm font-medium transition-all duration-200',
                        activeTab === 'cloud'
                          ? 'bg-background text-foreground shadow-sm'
                          : 'text-foreground/50 hover:text-foreground/70'
                      )}
                    >
                      Cloud
                    </button>
                  </div>
                </div>
              </div>

              {/* Divider */}
              <div className="h-px bg-foreground/10 -mx-6 lg:-mx-12" />

              {/* Pricing Grid */}
              <div className="-mx-6 lg:-mx-12">
                {activeTab === 'byoc' ? <BYOCPricing /> : <CloudPricing />}
              </div>

              {/* FAQ Section */}
              <div className="grid lg:grid-cols-3 border-t border-foreground/10 -mx-6 lg:-mx-12">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-center">
                  <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em] sm:text-3xl">
                    Questions
                  </h2>
                  <p className="text-foreground/50 mt-3 text-base">
                    Common questions answered.
                  </p>
                </div>
                <div className="lg:col-span-2 divide-y divide-foreground/10">
                  <FAQ
                    question="What's the difference between Self-Hosted and Cloud?"
                    answer="Self-Hosted is free and open source - deploy on your own AWS, Azure, or GCP. You control infrastructure and pay your cloud provider directly. Cloud is fully managed by us with a $29/mo base fee plus per-user tier pricing."
                  />
                  <FAQ
                    question="How does per-user pricing work?"
                    answer="Each user picks a tier based on their workload needs. Users can be on different tiers within the same team. Upgrade or downgrade anytime."
                  />
                  <FAQ
                    question="What happens when hours are exceeded?"
                    answer="Extra hours are billed per tier: $0.18/hr for Tier 1, $0.30/hr for Tier 2, $0.55/hr for Tier 3. Auto-suspend helps optimize costs by pausing idle workspaces."
                  />
                  <FAQ
                    question="Can I try before committing?"
                    answer="Yes! Self-Hosted is completely free with all core features. For Cloud, contact us for a trial period to test the fully managed experience."
                  />
                </div>
              </div>
            </GridContainer>
          </div>

          <WebsiteFooter />
        </main>
      </ScrollArea>
    </div>
  )
}

function BYOCPricing() {
  return (
    <div className="grid lg:grid-cols-2">
      {/* Free */}
      <div className="p-8 lg:p-10 border-b border-foreground/10 lg:border-r group transition-colors hover:bg-foreground/[0.02]">
        <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">
          Open Source
        </p>
        <h3 className="text-foreground mt-4 text-2xl font-bold tracking-[-0.02em]">Free</h3>
        <p className="text-foreground/50 mt-3 text-sm leading-relaxed transition-colors group-hover:text-foreground/70">
          For individuals and small teams
        </p>
        <div className="mt-8">
          <span className="text-foreground text-5xl font-bold tracking-tight">$0</span>
          <span className="text-foreground/40 ml-2 text-sm">forever</span>
        </div>
        <ul className="mt-10 space-y-4">
          <Li>Deploy on AWS, Azure, or GCP</Li>
          <Li>Unlimited workspaces</Li>
          <Li>Unlimited environments</Li>
          <Li>All core features</Li>
          <Li>Community support</Li>
        </ul>
        <GetStartedButton size="lg" className="mt-10 w-full" />
      </div>

      {/* Enterprise */}
      <div className="p-8 lg:p-10 border-b border-foreground/10 group transition-colors hover:bg-foreground/[0.02]">
        <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">
          Custom
        </p>
        <h3 className="text-foreground mt-4 text-2xl font-bold tracking-[-0.02em]">Enterprise</h3>
        <p className="text-foreground/50 mt-3 text-sm leading-relaxed transition-colors group-hover:text-foreground/70">
          For organizations with advanced needs
        </p>
        <div className="mt-8">
          <span className="text-foreground text-5xl font-bold tracking-tight">Custom</span>
        </div>
        <ul className="mt-10 space-y-4">
          <Li>Everything in Free</Li>
          <Li>On-premise deployment</Li>
          <Li>Organization management</Li>
          <Li>SSO & advanced auth</Li>
          <Li>Dedicated support & SLA</Li>
        </ul>
        <Button asChild variant="outline" size="lg" className="mt-10 w-full">
          <Link href="/contact">
            Contact Sales
            <ArrowRight className="ml-2 h-4 w-4" />
          </Link>
        </Button>
      </div>
    </div>
  )
}

function CloudPricing() {
  return (
    <div>
      {/* Control Plane Row */}
      <div className="grid lg:grid-cols-3 border-b border-foreground/10">
        <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-center">
          <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em] sm:text-3xl">
            Cloud Pricing
          </h2>
          <p className="text-foreground/50 mt-3 text-base">
            Fully managed by us.
          </p>
        </div>
        <div className="lg:col-span-2 p-8 lg:p-10 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">
              Base Fee
            </p>
            <h3 className="text-foreground mt-2 text-xl font-bold tracking-[-0.02em]">Control Plane</h3>
            <p className="text-foreground/50 mt-1 text-sm">
              Dashboard, user management, billing
            </p>
          </div>
          <div className="text-left sm:text-right">
            <span className="text-foreground text-4xl font-bold tracking-tight">$29</span>
            <span className="text-foreground/40 text-sm">/mo</span>
          </div>
        </div>
      </div>

      {/* Tiers Row */}
      <div className="grid lg:grid-cols-3 border-b border-foreground/10">
        <Tier
          name="Tier 1"
          description="Light workloads"
          price={29}
          extraHourlyRate={0.18}
          features={['8 vCPUs, 16GB RAM', '160 hrs/mo', '100GB storage', '15 min suspend']}
          className="border-b lg:border-b-0 lg:border-r border-foreground/10"
        />
        <Tier
          name="Tier 2"
          description="Standard workloads"
          price={49}
          extraHourlyRate={0.30}
          features={['12 vCPUs, 32GB RAM', '160 hrs/mo', '200GB storage', '30 min suspend']}
          className="border-b lg:border-b-0 lg:border-r border-foreground/10"
        />
        <Tier
          name="Tier 3"
          description="Power users"
          price={89}
          extraHourlyRate={0.55}
          features={['16 vCPUs, 64GB RAM', '160 hrs/mo', '500GB storage', '1 hr suspend']}
        />
      </div>
    </div>
  )
}

function Tier({
  name,
  description,
  price,
  extraHourlyRate,
  features,
  className
}: {
  name: string
  description: string
  price: number
  extraHourlyRate?: number
  features: string[]
  className?: string
}) {
  return (
    <div className={cn('p-8 lg:p-10 group cursor-default transition-colors hover:bg-foreground/[0.02]', className)}>
      <h4 className="text-foreground text-lg font-bold tracking-[-0.02em]">{name}</h4>
      <p className="text-foreground/50 mt-1 text-sm transition-colors group-hover:text-foreground/70">{description}</p>
      <div className="mt-6">
        <span className="text-foreground text-3xl font-bold tracking-tight">${price}</span>
        <span className="text-foreground/40 text-sm">/user/mo</span>
      </div>
      <ul className="mt-6 space-y-3">
        {features.map((f, i) => (
          <li key={i} className="text-foreground/50 text-sm transition-colors group-hover:text-foreground/60">{f}</li>
        ))}
      </ul>
      {extraHourlyRate && (
        <div className="mt-6 pt-6 border-t border-foreground/10">
          <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Extra Hours</p>
          <div className="mt-2">
            <span className="text-foreground text-lg font-bold">${extraHourlyRate.toFixed(2)}</span>
            <span className="text-foreground/40 text-sm">/hr</span>
          </div>
        </div>
      )}
    </div>
  )
}

function Li({ children }: { children: React.ReactNode }) {
  return (
    <li className="flex items-start gap-3 text-foreground/60 text-sm leading-relaxed transition-colors group-hover:text-foreground/80">
      <Check className="h-4 w-4 text-foreground/40 mt-0.5 shrink-0 transition-colors group-hover:text-foreground/60" />
      <span>{children}</span>
    </li>
  )
}

function FAQ({ question, answer }: { question: string; answer: string }) {
  return (
    <div className="p-8 lg:p-10 group cursor-default transition-colors hover:bg-foreground/[0.02]">
      <h3 className="text-foreground text-base font-semibold">{question}</h3>
      <p className="text-foreground/50 mt-3 text-sm leading-relaxed transition-colors group-hover:text-foreground/60">{answer}</p>
    </div>
  )
}

export default function Page() {
  return <PricingPage />
}
