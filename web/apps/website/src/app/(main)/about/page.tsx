'use client'

import { ScrollArea, Button } from '@kloudlite/ui'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { cn } from '@kloudlite/lib'
import {
  Github,
  Linkedin,
  Twitter,
  Calendar,
  Zap,
  Terminal,
  Settings,
  GitBranch,
  Boxes,
  ArrowRight
} from 'lucide-react'
import Link from 'next/link'
import { PageHeroTitle } from '@/components/page-hero-title'
import { GetStartedButton } from '@/components/get-started-button'

// Cross marker component with pulse animation
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
        {/* Static border lines */}
        <div className="absolute inset-y-0 left-0 w-px bg-foreground/10" />
        <div className="absolute inset-y-0 right-0 w-px bg-foreground/10" />
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
      <div className="relative">{children}</div>
    </div>
  )
}

export default function AboutPage() {
  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="about" />
        <main>
          <div className="px-6 pt-8 lg:px-8 lg:pt-12">
            <GridContainer className="px-6 lg:px-12">
              {/* Hero Section */}
              <div className="py-20 lg:py-28">
                <div className="text-center">
                  <PageHeroTitle accentedWord="Developers.">
                    Built by
                  </PageHeroTitle>
                  <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-lg lg:text-xl leading-relaxed">
                    We&apos;re engineers who got tired of waiting. Waiting for Docker to build.
                    Waiting for environments to sync. Waiting to see if code actually works.
                  </p>
                  <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
                    <GetStartedButton size="lg" className="w-full sm:w-auto rounded-none" />
                    <Button variant="outline" size="lg" className="w-full sm:w-auto rounded-none" asChild>
                      <Link href="/docs">Documentation</Link>
                    </Button>
                  </div>
                </div>
              </div>

              {/* Enhanced Stats Section */}
              <div className="grid grid-cols-2 lg:grid-cols-4 -mx-6 lg:-mx-12 border-t border-b border-foreground/10">
                <div className="p-8 lg:p-10 border-r border-foreground/10 text-center group hover:bg-foreground/[0.03] transition-colors">
                  <div className="text-primary mb-3">
                    <Calendar className="h-6 w-6 mx-auto" />
                  </div>
                  <p className="text-foreground text-3xl lg:text-4xl font-mono font-semibold tracking-tight">2023</p>
                  <p className="text-muted-foreground mt-2 text-sm group-hover:text-foreground transition-colors">Founded</p>
                </div>

                <div className="p-8 lg:p-10 lg:border-r border-foreground/10 text-center group hover:bg-foreground/[0.03] transition-colors">
                  <div className="text-primary mb-3">
                    <Github className="h-6 w-6 mx-auto" />
                  </div>
                  <p className="text-foreground text-3xl lg:text-4xl font-mono font-semibold tracking-tight">100%</p>
                  <p className="text-muted-foreground mt-2 text-sm group-hover:text-foreground transition-colors">Open Source</p>
                </div>

                <div className="p-8 lg:p-10 border-r border-t lg:border-t-0 border-foreground/10 text-center group hover:bg-foreground/[0.03] transition-colors">
                  <div className="text-primary mb-3">
                    <Zap className="h-6 w-6 mx-auto" />
                  </div>
                  <p className="text-foreground text-3xl lg:text-4xl font-mono font-semibold tracking-tight">10x</p>
                  <p className="text-muted-foreground mt-2 text-sm group-hover:text-foreground transition-colors">Faster Feedback</p>
                </div>

                <div className="p-8 lg:p-10 border-t lg:border-t-0 border-foreground/10 text-center group hover:bg-foreground/[0.03] transition-colors">
                  <div className="text-primary mb-3">
                    <Terminal className="h-6 w-6 mx-auto" />
                  </div>
                  <p className="text-foreground text-3xl lg:text-4xl font-mono font-semibold tracking-tight">0</p>
                  <p className="text-muted-foreground mt-2 text-sm group-hover:text-foreground transition-colors">Setup Required</p>
                </div>
              </div>

              {/* Main Content Grid */}
              <div className="grid sm:grid-cols-2 border-t border-foreground/10 -mx-6 lg:-mx-12">

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Mission Header */}
                <div className="sm:col-span-2 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                  <h2 className="text-foreground text-4xl lg:text-5xl font-bold tracking-tight">
                    Our mission is to eliminate <span className="relative inline-block">
                      <span className="relative z-10">waiting.</span>
                      <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                    </span>
                  </h2>
                  <p className="text-muted-foreground mt-4 text-base lg:text-lg max-w-3xl">
                    Every second a developer spends waiting for builds, deployments, or environment setup is a second lost from actual problem-solving.
                  </p>
                </div>

                {/* Mission Content - Problem */}
                <div className="group p-8 lg:p-12 border-b border-foreground/10 sm:border-r bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors">
                  <p className="text-primary text-xs font-semibold uppercase tracking-wider mb-3">The Problem</p>
                  <h3 className="text-foreground text-xl font-bold mb-3">Distributed apps, localhost development</h3>
                  <p className="text-muted-foreground text-base leading-relaxed mb-4">
                    Modern applications are distributed across microservices, databases, queues, and third-party APIs. But developers still code on localhost, disconnected from reality.
                  </p>
                  <Link href="/blog/distributed-apps-localhost-problem" className="text-primary text-sm font-medium inline-flex items-center gap-1 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Read more <ArrowRight className="h-3 w-3" />
                  </Link>
                </div>

                {/* Mission Content - Gap */}
                <div className="group p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors">
                  <p className="text-primary text-xs font-semibold uppercase tracking-wider mb-3">The Gap</p>
                  <h3 className="text-foreground text-xl font-bold mb-3">Mocks don&apos;t match production</h3>
                  <p className="text-muted-foreground text-base leading-relaxed mb-4">
                    Docker Compose is slow. Mocked services behave differently than real ones. By the time you find bugs in staging, you&apos;ve wasted hours.
                  </p>
                  <Link href="/blog/mocks-dont-match-production" className="text-primary text-sm font-medium inline-flex items-center gap-1 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Read more <ArrowRight className="h-3 w-3" />
                  </Link>
                </div>

                {/* Mission Content - Solution (spans full width) */}
                <div className="group sm:col-span-2 p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors">
                  <p className="text-primary text-xs font-semibold uppercase tracking-wider mb-3">Our Solution</p>
                  <h3 className="text-foreground text-xl font-bold mb-3">Cloud dev environments connected to real services</h3>
                  <p className="text-muted-foreground text-base leading-relaxed max-w-3xl mb-4">
                    Kloudlite gives you cloud-hosted workspaces that connect directly to your staging, QA, or even production environments. Write code, intercept service traffic, and see real results instantly—no mocks, no waiting.
                  </p>
                  <Link href="/blog/cloud-dev-environments-solution" className="text-primary text-sm font-medium inline-flex items-center gap-1 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Read more <ArrowRight className="h-3 w-3" />
                  </Link>
                </div>

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Principles Header */}
                <div className="sm:col-span-2 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                  <h2 className="text-foreground text-4xl lg:text-5xl font-bold tracking-tight">
                    Core <span className="relative inline-block">
                      <span className="relative z-10">principles.</span>
                      <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                    </span>
                  </h2>
                </div>

                {/* Principle 1 */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:border-r">
                  <FeatureCard
                    icon={<Zap className="h-6 w-6" />}
                    title="Speed Above All"
                    description="Every millisecond matters. From workspace startup (<30s) to service intercepts (instant), we obsess over reducing latency at every step."
                  />
                  <Link href="/blog/speed-above-all" className="text-primary text-sm font-medium inline-flex items-center gap-1 mt-4 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Read more <ArrowRight className="h-3 w-3" />
                  </Link>
                </FeatureCardContainer>

                {/* Principle 2 */}
                <FeatureCardContainer className="border-b border-foreground/10">
                  <FeatureCard
                    icon={<Settings className="h-6 w-6" />}
                    title="Zero Configuration"
                    description="Developers shouldn't need a degree in DevOps to write code. Our tools work out of the box—no YAML hell, no infrastructure expertise required."
                  />
                  <Link href="/blog/zero-configuration" className="text-primary text-sm font-medium inline-flex items-center gap-1 mt-4 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Read more <ArrowRight className="h-3 w-3" />
                  </Link>
                </FeatureCardContainer>

                {/* Cross Marker between rows */}
                <div className="sm:col-span-2 h-0 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden sm:block" />
                </div>

                {/* Principle 3 */}
                <FeatureCardContainer className="border-b border-foreground/10 sm:col-span-2">
                  <FeatureCard
                    icon={<GitBranch className="h-6 w-6" />}
                    title="Open by Default"
                    description="Our core platform is open source and always will be. We build in public, welcome contributions, and believe transparency creates better software."
                  />
                  <Link href="/blog/open-by-default" className="text-primary text-sm font-medium inline-flex items-center gap-1 mt-4 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                    Read more <ArrowRight className="h-3 w-3" />
                  </Link>
                </FeatureCardContainer>

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-2/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Architecture Section */}
                <div className="group sm:col-span-2 p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors cursor-default min-h-[400px] flex flex-col lg:flex-row gap-8 items-center">
                  <div className="flex-1">
                    <div className="text-primary mb-4">
                      <Boxes className="h-8 w-8" />
                    </div>
                    <h3 className="text-foreground text-3xl lg:text-4xl font-bold mb-4">
                      Kubernetes-native architecture.
                    </h3>
                    <p className="text-muted-foreground text-base lg:text-lg leading-relaxed max-w-xl mb-4">
                      Built on Kubernetes with custom CRDs and controllers. Your workspaces and environments are declared as resources and reconciled automatically.
                    </p>
                    <p className="text-muted-foreground text-sm leading-relaxed max-w-xl mb-4">
                      We don&apos;t hide the infrastructure—we embrace it. Everything from workspace pods to service intercepts is managed through Kubernetes primitives you already know.
                    </p>
                    <Link href="/blog/kubernetes-native-architecture" className="text-primary text-sm font-medium inline-flex items-center gap-1 hover:gap-2 transition-all opacity-0 group-hover:opacity-100">
                      Learn about our architecture <ArrowRight className="h-3 w-3" />
                    </Link>
                  </div>

                  <div className="flex-1 w-full">
                    <div className="bg-foreground/[0.03] border border-foreground/10 p-6 rounded-sm space-y-4">
                      <div className="flex items-start gap-3">
                        <div className="w-2 h-2 bg-primary mt-2 flex-shrink-0" />
                        <div>
                          <p className="text-foreground font-semibold text-sm">Custom Resource Definitions</p>
                          <p className="text-muted-foreground text-xs mt-1">Workspace, Environment, ServiceIntercept CRDs</p>
                        </div>
                      </div>
                      <div className="flex items-start gap-3">
                        <div className="w-2 h-2 bg-primary mt-2 flex-shrink-0" />
                        <div>
                          <p className="text-foreground font-semibold text-sm">Kubernetes Controllers</p>
                          <p className="text-muted-foreground text-xs mt-1">Reconcile desired state with actual infrastructure</p>
                        </div>
                      </div>
                      <div className="flex items-start gap-3">
                        <div className="w-2 h-2 bg-primary mt-2 flex-shrink-0" />
                        <div>
                          <p className="text-foreground font-semibold text-sm">Nix Package Manager</p>
                          <p className="text-muted-foreground text-xs mt-1">Reproducible package management at the node level</p>
                        </div>
                      </div>
                      <div className="flex items-start gap-3">
                        <div className="w-2 h-2 bg-primary mt-2 flex-shrink-0" />
                        <div>
                          <p className="text-foreground font-semibold text-sm">Service Mesh Integration</p>
                          <p className="text-muted-foreground text-xs mt-1">SOCAT-based traffic forwarding for intercepts</p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* Open Source Hero Section */}
                <div className="sm:col-span-2 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                  <div className="max-w-3xl mx-auto text-center">
                    <div className="text-primary mb-4 flex justify-center">
                      <Github className="h-10 w-10" />
                    </div>
                    <h2 className="text-foreground text-4xl lg:text-5xl font-bold tracking-tight mb-4">
                      100% Open Source.
                    </h2>
                    <p className="text-muted-foreground text-base lg:text-lg leading-relaxed mb-6">
                      Our entire platform—controllers, CRDs, CLI, workspace images—is open source under MIT license. Star us, fork us, contribute, or run it yourself.
                    </p>
                    <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mt-8">
                      <Button variant="default" size="lg" className="w-full sm:w-auto rounded-none" asChild>
                        <Link href="https://github.com/kloudlite/kloudlite" target="_blank" rel="noopener noreferrer">
                          <Github className="h-4 w-4 mr-2" />
                          View on GitHub
                        </Link>
                      </Button>
                      <Button variant="outline" size="lg" className="w-full sm:w-auto rounded-none" asChild>
                        <Link href="/docs/architecture">Architecture Docs</Link>
                      </Button>
                    </div>
                  </div>
                </div>

                {/* Section Spacer */}
                <div className="sm:col-span-2 h-8 sm:h-16 border-b border-foreground/10 relative">
                  <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
                </div>

                {/* CTA Section */}
                <div className="p-8 lg:p-16 border-b border-foreground/10 sm:col-span-2 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 bg-foreground/[0.015]">
                  <div>
                    <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
                      Ready to build faster?
                    </h2>
                    <p className="text-muted-foreground mt-2 text-base lg:text-lg">
                      Create your first workspace and connect to your services in minutes.
                    </p>
                  </div>
                  <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                    <GetStartedButton size="lg" className="w-full sm:w-auto rounded-none" />
                    <Button variant="outline" size="lg" className="w-full sm:w-auto rounded-none" asChild>
                      <Link href="/docs">Documentation</Link>
                    </Button>
                  </div>
                </div>
              </div>

              {/* Social Links */}
              <div className="p-8 lg:p-10 -mx-6 lg:-mx-12 flex items-center justify-between">
                <p className="text-muted-foreground text-sm">Follow our journey</p>
                <div className="flex gap-2">
                  <a
                    href="https://github.com/kloudlite"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-2.5 border border-foreground/10 text-muted-foreground hover:text-foreground hover:border-foreground/20 transition-colors"
                  >
                    <Github className="h-4 w-4" />
                  </a>
                  <a
                    href="https://twitter.com/kloudlite"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-2.5 border border-foreground/10 text-muted-foreground hover:text-foreground hover:border-foreground/20 transition-colors"
                  >
                    <Twitter className="h-4 w-4" />
                  </a>
                  <a
                    href="https://linkedin.com/company/kloudlite"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-2.5 border border-foreground/10 text-muted-foreground hover:text-foreground hover:border-foreground/20 transition-colors"
                  >
                    <Linkedin className="h-4 w-4" />
                  </a>
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
