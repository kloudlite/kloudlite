'use client'

import { useState, useEffect, useRef } from 'react'
import { Button, ScrollArea } from '@kloudlite/ui'
import Link from 'next/link'
import { ArrowRight, Check } from 'lucide-react'
import { GetStartedButton } from '@/components/get-started-button'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { PageHeroTitle } from '@/components/page-hero-title'
import { cn } from '@kloudlite/lib'
// Animation handled with CSS transitions

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
      <style jsx>{`
        @keyframes pulseTopLeftToRight {
          0% { left: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { left: 100%; opacity: 0; }
        }
        @keyframes pulseRightTopToBottom {
          0% { top: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { top: 100%; opacity: 0; }
        }
        @keyframes pulseBottomRightToLeft {
          0% { right: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { right: 100%; opacity: 0; }
        }
        @keyframes pulseLeftBottomToTop {
          0% { bottom: 0%; opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { bottom: 100%; opacity: 0; }
        }
        .pulse-top {
          animation: pulseTopLeftToRight 4s ease-in-out infinite;
        }
        .pulse-right {
          animation: pulseRightTopToBottom 4s ease-in-out infinite 1s;
        }
        .pulse-bottom {
          animation: pulseBottomRightToLeft 4s ease-in-out infinite 2s;
        }
        .pulse-left {
          animation: pulseLeftBottomToTop 4s ease-in-out infinite 3s;
        }
      `}</style>
      {/* Grid lines */}
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        {/* Vertical lines */}
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />

        {/* Horizontal lines */}
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />

        {/* Animated pulses */}
        <div className="pulse-top absolute top-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
        <div className="pulse-right absolute right-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
        <div className="pulse-bottom absolute bottom-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
        <div className="pulse-left absolute left-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />

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
  const byocRef = useRef<HTMLButtonElement>(null)
  const cloudRef = useRef<HTMLButtonElement>(null)
  const [underlineStyle, setUnderlineStyle] = useState({ left: 0, width: 0 })

  // Update underline position
  useEffect(() => {
    const updatePosition = () => {
      const activeRef = activeTab === 'byoc' ? byocRef : cloudRef
      if (activeRef.current) {
        const fullWidth = activeRef.current.offsetWidth
        const underlineWidth = fullWidth * 0.6 // 60% of button width
        const leftOffset = activeRef.current.offsetLeft + (fullWidth - underlineWidth) / 2

        setUnderlineStyle({
          left: leftOffset,
          width: underlineWidth
        })
      }
    }

    // Small delay to ensure layout is ready
    setTimeout(updatePosition, 10)

    window.addEventListener('resize', updatePosition)
    return () => window.removeEventListener('resize', updatePosition)
  }, [activeTab])

  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="pricing" alwaysShowBorder />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="pt-20 pb-8 lg:pt-28 lg:pb-12">
                <div className="text-center">
                  <PageHeroTitle accentedWord="pricing.">
                    Simple, transparent
                  </PageHeroTitle>

                  <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-lg leading-relaxed">
                    Start free with your own infrastructure, or let us manage everything for you.
                  </p>

                  {/* Tab Switcher */}
                  <div className="mt-10 inline-flex gap-1 relative">
                    <button
                      ref={byocRef}
                      onClick={() => setActiveTab('byoc')}
                      className={cn(
                        'relative px-6 py-2.5 text-base font-medium transition-all duration-200 cursor-pointer',
                        'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
                        activeTab === 'byoc'
                          ? 'text-foreground'
                          : 'text-muted-foreground hover:text-foreground'
                      )}
                    >
                      Self-Hosted
                    </button>
                    <button
                      ref={cloudRef}
                      onClick={() => setActiveTab('cloud')}
                      className={cn(
                        'relative px-6 py-2.5 text-base font-medium transition-all duration-200 cursor-pointer',
                        'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
                        activeTab === 'cloud'
                          ? 'text-foreground'
                          : 'text-muted-foreground hover:text-foreground'
                      )}
                    >
                      Cloud
                    </button>
                    {/* Animated underline with CSS transition */}
                    {underlineStyle.width > 0 && (
                      <div
                        className="absolute bottom-1 h-[2px] bg-primary transition-all duration-300 ease-out"
                        style={{
                          left: `${underlineStyle.left}px`,
                          width: `${underlineStyle.width}px`,
                        }}
                      />
                    )}
                  </div>
                </div>
              </div>

              {/* Divider */}
              <div className="h-px bg-foreground/10 -mx-6 lg:-mx-12" />

              {/* Pricing Grid */}
              <div className="-mx-6 lg:-mx-12">
                <div className={cn(
                  'transition-opacity duration-300',
                  activeTab === 'byoc' ? 'opacity-100' : 'opacity-0 hidden'
                )}>
                  <BYOCPricing />
                </div>
                <div className={cn(
                  'transition-opacity duration-300',
                  activeTab === 'cloud' ? 'opacity-100' : 'opacity-0 hidden'
                )}>
                  <CloudPricing />
                </div>
              </div>

              {/* FAQ Section */}
              <div className="grid lg:grid-cols-3 border-t border-foreground/10 -mx-6 lg:-mx-12">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-center">
                  <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em] sm:text-3xl">
                    Questions
                  </h2>
                  <p className="text-muted-foreground mt-3 text-base">
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
      <div className="p-8 lg:p-12 border-b border-foreground/10 lg:border-r group transition-colors hover:bg-foreground/[0.02] bg-foreground/[0.01]">
        <div className="inline-block px-3 py-1 bg-primary/10 border border-primary/20 rounded-sm">
          <p className="text-primary text-xs font-semibold uppercase tracking-wider">
            Open Source
          </p>
        </div>
        <h3 className="text-foreground mt-6 text-3xl font-bold tracking-[-0.02em]">Free</h3>
        <p className="text-muted-foreground mt-3 text-base leading-relaxed">
          For individuals and small teams
        </p>
        <div className="mt-8">
          <span className="text-foreground text-6xl font-bold tracking-tight">$0</span>
          <span className="text-muted-foreground ml-2 text-lg">forever</span>
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
      <div className="p-8 lg:p-12 border-b border-foreground/10 group transition-colors hover:bg-foreground/[0.02] bg-foreground/[0.01]">
        <div className="inline-block px-3 py-1 bg-foreground/5 border border-foreground/10 rounded-sm">
          <p className="text-muted-foreground text-xs font-semibold uppercase tracking-wider">
            Custom
          </p>
        </div>
        <h3 className="text-foreground mt-6 text-3xl font-bold tracking-[-0.02em]">Enterprise</h3>
        <p className="text-muted-foreground mt-3 text-base leading-relaxed">
          For organizations with advanced needs
        </p>
        <div className="mt-8">
          <span className="text-foreground text-6xl font-bold tracking-tight">Custom</span>
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
  const [tier1Users, setTier1Users] = useState(2)
  const [tier2Users, setTier2Users] = useState(3)
  const [tier3Users, setTier3Users] = useState(1)
  const [tier1ExtraHrs, setTier1ExtraHrs] = useState(0)
  const [tier2ExtraHrs, setTier2ExtraHrs] = useState(20)
  const [tier3ExtraHrs, setTier3ExtraHrs] = useState(0)

  const baseFee = 29
  const tier1Price = 29
  const tier2Price = 49
  const tier3Price = 89
  const tier1HrRate = 0.18
  const tier2HrRate = 0.30
  const tier3HrRate = 0.55

  const tier1Total = tier1Users * tier1Price
  const tier2Total = tier2Users * tier2Price
  const tier3Total = tier3Users * tier3Price
  const tier1ExtraCost = tier1ExtraHrs * tier1HrRate
  const tier2ExtraCost = tier2ExtraHrs * tier2HrRate
  const tier3ExtraCost = tier3ExtraHrs * tier3HrRate
  const totalExtraHrs = tier1ExtraCost + tier2ExtraCost + tier3ExtraCost
  const totalUsers = tier1Users + tier2Users + tier3Users
  const total = baseFee + tier1Total + tier2Total + tier3Total + totalExtraHrs

  return (
    <div>
      {/* Control Plane Row */}
      <div className="grid lg:grid-cols-3 border-b border-foreground/10 bg-foreground/[0.01]">
        <div className="p-8 lg:p-12 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-center">
          <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em] sm:text-3xl">
            Cloud Pricing
          </h2>
          <p className="text-muted-foreground mt-3 text-base">
            Fully managed by us.
          </p>
        </div>
        <div className="lg:col-span-2 p-8 lg:p-12 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6">
          <div>
            <div className="inline-block px-3 py-1 bg-foreground/5 border border-foreground/10 rounded-sm mb-4">
              <p className="text-muted-foreground text-xs font-semibold uppercase tracking-wider">
                Base Fee
              </p>
            </div>
            <h3 className="text-foreground text-xl font-bold tracking-[-0.02em]">Control Plane</h3>
            <p className="text-muted-foreground mt-2 text-base">
              Dashboard, user management, billing
            </p>
          </div>
          <div className="text-left sm:text-right">
            <span className="text-foreground text-5xl font-bold tracking-tight">$29</span>
            <span className="text-muted-foreground text-lg">/mo</span>
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
          highlighted
        />
        <Tier
          name="Tier 3"
          description="Power users"
          price={89}
          extraHourlyRate={0.55}
          features={['16 vCPUs, 64GB RAM', '160 hrs/mo', '500GB storage', '1 hr suspend']}
        />
      </div>

      {/* Calculator Row */}
      <div className="grid lg:grid-cols-3 border-b border-foreground/10 bg-foreground/[0.01]">
        <div className="p-8 lg:p-12 border-b lg:border-b-0 lg:border-r border-foreground/10 flex flex-col justify-center">
          <h2 className="text-foreground text-2xl font-bold tracking-[-0.02em]">
            Cost Estimate
          </h2>
          <p className="text-muted-foreground mt-3 text-base">
            Calculate your monthly cost.
          </p>
        </div>
        <div className="lg:col-span-2 p-8 lg:p-12">
          {/* Users Row */}
          <div className="grid sm:grid-cols-3 gap-6">
            <div>
              <label className="text-muted-foreground text-sm font-medium uppercase tracking-wider">Tier 1 Users</label>
              <div className="mt-2 flex items-center gap-2">
                <button
                  onClick={() => setTier1Users(Math.max(0, tier1Users - 1))}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  −
                </button>
                <span className="text-foreground text-lg font-semibold w-8 text-center">{tier1Users}</span>
                <button
                  onClick={() => setTier1Users(tier1Users + 1)}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  +
                </button>
              </div>
              <p className="text-muted-foreground text-xs mt-2">${tier1Price}/user = ${tier1Total}</p>
            </div>
            <div>
              <label className="text-muted-foreground text-sm font-medium uppercase tracking-wider">Tier 2 Users</label>
              <div className="mt-2 flex items-center gap-2">
                <button
                  onClick={() => setTier2Users(Math.max(0, tier2Users - 1))}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  −
                </button>
                <span className="text-foreground text-lg font-semibold w-8 text-center">{tier2Users}</span>
                <button
                  onClick={() => setTier2Users(tier2Users + 1)}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  +
                </button>
              </div>
              <p className="text-muted-foreground text-xs mt-2">${tier2Price}/user = ${tier2Total}</p>
            </div>
            <div>
              <label className="text-muted-foreground text-sm font-medium uppercase tracking-wider">Tier 3 Users</label>
              <div className="mt-2 flex items-center gap-2">
                <button
                  onClick={() => setTier3Users(Math.max(0, tier3Users - 1))}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  −
                </button>
                <span className="text-foreground text-lg font-semibold w-8 text-center">{tier3Users}</span>
                <button
                  onClick={() => setTier3Users(tier3Users + 1)}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  +
                </button>
              </div>
              <p className="text-muted-foreground text-xs mt-2">${tier3Price}/user = ${tier3Total}</p>
            </div>
          </div>

          {/* Extra Hours Row */}
          <div className="grid sm:grid-cols-3 gap-6 mt-6 pt-6 border-t border-foreground/10">
            <div>
              <label className="text-muted-foreground text-sm font-medium uppercase tracking-wider">Extra Hrs (T1)</label>
              <div className="mt-2 flex items-center gap-2">
                <button
                  onClick={() => setTier1ExtraHrs(Math.max(0, tier1ExtraHrs - 10))}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  −
                </button>
                <span className="text-foreground text-lg font-semibold w-8 text-center">{tier1ExtraHrs}</span>
                <button
                  onClick={() => setTier1ExtraHrs(tier1ExtraHrs + 10)}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  +
                </button>
              </div>
              <p className="text-muted-foreground text-xs mt-2">${tier1HrRate}/hr = ${tier1ExtraCost.toFixed(2)}</p>
            </div>
            <div>
              <label className="text-muted-foreground text-sm font-medium uppercase tracking-wider">Extra Hrs (T2)</label>
              <div className="mt-2 flex items-center gap-2">
                <button
                  onClick={() => setTier2ExtraHrs(Math.max(0, tier2ExtraHrs - 10))}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  −
                </button>
                <span className="text-foreground text-lg font-semibold w-8 text-center">{tier2ExtraHrs}</span>
                <button
                  onClick={() => setTier2ExtraHrs(tier2ExtraHrs + 10)}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  +
                </button>
              </div>
              <p className="text-muted-foreground text-xs mt-2">${tier2HrRate.toFixed(2)}/hr = ${tier2ExtraCost.toFixed(2)}</p>
            </div>
            <div>
              <label className="text-muted-foreground text-sm font-medium uppercase tracking-wider">Extra Hrs (T3)</label>
              <div className="mt-2 flex items-center gap-2">
                <button
                  onClick={() => setTier3ExtraHrs(Math.max(0, tier3ExtraHrs - 10))}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  −
                </button>
                <span className="text-foreground text-lg font-semibold w-8 text-center">{tier3ExtraHrs}</span>
                <button
                  onClick={() => setTier3ExtraHrs(tier3ExtraHrs + 10)}
                  className="w-9 h-9 flex items-center justify-center border border-foreground/10 bg-foreground/[0.02] text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground transition-colors font-medium"
                >
                  +
                </button>
              </div>
              <p className="text-muted-foreground text-xs mt-2">${tier3HrRate}/hr = ${tier3ExtraCost.toFixed(2)}</p>
            </div>
          </div>

          {/* Total */}
          <div className="mt-6 pt-6 border-t border-foreground/10 flex items-end justify-between">
            <div className="text-muted-foreground text-base space-y-1">
              <p>{totalUsers} user{totalUsers !== 1 ? 's' : ''}: ${tier1Total + tier2Total + tier3Total}</p>
              {totalExtraHrs > 0 && <p>Extra hours: ${totalExtraHrs.toFixed(2)}</p>}
              <p>Base fee: ${baseFee}</p>
            </div>
            <div className="text-right">
              <p className="text-muted-foreground text-xs font-medium uppercase tracking-wider">Estimated Total</p>
              <p className="text-foreground text-3xl font-bold tracking-tight mt-1">
                ${total.toFixed(0)}<span className="text-muted-foreground text-sm font-normal">/mo</span>
              </p>
            </div>
          </div>
        </div>
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
  className,
  highlighted = false
}: {
  name: string
  description: string
  price: number
  extraHourlyRate?: number
  features: string[]
  className?: string
  highlighted?: boolean
}) {
  return (
    <div className={cn('p-8 lg:p-10 group cursor-default transition-colors', highlighted ? 'bg-primary/[0.02] hover:bg-primary/[0.03]' : 'hover:bg-foreground/[0.02]', className)}>
      {highlighted && (
        <div className="inline-block px-3 py-1 bg-primary/10 border border-primary/20 rounded-sm mb-4">
          <p className="text-primary text-xs font-semibold uppercase tracking-wider">
            Most Popular
          </p>
        </div>
      )}
      <h4 className="text-foreground text-lg font-bold tracking-[-0.02em]">{name}</h4>
      <p className="text-muted-foreground mt-2 text-base transition-colors group-hover:text-foreground">{description}</p>
      <div className="mt-6">
        <span className="text-foreground text-3xl font-bold tracking-tight">${price}</span>
        <span className="text-muted-foreground text-sm">/user/mo</span>
      </div>
      <ul className="mt-6 space-y-3">
        {features.map((f, i) => (
          <li key={i} className="text-muted-foreground text-base transition-colors group-hover:text-foreground">{f}</li>
        ))}
      </ul>
      {extraHourlyRate && (
        <div className="mt-6 pt-6 border-t border-foreground/10">
          <p className="text-muted-foreground text-xs font-semibold uppercase tracking-wider">Extra Hours</p>
          <div className="mt-2">
            <span className="text-foreground text-lg font-bold">${extraHourlyRate.toFixed(2)}</span>
            <span className="text-muted-foreground text-sm">/hr</span>
          </div>
        </div>
      )}
    </div>
  )
}

function Li({ children }: { children: React.ReactNode }) {
  return (
    <li className="flex items-start gap-3 text-muted-foreground text-base leading-relaxed transition-colors group-hover:text-foreground">
      <Check className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0 transition-colors group-hover:text-foreground" />
      <span>{children}</span>
    </li>
  )
}

function FAQ({ question, answer }: { question: string; answer: string }) {
  return (
    <div className="p-8 lg:p-10 group cursor-default transition-colors hover:bg-foreground/[0.02]">
      <h3 className="text-foreground text-base font-semibold">{question}</h3>
      <p className="text-muted-foreground mt-3 text-base leading-relaxed transition-colors group-hover:text-foreground">{answer}</p>
    </div>
  )
}

export default function Page() {
  return <PricingPage />
}
