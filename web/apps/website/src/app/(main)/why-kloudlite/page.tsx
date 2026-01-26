'use client'

import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { GetStartedButton } from '@/components/get-started-button'
import { cn } from '@kloudlite/lib'
import { Zap, Clock, Target, Check, X } from 'lucide-react'
import Link from 'next/link'

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
      <div className="absolute inset-0 pointer-events-none overflow-visible">
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
        <div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />
        <div className="absolute inset-x-0 bottom-0 h-px bg-foreground/10" />

        {/* Animated pulses */}
        <div className="pulse-top absolute top-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
        <div className="pulse-right absolute right-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />
        <div className="pulse-bottom absolute bottom-0 w-12 h-px bg-gradient-to-r from-transparent via-primary to-transparent" />
        <div className="pulse-left absolute left-0 h-12 w-px bg-gradient-to-b from-transparent via-primary to-transparent" />

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
        <WebsiteHeader currentPage="why-kloudlite" alwaysShowBorder />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero */}
              <div className="py-20 lg:py-28">
                <div className="text-center">
                  <h1 className="text-[2.5rem] font-semibold leading-[1.1] tracking-[-0.02em] sm:text-5xl md:text-6xl lg:text-[4rem] text-foreground">
                    Development loops are <span className="relative inline-block">
                      <span className="relative z-10">broken.</span>
                      <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                    </span>
                  </h1>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-lg leading-relaxed">
                    Hours wasted on setup. Days lost to deployment cycles. Production bugs that could've been caught earlier. We're fixing that.
                  </p>
                </div>
              </div>

              {/* The Problem - Before/After Grid */}
              <div className="grid sm:grid-cols-2 -mx-6 lg:-mx-12 border-t border-foreground/10">
                {/* Traditional Workflow */}
                <div className="p-8 lg:p-12 border-b sm:border-r border-foreground/10 bg-foreground/[0.015]">
                  <div className="flex items-center gap-3 mb-6">
                    <div className="p-2 rounded-sm bg-destructive/10">
                      <X className="h-5 w-5 text-destructive" />
                    </div>
                    <h2 className="text-foreground text-2xl font-bold">Traditional Workflow</h2>
                  </div>

                  <div className="space-y-4">
                    <div className="flex items-start gap-3">
                      <Clock className="h-5 w-5 text-muted-foreground mt-0.5 flex-shrink-0" />
                      <div>
                        <p className="text-foreground font-medium">Hours of local setup</p>
                        <p className="text-muted-foreground text-sm mt-1">Install dependencies, configure databases, mock services</p>
                      </div>
                    </div>

                    <div className="flex items-start gap-3">
                      <Clock className="h-5 w-5 text-muted-foreground mt-0.5 flex-shrink-0" />
                      <div>
                        <p className="text-foreground font-medium">Wait for CI/CD pipelines</p>
                        <p className="text-muted-foreground text-sm mt-1">Push code, wait for builds, deploy to staging</p>
                      </div>
                    </div>

                    <div className="flex items-start gap-3">
                      <Clock className="h-5 w-5 text-muted-foreground mt-0.5 flex-shrink-0" />
                      <div>
                        <p className="text-foreground font-medium">Debug with mocked data</p>
                        <p className="text-muted-foreground text-sm mt-1">Find issues in production that mocks didn't catch</p>
                      </div>
                    </div>
                  </div>

                  <div className="mt-8 p-4 border border-foreground/10 bg-foreground/[0.02] rounded-sm">
                    <p className="text-muted-foreground text-sm font-mono">
                      <span className="text-destructive font-bold">Days</span> from idea to production
                    </p>
                  </div>
                </div>

                {/* With Kloudlite */}
                <div className="p-8 lg:p-12 border-b border-foreground/10 bg-primary/[0.02]">
                  <div className="flex items-center gap-3 mb-6">
                    <div className="p-2 rounded-sm bg-primary/10">
                      <Check className="h-5 w-5 text-primary" />
                    </div>
                    <h2 className="text-foreground text-2xl font-bold">With Kloudlite</h2>
                  </div>

                  <div className="space-y-4">
                    <div className="flex items-start gap-3">
                      <Zap className="h-5 w-5 text-primary mt-0.5 flex-shrink-0" />
                      <div>
                        <p className="text-foreground font-medium">Instant workspace</p>
                        <p className="text-muted-foreground text-sm mt-1">Pre-configured environment, ready in seconds</p>
                      </div>
                    </div>

                    <div className="flex items-start gap-3">
                      <Zap className="h-5 w-5 text-primary mt-0.5 flex-shrink-0" />
                      <div>
                        <p className="text-foreground font-medium">Code against live services</p>
                        <p className="text-muted-foreground text-sm mt-1">Connect directly to real databases and APIs</p>
                      </div>
                    </div>

                    <div className="flex items-start gap-3">
                      <Zap className="h-5 w-5 text-primary mt-0.5 flex-shrink-0" />
                      <div>
                        <p className="text-foreground font-medium">Test with production data</p>
                        <p className="text-muted-foreground text-sm mt-1">Intercept live traffic, debug real issues</p>
                      </div>
                    </div>
                  </div>

                  <div className="mt-8 p-4 border border-primary/20 bg-primary/[0.05] rounded-sm">
                    <p className="text-muted-foreground text-sm font-mono">
                      <span className="text-primary font-bold">Minutes</span> from idea to testing
                    </p>
                  </div>
                </div>
              </div>

              {/* Key Benefits */}
              <div className="grid sm:grid-cols-3 -mx-6 lg:-mx-12 border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-b sm:border-b-0 sm:border-r border-foreground/10 hover:bg-foreground/[0.02] transition-colors group cursor-default">
                  <div className="p-3 rounded-sm bg-primary/10 w-fit mb-4">
                    <Target className="h-6 w-6 text-primary" />
                  </div>
                  <h3 className="text-foreground text-lg font-bold mb-2">Zero Setup</h3>
                  <p className="text-muted-foreground text-base leading-relaxed">
                    No more "works on my machine". Every developer gets the same, production-like environment.
                  </p>
                </div>

                <div className="p-8 lg:p-10 border-b sm:border-b-0 sm:border-r border-foreground/10 hover:bg-foreground/[0.02] transition-colors group cursor-default">
                  <div className="p-3 rounded-sm bg-primary/10 w-fit mb-4">
                    <Zap className="h-6 w-6 text-primary" />
                  </div>
                  <h3 className="text-foreground text-lg font-bold mb-2">Real Services</h3>
                  <p className="text-muted-foreground text-base leading-relaxed">
                    Connect to actual databases, queues, and APIs. No mocks, no compromises.
                  </p>
                </div>

                <div className="p-8 lg:p-10 hover:bg-foreground/[0.02] transition-colors group cursor-default">
                  <div className="p-3 rounded-sm bg-primary/10 w-fit mb-4">
                    <Clock className="h-6 w-6 text-primary" />
                  </div>
                  <h3 className="text-foreground text-lg font-bold mb-2">Faster Feedback</h3>
                  <p className="text-muted-foreground text-base leading-relaxed">
                    Test against live data. Find bugs before they reach production. Ship with confidence.
                  </p>
                </div>
              </div>

              {/* CTA */}
              <div className="p-8 lg:p-16 -mx-6 lg:-mx-12 border-b border-foreground/10 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 bg-foreground/[0.015]">
                <div>
                  <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                    Ready to speed up?
                  </h2>
                  <p className="text-muted-foreground mt-2 text-base">
                    Start building faster with Kloudlite.
                  </p>
                </div>
                <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                  <GetStartedButton size="lg" className="w-full sm:w-auto" />
                  <Button
                    asChild
                    variant="outline"
                    size="lg"
                    className="w-full sm:w-auto"
                  >
                    <Link href="/docs">
                      Documentation
                    </Link>
                  </Button>
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
