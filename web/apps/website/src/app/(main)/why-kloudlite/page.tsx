import { ScrollArea } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import { ArrowRight } from 'lucide-react'
import Link from 'next/link'
import { GetStartedButton } from '@/components/get-started-button'

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

export default function WhyKloudlitePage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="why-kloudlite" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero */}
              <div className="py-20 lg:py-28">
                <div className="text-center">
                  <h1 className="text-[2.5rem] font-bold leading-[1.08] tracking-[-0.035em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground">Why Kloudlite?</span>
                  </h1>
                  <p className="text-foreground/55 mx-auto mt-6 max-w-md text-lg leading-relaxed">
                    Code against live services without deploying.
                  </p>
                  <div className="mt-10">
                    <GetStartedButton size="lg" />
                  </div>
                </div>
              </div>

              {/* Core Value */}
              <div className="grid lg:grid-cols-3 -mx-6 lg:-mx-12 border-t border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <p className="text-foreground/30 text-4xl font-bold font-mono">01</p>
                  <h3 className="text-foreground mt-4 text-lg font-bold tracking-[-0.02em]">No Local Setup</h3>
                  <p className="text-foreground/70 mt-2 text-sm leading-relaxed">
                    Cloud workspaces with your repo, packages, and tools. Ready in seconds.
                  </p>
                </div>
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <p className="text-foreground/30 text-4xl font-bold font-mono">02</p>
                  <h3 className="text-foreground mt-4 text-lg font-bold tracking-[-0.02em]">Real Services</h3>
                  <p className="text-foreground/70 mt-2 text-sm leading-relaxed">
                    Connect to your running databases, APIs, and queues. No mocks.
                  </p>
                </div>
                <div className="p-8 lg:p-10">
                  <p className="text-foreground/30 text-4xl font-bold font-mono">03</p>
                  <h3 className="text-foreground mt-4 text-lg font-bold tracking-[-0.02em]">Traffic Intercept</h3>
                  <p className="text-foreground/70 mt-2 text-sm leading-relaxed">
                    Route live requests to your workspace. Debug with production data.
                  </p>
                </div>
              </div>

              {/* The Shift */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <p className="text-foreground/60 text-xs font-semibold uppercase tracking-wider">Without Kloudlite</p>
                  <div className="mt-6 space-y-3 font-mono text-sm">
                    <p className="text-foreground/60">git clone → install deps → setup db → configure env → mock services → code → push → wait for CI → deploy → test → find bug → repeat</p>
                  </div>
                  <p className="text-foreground/50 text-xs mt-6 font-mono">Hours to days per feature</p>
                </div>
                <div className="p-8 lg:p-10 bg-foreground/[0.02]">
                  <p className="text-primary text-xs font-semibold uppercase tracking-wider">With Kloudlite</p>
                  <div className="mt-6 space-y-3 font-mono text-sm">
                    <p className="text-foreground/80">open workspace → connect env → code → done</p>
                  </div>
                  <p className="text-primary/80 text-xs mt-6 font-mono">Minutes</p>
                </div>
              </div>

              {/* Technical Highlights */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <p className="text-foreground/60 text-xs font-semibold uppercase tracking-wider">Architecture</p>
                  <ul className="mt-6 space-y-2 text-sm text-foreground/70">
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-full" />
                      VPN-based connectivity to environments
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-full" />
                      Docker Compose compatible services
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-full" />
                      Nix-based reproducible packages
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-full" />
                      SSH, VS Code, and IDE access
                    </li>
                  </ul>
                </div>
                <div className="p-8 lg:p-10">
                  <p className="text-foreground/60 text-xs font-semibold uppercase tracking-wider">Built For</p>
                  <ul className="mt-6 space-y-2 text-sm text-foreground/70">
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-full" />
                      Microservices and distributed systems
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-full" />
                      Teams needing fast onboarding
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-full" />
                      AI coding agents (Claude, Cursor)
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-full" />
                      Integration testing against real data
                    </li>
                  </ul>
                </div>
              </div>

              {/* CTA */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12">
                <Link
                  href="/docs"
                  className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/60 text-xs font-semibold uppercase tracking-wider">Learn</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">Documentation</h3>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/40 group-hover:text-foreground/60 group-hover:translate-x-1 transition-all" />
                  </div>
                </Link>
                <Link
                  href="/pricing"
                  className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-foreground/60 text-xs font-semibold uppercase tracking-wider">Start</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">View Pricing</h3>
                    </div>
                    <ArrowRight className="h-5 w-5 text-foreground/40 group-hover:text-foreground/60 group-hover:translate-x-1 transition-all" />
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
