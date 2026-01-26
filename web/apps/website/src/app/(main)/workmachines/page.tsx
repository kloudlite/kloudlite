import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { Shield, Server, Power, Layers, Cpu, Lock, Gauge } from 'lucide-react'
import Link from 'next/link'
import { GetStartedButton } from '@/components/get-started-button'
import { PageHeroTitle } from '@/components/page-hero-title'

// Cross marker component
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20 animate-pulse" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20 animate-pulse" />
    </div>
  )
}

// Feature card components
function FeatureCardContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn("group relative p-8 lg:p-12 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors overflow-hidden", className)}>
      <div className="absolute left-0 top-0 w-[3px] h-full bg-primary scale-y-0 group-hover:scale-y-100 transition-transform duration-300 origin-top" />
      {children}
    </div>
  )
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode; title: string; description: string }) {
  return (
    <div className="cursor-default">
      <div className="text-muted-foreground mb-4 transition-colors group-hover:text-primary">{icon}</div>
      <h3 className="text-foreground text-lg font-bold">{title}</h3>
      <p className="text-muted-foreground mt-3 text-base leading-relaxed font-medium transition-colors group-hover:text-foreground">{description}</p>
    </div>
  )
}

function GridContainer({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('relative mx-auto max-w-5xl', className)}>
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />
        <CrossMarker className="top-0 left-0 -translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="top-0 right-0 translate-x-1/2 -translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 left-0 -translate-x-1/2 translate-y-1/2 w-5 h-5" />
        <CrossMarker className="bottom-0 right-0 translate-x-1/2 translate-y-1/2 w-5 h-5" />
      </div>
      <div className="relative">{children}</div>
    </div>
  )
}

export default function WorkmachinesPage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="workmachines" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-28">
                <div className="text-center">
                  <PageHeroTitle accentedWord="Workmachines.">
                    Development Infrastructure
                  </PageHeroTitle>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-lg lg:text-xl leading-relaxed">
                    Secure VPN gateway and dedicated compute for your workspaces. Your personal development server in the cloud.
                  </p>
                  <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
                    <GetStartedButton size="lg" className="w-full sm:w-auto" />
                    <Button variant="outline" size="lg" className="w-full sm:w-auto rounded-none" asChild>
                      <Link href="/docs/workmachines">
                        Documentation
                      </Link>
                    </Button>
                  </div>
                </div>
              </div>

              <div className="grid sm:grid-cols-2 border-t border-foreground/10 -mx-6 lg:-mx-12">

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Specs Section - Full Width */}
                <div className="sm:col-span-2 border-b border-foreground/10 bg-foreground/[0.015]">
                  <div className="grid grid-cols-3">
                    <div className="p-8 lg:p-12 border-r border-foreground/10 text-center group hover:bg-foreground/[0.03] transition-colors">
                      <p className="text-foreground text-3xl lg:text-4xl font-mono font-semibold tracking-tight">16</p>
                      <p className="text-muted-foreground mt-2 text-sm group-hover:text-foreground transition-colors">vCPUs Max</p>
                    </div>
                    <div className="p-8 lg:p-12 border-r border-foreground/10 text-center group hover:bg-foreground/[0.03] transition-colors">
                      <p className="text-foreground text-3xl lg:text-4xl font-mono font-semibold tracking-tight">64GB</p>
                      <p className="text-muted-foreground mt-2 text-sm group-hover:text-foreground transition-colors">Memory Max</p>
                    </div>
                    <div className="p-8 lg:p-12 text-center group hover:bg-foreground/[0.03] transition-colors">
                      <p className="text-foreground text-3xl lg:text-4xl font-mono font-semibold tracking-tight">500GB</p>
                      <p className="text-muted-foreground mt-2 text-sm group-hover:text-foreground transition-colors">Storage Max</p>
                    </div>
                  </div>
                </div>

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Section Header */}
                <div className="sm:col-span-2 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                  <h2 className="text-foreground text-4xl lg:text-5xl font-bold tracking-tight">
                    Dedicated infrastructure for <span className="relative inline-block">
                      <span className="relative z-10">your development.</span>
                      <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                    </span>
                  </h2>
                  <p className="text-muted-foreground mt-4 text-base lg:text-lg max-w-2xl">
                    Each workmachine is your personal development server with VPN access, persistent storage, and dedicated compute.
                  </p>
                </div>

                {/* Feature 1 */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Shield className="h-5 w-5" />}
                    title="VPN Gateway"
                    description="Secure, encrypted access to your workspaces and environment services from anywhere. No public endpoints needed."
                  />
                </FeatureCardContainer>

                {/* Feature 2 */}
                <FeatureCardContainer className="border-b border-foreground/10">
                  <FeatureCard
                    icon={<Server className="h-5 w-5" />}
                    title="Compute & Storage"
                    description="Dedicated CPU and memory for your workloads. Persistent storage that survives restarts and keeps your data safe."
                  />
                </FeatureCardContainer>

                {/* Cross Marker - After Row 1 */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Feature 3 */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Power className="h-5 w-5" />}
                    title="Auto Stop"
                    description="Idle machines stop automatically to save resources. Resume in seconds when you need them—no cold starts."
                  />
                </FeatureCardContainer>

                {/* Feature 4 */}
                <FeatureCardContainer className="border-b border-foreground/10">
                  <FeatureCard
                    icon={<Layers className="h-5 w-5" />}
                    title="Flexible Tiers"
                    description="Choose from multiple machine sizes. Start small and scale up as your needs grow—from 1 vCPU to 16 vCPUs."
                  />
                </FeatureCardContainer>

                {/* Cross Marker - After Row 2 */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Feature 5 */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Cpu className="h-5 w-5" />}
                    title="GPU Enabled"
                    description="Accelerate AI/ML workloads with GPU-enabled nodes. Perfect for training models, running inference, and data processing."
                  />
                </FeatureCardContainer>

                {/* Feature 6 */}
                <FeatureCardContainer className="border-b border-foreground/10">
                  <FeatureCard
                    icon={<Lock className="h-5 w-5" />}
                    title="Network Isolation"
                    description="Complete network isolation between workspaces. Private networking ensures your data stays secure and separate."
                  />
                </FeatureCardContainer>

                {/* Cross Marker - After Row 3 */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-2/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Feature 7 */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Gauge className="h-5 w-5" />}
                    title="Performance Monitoring"
                    description="Real-time metrics and monitoring for CPU, memory, and network usage. Track your machine performance at a glance."
                  />
                </FeatureCardContainer>

                {/* Feature 8 */}
                <FeatureCardContainer className="border-b border-foreground/10">
                  <FeatureCard
                    icon={<Server className="h-5 w-5" />}
                    title="High Availability"
                    description="Built on reliable infrastructure with automatic failover. Your machines stay available when you need them most."
                  />
                </FeatureCardContainer>

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* CTA Section */}
                <div className="p-8 lg:p-16 border-b border-foreground/10 sm:col-span-2 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 bg-foreground/[0.015]">
                  <div>
                    <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                      Ready to get started?
                    </h2>
                    <p className="text-muted-foreground mt-2 text-base lg:text-lg">
                      Create your first workmachine in minutes.
                    </p>
                  </div>
                  <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                    <GetStartedButton size="lg" className="w-full sm:w-auto" />
                    <Button
                      asChild
                      variant="outline"
                      size="lg"
                      className="w-full sm:w-auto rounded-none"
                    >
                      <Link href="/pricing">
                        View Pricing
                      </Link>
                    </Button>
                  </div>
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
