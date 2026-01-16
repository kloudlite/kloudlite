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
              <div className="py-20 lg:py-32">
                <div className="text-center">
                  <h1 className="text-[2.5rem] font-semibold leading-[1.1] tracking-[-0.02em] sm:text-5xl md:text-6xl lg:text-[4rem]">
                    <span className="text-foreground">Why Kloudlite?</span>
                  </h1>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-lg text-lg leading-relaxed">
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
                  <p className="text-muted-foreground/50 text-4xl font-bold font-mono">01</p>
                  <h3 className="text-foreground mt-4 text-lg font-bold tracking-[-0.02em]">No Local Setup</h3>
                  <p className="text-muted-foreground mt-3 text-base leading-relaxed">
                    Cloud workspaces with your repo, packages, and tools. Ready in seconds.
                  </p>
                </div>
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <p className="text-muted-foreground/50 text-4xl font-bold font-mono">02</p>
                  <h3 className="text-foreground mt-4 text-lg font-bold tracking-[-0.02em]">Real Services</h3>
                  <p className="text-muted-foreground mt-3 text-base leading-relaxed">
                    Connect to your running databases, APIs, and queues. No mocks.
                  </p>
                </div>
                <div className="p-8 lg:p-10">
                  <p className="text-muted-foreground/50 text-4xl font-bold font-mono">03</p>
                  <h3 className="text-foreground mt-4 text-lg font-bold tracking-[-0.02em]">Traffic Intercept</h3>
                  <p className="text-muted-foreground mt-3 text-base leading-relaxed">
                    Route live requests to your workspace. Debug with production data.
                  </p>
                </div>
              </div>

              {/* The Shift */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <p className="text-muted-foreground text-sm font-semibold uppercase tracking-wider">Without Kloudlite</p>
                  <div className="mt-6 space-y-3 font-mono text-base">
                    <p className="text-foreground">git clone → install deps → setup db → configure env → mock services → code → push → wait for CI → deploy → test → find bug → repeat</p>
                  </div>
                  <p className="text-muted-foreground text-sm mt-6 font-mono">Hours to days per feature</p>
                </div>
                <div className="p-8 lg:p-10 bg-foreground/[0.02]">
                  <p className="text-primary text-sm font-semibold uppercase tracking-wider">With Kloudlite</p>
                  <div className="mt-6 space-y-3 font-mono text-base">
                    <p className="text-foreground">open workspace → connect env → code → done</p>
                  </div>
                  <p className="text-primary text-sm mt-6 font-mono">Minutes</p>
                </div>
              </div>

              {/* Technical Highlights */}
              <div className="grid lg:grid-cols-2 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-foreground/10">
                  <p className="text-muted-foreground text-sm font-semibold uppercase tracking-wider">Architecture</p>
                  <ul className="mt-6 space-y-3 text-base text-muted-foreground">
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-none" />
                      VPN-based connectivity to environments
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-none" />
                      Docker Compose compatible services
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-none" />
                      Nix-based reproducible packages
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-none" />
                      SSH, VS Code, and IDE access
                    </li>
                  </ul>
                </div>
                <div className="p-8 lg:p-10">
                  <p className="text-muted-foreground text-sm font-semibold uppercase tracking-wider">Built For</p>
                  <ul className="mt-6 space-y-3 text-base text-muted-foreground">
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-none" />
                      Microservices and distributed systems
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-none" />
                      Teams needing fast onboarding
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-none" />
                      AI coding agents (Claude, Cursor)
                    </li>
                    <li className="flex items-center gap-2">
                      <span className="w-1 h-1 bg-foreground/50 rounded-none" />
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
                      <p className="text-muted-foreground text-sm font-semibold uppercase tracking-wider">Learn</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">Documentation</h3>
                    </div>
                    <ArrowRight className="h-5 w-5 text-muted-foreground group-hover:text-foreground group-hover:translate-x-1 transition-all" />
                  </div>
                </Link>
                <Link
                  href="/pricing"
                  className="p-8 lg:p-10 group hover:bg-foreground/[0.02] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-muted-foreground text-sm font-semibold uppercase tracking-wider">Start</p>
                      <h3 className="text-foreground mt-2 text-lg font-bold tracking-[-0.02em]">View Pricing</h3>
                    </div>
                    <ArrowRight className="h-5 w-5 text-muted-foreground group-hover:text-foreground group-hover:translate-x-1 transition-all" />
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
