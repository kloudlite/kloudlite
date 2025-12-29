import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { ArrowRight } from 'lucide-react'
import Link from 'next/link'
import { GetStartedButton } from '@/components/get-started-button'

// Cross marker component
function CrossMarker({ className }: { className?: string }) {
  return (
    <div className={cn('absolute', className)}>
      <div className="absolute left-1/2 top-0 -translate-x-1/2 w-px h-5 bg-foreground/20" />
      <div className="absolute top-1/2 left-0 -translate-y-1/2 h-px w-5 bg-foreground/20" />
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
        <WebsiteHeader />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-28">
                <div className="text-center">
                  <p className="text-primary text-xs font-semibold uppercase tracking-wider mb-6">Infrastructure</p>
                  <h1 className="text-[2.5rem] font-bold leading-[1.08] tracking-[-0.035em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground">Workmachines</span>
                  </h1>
                  <p className="text-foreground/55 mx-auto mt-6 max-w-lg text-lg leading-relaxed">
                    Your secure gateway to workspaces and environments.
                  </p>
                  <div className="mt-10 flex items-center justify-center gap-4">
                    <GetStartedButton size="lg" />
                    <Button variant="outline" size="lg" className="rounded-none" asChild>
                      <Link href="/docs/workmachines">
                        Documentation
                      </Link>
                    </Button>
                  </div>
                </div>
              </div>

              {/* Specs */}
              <div className="grid grid-cols-3 -mx-6 lg:-mx-12 border-t border-b border-foreground/10">
                <div className="p-6 lg:p-8 border-r border-foreground/10 text-center">
                  <p className="text-foreground text-2xl lg:text-3xl font-bold tracking-tight font-mono">16</p>
                  <p className="text-foreground/40 mt-1 text-xs">vCPUs Max</p>
                </div>
                <div className="p-6 lg:p-8 border-r border-foreground/10 text-center">
                  <p className="text-foreground text-2xl lg:text-3xl font-bold tracking-tight font-mono">64GB</p>
                  <p className="text-foreground/40 mt-1 text-xs">Memory Max</p>
                </div>
                <div className="p-6 lg:p-8 text-center">
                  <p className="text-foreground text-2xl lg:text-3xl font-bold tracking-tight font-mono">500GB</p>
                  <p className="text-foreground/40 mt-1 text-xs">Storage Max</p>
                </div>
              </div>

              {/* Core Capabilities */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">01</p>
                  <h3 className="text-foreground mt-3 text-lg font-bold tracking-[-0.02em]">VPN Gateway</h3>
                  <p className="text-foreground/50 mt-2 text-sm leading-relaxed">
                    Secure, encrypted access to your workspaces and environment services from anywhere.
                  </p>
                </div>
                <div className="p-8 lg:p-10">
                  <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">02</p>
                  <h3 className="text-foreground mt-3 text-lg font-bold tracking-[-0.02em]">Compute & Storage</h3>
                  <p className="text-foreground/50 mt-2 text-sm leading-relaxed">
                    Dedicated resources for your workloads. Persistent storage that survives restarts.
                  </p>
                </div>
              </div>

              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">03</p>
                  <h3 className="text-foreground mt-3 text-lg font-bold tracking-[-0.02em]">Auto Stop</h3>
                  <p className="text-foreground/50 mt-2 text-sm leading-relaxed">
                    Idle machines stop automatically. Resume in seconds when you need them.
                  </p>
                </div>
                <div className="p-8 lg:p-10">
                  <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">04</p>
                  <h3 className="text-foreground mt-3 text-lg font-bold tracking-[-0.02em]">Multiple Tiers</h3>
                  <p className="text-foreground/50 mt-2 text-sm leading-relaxed">
                    Choose the right size for your workload. Scale up or down as needed.
                  </p>
                </div>
              </div>

              {/* CTA Section */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12">
                <Link
                  href="/docs/workmachines"
                  className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Learn More</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">Read the Docs</h3>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/20 group-hover:text-foreground/40 group-hover:translate-x-1 transition-all" />
                  </div>
                </Link>

                <Link
                  href="/pricing"
                  className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/40 text-xs font-semibold uppercase tracking-wider">Pricing</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">View Plans</h3>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/20 group-hover:text-foreground/40 group-hover:translate-x-1 transition-all" />
                  </div>
                </Link>
              </div>
            </GridContainer>
          </div>

          <WebsiteFooter />
        </main>
      </ScrollArea>
    </div>
  )
}
