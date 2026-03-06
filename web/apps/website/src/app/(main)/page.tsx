'use client'

import Link from 'next/link'
import { Button, ScrollArea, VSCodeIcon, JetBrainsIcon, AntigravityIcon, ZedIcon, CursorIcon } from '@kloudlite/ui'
import { GetStartedButton } from '@/components/get-started-button'
import { WebsiteHeader } from '@/components/website-header'
import { WebsiteFooter } from '@/components/website-footer'
import { PageHeroTitle } from '@/components/page-hero-title'
import { ArrowRight, Copy, Layers, ArrowLeftRight, Route, Package, Terminal, Camera, Sparkles, Code2 } from 'lucide-react'
import { GridContainer } from '@/components/home-page/grid-container'
import { TypewriterText } from '@/components/home-page/typewriter-text'
import { WorkflowVisualization } from '@/components/home-page/workflow-visualization'
import { FeatureCard, FeatureCardContainer } from '@/components/home-page/feature-card'
import { CrossMarker } from '@/components/home-page/cross-marker'
import { testimonials } from '@/components/home-page/testimonials'
import { cn } from '@kloudlite/lib'

function WebsiteLandingPage() {

  return (
    <div className="bg-background h-screen">
      <ScrollArea className="h-full">
        <WebsiteHeader currentPage="home" />
        <main>
        {/* Main Content in Grid Container */}
        <div className="px-6 pt-8 lg:px-8 lg:pt-12">
          <GridContainer className="px-6 lg:px-12">
            {/* Hero Section */}
            <div className="py-20 lg:py-28">
              <div className="text-center">
                <PageHeroTitle accentedWord="Environments.">
                  Cloud Development
                </PageHeroTitle>

                <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-base lg:text-lg leading-relaxed">
                  Reduce the development loop. No setup, no builds, no deployments.
                </p>

                <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
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

              {/* For Developer typewriter - Hidden on mobile */}
              <div className="hidden sm:block mt-20 text-center">
                <p className="text-muted-foreground text-lg sm:text-xl font-medium">
                  For <TypewriterText />
                </p>
              </div>

              {/* Workflow Visualization - Hidden on mobile */}
              <div className="hidden lg:block">
                <WorkflowVisualization />
              </div>
            </div>

            {/* Features Grid */}
            <div className="grid sm:grid-cols-2 lg:grid-cols-3 border-t border-foreground/10 -mx-6 lg:-mx-12">
              {/* Row 1 - Built for developers */}
              <div className="sm:col-span-2 lg:col-span-1 p-8 lg:p-12 border-l border-b border-foreground/10 lg:border-r flex flex-col justify-center bg-foreground/[0.015]">
                <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-3xl lg:text-4xl">
                  Built for developers
                </h2>
                <p className="text-muted-foreground mt-6 text-base lg:text-lg">
                  Focus on code, not ops.
                </p>
              </div>
              <FeatureCardContainer className="border-l border-b border-foreground/10 sm:border-r lg:border-r" href="/blog/environment-forking">
                <FeatureCard
                  icon={<Copy className="h-7 w-7" />}
                  title="Environment Forking"
                  description="Fork entire environments with a single command."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              <FeatureCardContainer className="border-l border-b border-foreground/10 sm:border-r" href="/blog/workspace-forking">
                <FeatureCard
                  icon={<Layers className="h-7 w-7" />}
                  title="Workspace Forking"
                  description="Fork workspaces instantly for parallel work."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              {/* Row 2 */}
              <FeatureCardContainer className="border-l sm:border-r lg:border-r border-b border-foreground/10" href="/blog/environment-switching">
                <FeatureCard
                  icon={<ArrowLeftRight className="h-7 w-7" />}
                  title="Environment Switching"
                  description="Switch contexts between environments without losing any state."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              <FeatureCardContainer className="border-l sm:border-r lg:border-r border-b border-foreground/10" href="/blog/service-intercepts">
                <FeatureCard
                  icon={<Route className="h-7 w-7" />}
                  title="Service Intercepts"
                  description="Route environment service traffic to your workspace."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              <FeatureCardContainer className="border-l sm:border-r border-b border-foreground/10" href="/blog/nix-package-management">
                <FeatureCard
                  icon={<Package className="h-7 w-7" />}
                  title="Package Management"
                  description="Nix-based reproducible package management."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              {/* Row 3 */}
              <FeatureCardContainer className="border-l sm:border-r lg:border-r border-b border-foreground/10" href="/blog/environment-snapshots">
                <FeatureCard
                  icon={<Camera className="h-7 w-7" />}
                  title="Environment Snapshots"
                  description="Capture and restore complete environment states instantly."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              <FeatureCardContainer className="border-l sm:border-r lg:border-r border-b border-foreground/10" href="/blog/workspace-snapshots">
                <FeatureCard
                  icon={<Camera className="h-7 w-7" />}
                  title="Workspace Snapshots"
                  description="Save and share workspace configurations effortlessly."
                />
                <span className="text-muted-foreground group-hover:text-primary text-sm font-medium inline-flex items-center gap-2 transition-all duration-500 opacity-100 md:opacity-0 md:group-hover:opacity-100 mt-6">
                  Learn more <ArrowRight className="h-4 w-4" />
                </span>
              </FeatureCardContainer>
              <div className="group relative p-8 lg:p-10 bg-background hover:bg-foreground/[0.02] transition-all duration-500 overflow-hidden border-l border-b border-foreground/10 sm:border-r flex items-center justify-center">
                {/* Subtle bottom accent line */}
                <div className="absolute bottom-0 left-0 w-0 h-[1px] bg-primary group-hover:w-full transition-all duration-700 ease-out" />

                <Link
                  href="/docs"
                  className="flex flex-col items-center gap-5 text-center"
                >
                  <div className="text-muted-foreground group-hover:text-primary transition-colors duration-500">
                    <ArrowRight className="h-8 w-8 opacity-60 group-hover:opacity-100 transition-all duration-500 group-hover:translate-x-2" />
                  </div>
                  <div className="space-y-2">
                    <span className="block text-lg font-semibold text-foreground group-hover:text-primary transition-colors duration-500">
                      Explore All Features
                    </span>
                    <span className="block text-sm text-muted-foreground">
                      Discover the complete platform
                    </span>
                  </div>
                </Link>
              </div>

              {/* Section Spacer */}
              <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
              </div>

              {/* Toolchain Section Header */}
              <div className="sm:col-span-2 lg:col-span-3 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                <h2 className="text-foreground text-4xl lg:text-4xl xl:text-5xl font-bold tracking-tight">
                  Your entire toolchain, <span className="relative inline-block">
                    <span className="relative z-10">connected.</span>
                    <span className="absolute bottom-0 left-0 right-0 h-1 bg-primary"></span>
                  </span>
                </h2>
                <p className="text-muted-foreground mt-4 text-base lg:text-lg max-w-2xl">
                  From AI agents to your favorite IDE, Kloudlite works where you work.
                </p>
              </div>

              {/* AI Ready Workspaces Card */}
              <div className="sm:col-span-2 lg:col-span-1 p-8 lg:p-12 border-b lg:border-r border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors group cursor-default relative min-h-[400px] lg:min-h-[500px] flex flex-col">
                {/* Background decorative icon */}
                <div className="absolute top-8 right-8 opacity-10">
                  <Sparkles className="h-32 w-32 text-muted-foreground" />
                </div>

                <div className="relative z-10">
                  <div className="text-primary mb-4">
                    <Sparkles className="h-8 w-8" />
                  </div>
                  <h3 className="text-foreground text-2xl font-bold mb-4">
                    AI Ready Workspaces
                  </h3>
                  <p className="text-muted-foreground text-base leading-relaxed mb-4">
                    Built-in support for the next generation of AI coding tools. Supports vibecoding sessions out of the box.
                  </p>
                  <Link href="/blog/ai-ready-workspaces" className="text-primary text-sm font-medium inline-flex items-center gap-1 hover:gap-2 transition-all opacity-100 md:opacity-0 md:group-hover:opacity-100">
                    Learn more <ArrowRight className="h-4 w-4" />
                  </Link>
                </div>

                {/* AI Tool Tags */}
                <div className="mt-auto relative z-10 flex flex-wrap gap-2">
                  <span className="px-3 py-1.5 text-xs font-semibold uppercase tracking-wider bg-foreground/5 border border-foreground/10 text-muted-foreground">CLAUDE CODE</span>
                  <span className="px-3 py-1.5 text-xs font-semibold uppercase tracking-wider bg-foreground/5 border border-foreground/10 text-muted-foreground">GEMINI CLI</span>
                  <span className="px-3 py-1.5 text-xs font-semibold uppercase tracking-wider bg-foreground/5 border border-foreground/10 text-muted-foreground">OPENCODE</span>
                  <span className="px-3 py-1.5 text-xs font-semibold uppercase tracking-wider bg-foreground/5 border border-foreground/10 text-muted-foreground">CODEX CLI</span>
                </div>
              </div>

              {/* Access from any IDE Card */}
              <div className="sm:col-span-2 lg:col-span-2 p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-colors group cursor-default relative min-h-[400px] lg:min-h-[500px] flex flex-col">
                {/* Background decorative icon */}
                <div className="absolute top-8 right-8 opacity-10">
                  <Code2 className="h-32 w-32 text-muted-foreground" />
                </div>

                <div className="relative z-10 mb-8">
                  <div className="text-primary mb-4">
                    <Code2 className="h-8 w-8" />
                  </div>
                  <h3 className="text-foreground text-2xl font-bold mb-4">
                    Access from any IDE
                  </h3>
                  <p className="text-muted-foreground text-base leading-relaxed mb-4">
                    Connect your local editor directly to cloud resources. Zero latency, full Intellisense.
                  </p>
                  <Link href="/blog/ide-integration" className="text-primary text-sm font-medium inline-flex items-center gap-1 hover:gap-2 transition-all opacity-100 md:opacity-0 md:group-hover:opacity-100">
                    Learn more <ArrowRight className="h-4 w-4" />
                  </Link>
                </div>

                {/* IDE Grid */}
                <div className="mt-auto relative z-10 grid grid-cols-2 lg:grid-cols-3 gap-3">
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <VSCodeIcon className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">VS Code</span>
                  </div>
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <CursorIcon className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">Cursor</span>
                  </div>
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <JetBrainsIcon className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">JetBrains</span>
                  </div>
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <AntigravityIcon className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">Antigravity</span>
                  </div>
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <ZedIcon className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">Zed</span>
                  </div>
                  <div className="flex items-center gap-3 px-4 py-3 rounded-sm border border-foreground/5 bg-foreground/[0.02] hover:border-foreground/10 hover:bg-foreground/[0.04] transition-all group/item">
                    <Terminal className="h-5 w-5 text-muted-foreground group-hover/item:text-foreground transition-colors" />
                    <span className="text-sm font-medium text-muted-foreground group-hover/item:text-foreground transition-colors">Web Terminal</span>
                  </div>
                </div>
              </div>

              {/* Section Spacer */}
              <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                <CrossMarker className="bottom-0 left-1/2 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
              </div>

              {/* Testimonials Section Header */}
              <div className="sm:col-span-2 lg:col-span-3 p-8 lg:p-16 border-b border-foreground/10 bg-foreground/[0.015]">
                <h2 className="text-foreground text-4xl lg:text-4xl xl:text-5xl font-bold tracking-tight">
                  Trusted by developers
                </h2>
                <p className="text-muted-foreground mt-4 text-base lg:text-base xl:text-lg max-w-2xl">
                  Teams around the world are building faster with Kloudlite.
                </p>
              </div>

              {/* Testimonials Grid - Rotating Content */}
              <p className="sr-only sm:col-span-2 lg:col-span-3" role="status" aria-live="polite" aria-atomic="true">
                Showing {Math.min(3, testimonials.length)} developer testimonials.
              </p>
              {testimonials.slice(0, 3).map((testimonial, index) => (
                <div
                  key={testimonial.name}
                  className={cn(
                    "p-8 lg:p-12 border-b border-foreground/10 bg-foreground/[0.015] hover:bg-foreground/[0.03] transition-all duration-300 group cursor-default",
                    "sm:col-span-2 lg:col-span-1",
                    index === 0 && "sm:border-r lg:border-r",
                    index === 1 && "lg:border-r"
                  )}
                >
                  <div className="mb-8 min-h-[140px] flex items-center">
                    <p className="text-foreground text-base leading-relaxed">
                      &quot;{testimonial.quote}&quot;
                    </p>
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <span className="text-primary font-bold text-sm">{testimonial.initials}</span>
                    </div>
                    <div>
                      <p className="text-foreground font-semibold text-sm">{testimonial.name}</p>
                      <p className="text-muted-foreground text-xs">{testimonial.title}, {testimonial.company}</p>
                    </div>
                  </div>
                </div>
              ))}

              {/* Section Spacer */}
              <div className="sm:col-span-2 lg:col-span-3 h-8 sm:h-16 border-b border-foreground/10 relative">
                <CrossMarker className="bottom-0 left-1/3 translate-y-1/2 -translate-x-1/2 w-5 h-5 hidden lg:block" />
              </div>

              {/* CTA Row - Start building faster */}
              <div className="p-8 lg:p-16 border-b border-foreground/10 sm:col-span-2 lg:col-span-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 bg-foreground/[0.015]">
                <div>
                  <h2 className="text-foreground text-3xl font-bold tracking-tight sm:text-3xl xl:text-4xl">
                    Start building faster
                  </h2>
                  <p className="text-muted-foreground mt-2 text-base">
                    Free with your own infrastructure.
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
                    <Link href="/pricing">
                      Pricing
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

export default function HomePage() {
  return <WebsiteLandingPage />
}
